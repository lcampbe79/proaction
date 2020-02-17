package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/internal/event"
	"github.com/proactionhq/proaction/pkg/githubapi"
	"github.com/proactionhq/proaction/pkg/scanner"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	githubPathRegex = regexp.MustCompile("/([^/?=]+)/([^/?=]+)/blob/([^/?=]+)/(.*)")
)

func ScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "scan",
		Short:         "r",
		Long:          ``,
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			if err := event.Init(v); err != nil {
				if v.GetBool("debug") {
					fmt.Printf("%s\n", err.Error())
				}
			}

			localFile := args[0]

			// Let's be kind. If someone put in a github.com url, we can probably download
			// the file, flip a few flags around and print the recommendations to stdout
			parsedURL, err := url.ParseRequestURI(localFile)
			if err == nil {
				// TODO we should support domains that aren't github.com
				if parsedURL.Hostname() == "github.com" {
					downloadedFile, err := downloadFileFromGitHub(parsedURL.Path)
					if err != nil {
						return errors.Wrap(err, "tried unsuccesfully to download file from github")
					}
					defer os.RemoveAll(downloadedFile)

					localFile = downloadedFile

					v.Set("show-diff", true)
				}
			}

			content, err := ioutil.ReadFile(localFile)
			if err != nil {
				return errors.Wrap(err, "failed to read workflow")
			}

			s := scanner.NewScanner()
			s.OriginalContent = string(content)

			if len(v.GetStringSlice("check")) == 0 {
				s.EnableAllChecks()
			} else {
				for _, check := range v.GetStringSlice("check") {
					s.EnabledChecks = append(s.EnabledChecks, check)
				}
			}

			err = s.ScanWorkflow()
			if err != nil {
				return errors.Wrap(err, "failed to scan workflow")
			}

			if len(s.Issues) == 0 {
				fmt.Println("No recommendations found!")
				os.Exit(0)
			}

			if !v.GetBool("quiet") {
				fmt.Printf("%#v", s.GetOutput())
			}

			if s.OriginalContent != s.RemediatedContent {
				if v.GetBool("show-diff") {
					dmp := diffmatchpatch.New()
					diffs := dmp.DiffMain(s.OriginalContent, s.RemediatedContent, false)
					fmt.Println(dmp.DiffPrettyText(dmp.DiffCleanupEfficiency(diffs)))
				} else if v.GetString("out") == "" {
					err := ioutil.WriteFile(localFile, []byte(s.RemediatedContent), 0755)
					if err != nil {
						return errors.Wrap(err, "failed to update workflow with remediations")
					}
				} else {
					d, _ := filepath.Split(v.GetString("out"))
					if err := os.MkdirAll(d, 0755); err != nil {
						return errors.Wrap(err, "failed to mkdir for out file")
					}

					err := ioutil.WriteFile(v.GetString("out"), []byte(s.RemediatedContent), 0755)
					if err != nil {
						return errors.Wrap(err, "failed to update workflow with remediations")
					}
				}

			}

			os.Exit(1)
			return nil
		},
	}

	cmd.Flags().StringSlice("check", []string{}, "check(s) to run. if empty, all checks will run")
	cmd.Flags().String("out", "", "when set, the updated workflow will be written to the file specified, instead of in place")
	cmd.Flags().Bool("dry-run", false, "when set, proaction will print the output and recommended changes, but will not make changes to the file")
	cmd.Flags().Bool("quiet", false, "when set, proaction will not print explanations but will only update the workflow files with recommendations")
	cmd.Flags().Bool("debug", false, "when set, echo debug statements")
	cmd.Flags().Bool("show-diff", false, "when set, instead of writing the file, just show a diff")

	return cmd
}

func downloadFileFromGitHub(path string) (string, error) {
	matches := githubPathRegex.FindStringSubmatch(path)

	if len(matches) != 5 {
		return "", fmt.Errorf("Expected 5 matches in regex, but found %d", len(matches))
	}

	owner := matches[1]
	repo := matches[2]
	branch := matches[3]
	filename := matches[4]

	githubClient := githubapi.NewGitHubClient()
	fileContents, _, _, err := githubClient.Repositories.GetContents(
		context.Background(), owner, repo, filename,
		&github.RepositoryContentGetOptions{
			Ref: fmt.Sprintf("heads/%s", branch),
		})
	if err != nil {
		return "", errors.Wrap(err, "failed to download contents from github")
	}

	tmpFile, err := ioutil.TempFile("", "proaction")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}

	content, err := fileContents.GetContent()
	if err != nil {
		return "", errors.Wrap(err, "failed to get contents")
	}

	if err := ioutil.WriteFile(tmpFile.Name(), []byte(content), 0755); err != nil {
		return "", errors.Wrap(err, "failed to save to temp file")
	}

	return tmpFile.Name(), nil
}
