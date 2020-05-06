package collect

import (
	"context"

	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/pkg/collect/types"
	"github.com/proactionhq/proaction/pkg/githubapi"
	"github.com/proactionhq/proaction/pkg/logger"
	"github.com/proactionhq/proaction/pkg/ref"
	"go.uber.org/zap"
)

func parseGitHubRef(input string, collectors []string) (*types.Output, error) {
	logger.Debug("parseGitHubRef",
		zap.String("input", input),
		zap.Strings("collectors", collectors))

	output := types.Output{}

	owner, repo, _, _, err := ref.RefToParts(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ref")
	}

	for _, collector := range collectors {
		if collector == "repoInfo" {
			err := retrieveRepoInfo(owner, repo, &output)
			if err != nil {

			}
		} else if collector == "branches" {

		} else if collector == "commits" {

		} else if collector == "tags" {

		} else if collector == "recommendations" {

		} else {
			return nil, errors.Errorf("unknown collector %q", collector)
		}
	}

	return &output, nil
}

func retrieveRepoInfo(owner string, repo string, output *types.Output) error {
	githubClient := githubapi.NewGitHubClient()

	githubRepo, _, err := githubClient.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return errors.Wrap(err, "failed to get github repo")
	}

	output.Owner = githubRepo.GetOwner().GetLogin()
	output.Repo = githubRepo.GetName()
	output.IsArchived = githubRepo.GetArchived()
	output.IsPublic = true // TODO the version of the client doesn't have this field
	output.DefaultBranch = githubRepo.GetDefaultBranch()
	output.Forks = []string{} // TODO
	output.IsFork = githubRepo.GetFork()
	output.Head = "" // TODO

	return nil
}
