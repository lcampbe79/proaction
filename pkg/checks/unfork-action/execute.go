package unforkaction

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/pkg/githubapi"
	"github.com/proactionhq/proaction/pkg/issue"
	"github.com/proactionhq/proaction/pkg/ref"
	"github.com/proactionhq/proaction/pkg/workflow"
)

func executeUnforkActionCheckForWorkflow(parsedWorkflow *workflow.ParsedWorkflow) ([]*issue.Issue, error) {
	issues := []*issue.Issue{}

 s	for jobName, job := range parsedWorkflow.Jobs {
		for _, step := range job.Steps {
			if step.Uses == "" {
				continue
			}

			if strings.HasPrefix(step.Uses, "docker://") {
				continue
			}

			isFork, upstreamOwner, upstreamRepo, err := isGitHubRefFork(step.Uses)
			if err != nil {
				return nil, errors.Wrap(err, "failed to check is github ref fork")
			}

			if !isFork {
				continue
			}

			message := mustGetIssueMessage(parsedWorkflow.Name, jobName, step)

			unforkedRef := fmt.Sprintf("%s/%s", upstreamOwner, upstreamRepo)

			i := issue.Issue{
				CheckType: CheckName,
				CheckData: map[string]interface{}{
					"jobName":             jobName,
					"originalGitHubRef":   step.Uses,
					"remediatedGitHubRef": unforkedRef,
				},
				Message:      message,
				CanRemediate: true,
			}

			issues = append(issues, &i)
		}
	}

	return issues, nil
}

func mustGetIssueMessage(workflowName string, jobName string, step workflow.ParsedWorklowStep) string {
	return ""
}

func isGitHubRefFork(githubRef string) (bool, string, string, error) {
	owner, repo, _, _, err := ref.RefToParts(githubRef)
	if err != nil {
		return false, "", "", errors.Wrap(err, "failed to parse ref")
	}

	githubClient := githubapi.NewGitHubClient()
	getRepoResponse, _, err := githubClient.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return false, "", "", errors.Wrap(err, "failed to get repo")
	}

	if !getRepoResponse.GetFork() {
		return false, "", "", nil
	}

	return true, getRepoResponse.GetParent().GetOwner().GetLogin(), getRepoResponse.GetParent().GetName(), nil
}
