package scanner

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	unstablegithubref "github.com/proactionhq/proaction/pkg/checks/unstable-github-ref"
	"github.com/proactionhq/proaction/pkg/issue"
	workflowtypes "github.com/proactionhq/proaction/pkg/workflow/types"
	"gopkg.in/yaml.v3"
)

type Scanner struct {
	OriginalContent   string
	RemediatedContent string
	Issues            []*issue.Issue
	EnabledChecks     []string
}

func NewScanner() *Scanner {
	return &Scanner{
		Issues:        []*issue.Issue{},
		EnabledChecks: []string{},
	}
}

func (s *Scanner) EnableAllChecks() {
	s.EnabledChecks = []string{
		"unfork-action",
		"unstable-docker-tag",
		"unstable-github-ref",
		"outdated-action",
	}
}

func (s *Scanner) ScanWorkflow() error {
	sort.Sort(byPriority(s.EnabledChecks))

	for _, check := range s.EnabledChecks {
		// unmarshal from content each time so that each step can build on the last
		// this is important because if an issue changes the line count in the workflow
		// doing this will allow all remediation to still target the correct lines

		parsedWorkflow := workflowtypes.GitHubWorkflow{}
		if err := yaml.Unmarshal([]byte(s.getContent()), &parsedWorkflow); err != nil {
			return errors.Wrap(err, "failed to parse workflow")
		}

		if check == "unstable-github-ref" {
			issues, err := unstablegithubref.DetectIssues(parsedWorkflow)
			if err != nil {
				return errors.Wrap(err, "failed to run unstable unstable-github ref check")
			}

			s.Issues = append(s.Issues, issues...)

			for _, i := range issues {
				updated, err := unstablegithubref.RemediateIssue(s.getContent(), i)
				if err != nil {
					return errors.Wrap(err, "failed to apply remediation")
				}
				s.RemediatedContent = updated
			}
		}
	}
	// 	else if check == "unstable-docker-tag" {
	// 		issues, err := unstabledockertag.Run(s.getContent(), &parsedWorkflow)
	// 		if err != nil {
	// 			return errors.Wrap(err, "failed to run unstable unstable-docker-tag check")
	// 		}

	// 		s.Issues = append(s.Issues, issues...)
	// 	} else if check == "outdated-action" {
	// 		issues, err := outdatedaction.Run(s.getContent(), &parsedWorkflow)
	// 		if err != nil {
	// 			return errors.Wrap(err, "failed to run unstable outdated-action check")
	// 		}

	// 		s.Issues = append(s.Issues, issues...)
	// 	} else if check == "unfork-action" {
	// 		issues, err := unforkaction.Run(s.getContent(), &parsedWorkflow)
	// 		if err != nil {
	// 			return errors.Wrap(err, "failed to run unstable unfork-action check")
	// 		}

	// 		s.Issues = append(s.Issues, issues...)
	// 	}
	// }

	return nil
}

func applyRemediation(content string, i issue.Issue) (string, error) {
	return content, nil
}

func (s Scanner) getContent() string {
	if s.RemediatedContent != "" {
		return s.RemediatedContent
	}

	return s.OriginalContent
}

func (s Scanner) GetOutput() string {
	output := ""
	for _, i := range s.Issues {
		output = fmt.Sprintf("%s* %s\n", output, i.Message)
	}

	return output
}
