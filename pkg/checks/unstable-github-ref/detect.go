package unstablegithubref

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/pkg/issue"
	progresstypes "github.com/proactionhq/proaction/pkg/progress/types"
	workflowtypes "github.com/proactionhq/proaction/pkg/workflow/types"
)

type UnstableReason int

const (
	IsStable           UnstableReason = iota
	UnknownReason      UnstableReason = iota
	UnsupportedRef     UnstableReason = iota
	NoSpecifiedVersion UnstableReason = iota
	IsMaster           UnstableReason = iota
	IsBranch           UnstableReason = iota
	HasUnstableHistory UnstableReason = iota
	TagNotFound        UnstableReason = iota
	NotRecommendedTag  UnstableReason = iota
)

var (
	CheckName = "unstable-github-ref"
)

// DetectIssues will analyze the parsedWorkflow and return a list of issues
func DetectIssues(parsedWorkflow workflowtypes.GitHubWorkflow, setProgressFunc progresstypes.SetProgressFunc) ([]*issue.Issue, error) {
	issues := []*issue.Issue{}

	for jobName, job := range parsedWorkflow.Jobs {
		setProgressFunc(jobName, true, false)
		for stepIdx, step := range job.Steps {
			if step.Uses.Value == "" {
				continue
			}

			// ignore docker uses
			if strings.HasPrefix(step.Uses.Value, "docker://") {
				continue
			}

			isStable, unstableReason, stableRef, err := isGitHubRefStable(step.Uses.Value)
			if err != nil {
				return nil, errors.Wrap(err, "failed to check is github ref stable")
			}
			if isStable {
				continue
			}

			message := mustGetIssueMessage(parsedWorkflow.Name, jobName, unstableReason, step)

			i := issue.Issue{
				CheckType:  CheckName,
				JobName:    jobName,
				StepIdx:    stepIdx,
				LineNumber: step.Uses.Line,

				CheckData: map[string]interface{}{
					"unstableReason":      unstableReason,
					"originalGitHubRef":   step.Uses.Value,
					"remediatedGitHubRef": stableRef,
				},
				Message:      message,
				CanRemediate: true,
			}

			issues = append(issues, &i)
		}
		setProgressFunc(jobName, false, true)
	}

	return issues, nil
}

func mustGetIssueMessage(workflowName string, jobName string, unstableReason UnstableReason, step *workflowtypes.Step) string {
	switch unstableReason {
	case IsStable:
		return ""
	case UnknownReason:
		return "unknown reason"
	case UnsupportedRef:
		return "unsupported ref"
	case NoSpecifiedVersion:
		return "no specified version"
	case IsMaster:
		return fmt.Sprintf("The job named %q in the %q workflow is referencing an action on the master branch of the %q repo. The master branch of %q is likely to change.",
			jobName, workflowName, step.Uses.Value, step.Uses.Value)
	case IsBranch:
		branch := strings.Split(step.Uses.Value, "@")[1]
		return fmt.Sprintf("The job named %q in the %q workflow is using an action from %q. This is unstable because %q is a branch, and the contents might change.",
			jobName, workflowName, step.Uses.Value, branch)
	case HasUnstableHistory:
		return "has unsatable history"
	case TagNotFound:
		return "tag not found"
	case NotRecommendedTag:
		return fmt.Sprintf("The job named %q in the %q workflow is referencing an action in the %q repo, but not using a recommended tag.",
			jobName, workflowName, step.Uses.Value)
	}

	return ""
}
