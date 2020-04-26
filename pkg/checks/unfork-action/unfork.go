package unforkaction

import (
	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/pkg/issue"
	workflowtypes "github.com/proactionhq/proaction/pkg/workflow/types"
)

var (
	CheckName = "unfork-action"
)

func Run(originalWorkflowContent string, parsedWorkflow *workflowtypes.GitHubWorkflow) ([]*issue.Issue, error) {
	issues, err := executeUnforkActionCheckForWorkflow(parsedWorkflow)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute unfork action check")
	}

	for _, i := range issues {
		err := remediateIssue(parsedWorkflow, i)
		if err != nil {
			return nil, errors.Wrap(err, "failed to remediate workflow")
		}
	}

	return issues, nil
}
