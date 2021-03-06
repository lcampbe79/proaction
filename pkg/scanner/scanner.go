package scanner

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	outdatedaction "github.com/proactionhq/proaction/pkg/checks/outdated-action"
	unforkaction "github.com/proactionhq/proaction/pkg/checks/unfork-action"
	unstabledockertag "github.com/proactionhq/proaction/pkg/checks/unstable-docker-tag"
	unstablegithubref "github.com/proactionhq/proaction/pkg/checks/unstable-github-ref"
	"github.com/proactionhq/proaction/pkg/issue"
	progresstypes "github.com/proactionhq/proaction/pkg/progress/types"
	workflowtypes "github.com/proactionhq/proaction/pkg/workflow/types"
	"gopkg.in/yaml.v3"
)

type Scanner struct {
	OriginalContent   string
	RemediatedContent string
	Issues            []*issue.Issue
	EnabledChecks     []string
	ParsedWorkflow    *workflowtypes.GitHubWorkflow
	JobNames          []string

	Progress map[string]progresstypes.Progress
}

func NewScanner(content string) (*Scanner, error) {
	parsedWorkflow := workflowtypes.GitHubWorkflow{}
	if err := yaml.Unmarshal([]byte(content), &parsedWorkflow); err != nil {
		return nil, errors.Wrap(err, "failed to parse content")
	}

	jobNames := []string{}
	for jobName := range parsedWorkflow.Jobs {
		jobNames = append(jobNames, jobName)
	}

	return &Scanner{
		OriginalContent: content,
		Issues:          []*issue.Issue{},
		EnabledChecks:   []string{},
		ParsedWorkflow:  &parsedWorkflow,
		JobNames:        jobNames,
	}, nil
}

func (s *Scanner) EnableChecks(checks []string) {
	s.EnabledChecks = checks
	s.initProgress()
}

func (s *Scanner) EnableAllChecks() {
	s.EnabledChecks = []string{
		"unfork-action",
		"unstable-docker-tag",
		"unstable-github-ref",
		"outdated-action",
	}
	s.initProgress()
}

func (s *Scanner) initProgress() {
	s.Progress = map[string]progresstypes.Progress{}

	for _, enabledCheck := range s.EnabledChecks {
		// not all checks will use job segmentation for status
		if enabledCheck == "unstable-github-ref" {
			progress := progresstypes.Progress{}
			for _, jobName := range s.JobNames {
				progress.Set(jobName, false, false)
			}
			s.Progress[enabledCheck] = progress
		} else if enabledCheck == "unfork-action" {
			progress := progresstypes.Progress{}
			for _, jobName := range s.JobNames {
				progress.Set(jobName, false, false)
			}
			s.Progress[enabledCheck] = progress
		} else if enabledCheck == "unstable-docker-tag" {
			progress := progresstypes.Progress{}
			for _, jobName := range s.JobNames {
				progress.Set(jobName, false, false)
			}
			s.Progress[enabledCheck] = progress
		} else if enabledCheck == "outdated-action" {
			progress := progresstypes.Progress{}
			for _, jobName := range s.JobNames {
				progress.Set(jobName, false, false)
			}
			s.Progress[enabledCheck] = progress
		}
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
			progress := s.Progress[check]
			issues, err := unstablegithubref.DetectIssues(parsedWorkflow, progress.Set)
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
		} else if check == "unstable-docker-tag" {
			progress := s.Progress[check]
			issues, err := unstabledockertag.DetectIssues(parsedWorkflow, progress.Set)
			if err != nil {
				return errors.Wrap(err, "failed to run unstable unstable-docker-tag check")
			}

			s.Issues = append(s.Issues, issues...)

			for _, i := range issues {
				updated, err := unstabledockertag.RemediateIssue(s.getContent(), i)
				if err != nil {
					return errors.Wrap(err, "failed to apply remediation")
				}
				s.RemediatedContent = updated
			}
		} else if check == "outdated-action" {
			progress := s.Progress[check]
			issues, err := outdatedaction.DetectIssues(parsedWorkflow, progress.Set)
			if err != nil {
				return errors.Wrap(err, "failed to run unstable outdated-action check")
			}

			s.Issues = append(s.Issues, issues...)

			for _, i := range issues {
				updated, err := outdatedaction.RemediateIssue(s.getContent(), i)
				if err != nil {
					return errors.Wrap(err, "failed to apply remediation")
				}
				s.RemediatedContent = updated
			}
		} else if check == "unfork-action" {
			progress := s.Progress[check]
			issues, err := unforkaction.DetectIssues(parsedWorkflow, progress.Set)
			if err != nil {
				return errors.Wrap(err, "failed to run unstable unfork-action check")
			}

			s.Issues = append(s.Issues, issues...)

			for _, i := range issues {
				updated, err := unforkaction.RemediateIssue(s.getContent(), i)
				if err != nil {
					return errors.Wrap(err, "failed to apply remediation")
				}
				s.RemediatedContent = updated
			}
		}
	}

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
