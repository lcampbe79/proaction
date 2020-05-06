package scanner

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/pkg/checks"
	checktypes "github.com/proactionhq/proaction/pkg/checks/types"
	"github.com/proactionhq/proaction/pkg/collect"
	collecttypes "github.com/proactionhq/proaction/pkg/collect/types"
	"github.com/proactionhq/proaction/pkg/issue"
	progresstypes "github.com/proactionhq/proaction/pkg/progress/types"
	workflowtypes "github.com/proactionhq/proaction/pkg/workflow/types"
	"gopkg.in/yaml.v3"
)

type Scanner struct {
	OriginalContent   []byte
	RemediatedContent []byte
	Issues            []*issue.Issue
	EnabledChecks     []*checktypes.Check
	ParsedWorkflow    *workflowtypes.GitHubWorkflow
	JobNames          []string

	Progress map[string]progresstypes.Progress
}

func NewScanner(content []byte) (*Scanner, error) {
	parsedWorkflow := workflowtypes.GitHubWorkflow{}
	if err := yaml.Unmarshal(content, &parsedWorkflow); err != nil {
		return nil, errors.Wrap(err, "failed to parse content")
	}

	jobNames := []string{}
	for jobName := range parsedWorkflow.Jobs {
		jobNames = append(jobNames, jobName)
	}

	return &Scanner{
		OriginalContent: content,
		Issues:          []*issue.Issue{},
		EnabledChecks:   []*checktypes.Check{},
		ParsedWorkflow:  &parsedWorkflow,
		JobNames:        jobNames,
	}, nil
}

func (s *Scanner) EnableChecks(checks []*checktypes.Check) {
	s.EnabledChecks = checks
	s.initProgress()
}

func (s *Scanner) EnableAllChecks() {
	s.EnabledChecks = []*checktypes.Check{
		checks.UnstableGitHubRef(),
	}
	s.initProgress()
}

func (s *Scanner) initProgress() {
	s.Progress = map[string]progresstypes.Progress{}

}

func (s *Scanner) ScanWorkflow() error {
	// build collectors
	collectors := []collecttypes.Collector{}
	for _, enabledCheck := range s.EnabledChecks {
		for _, checkCollector := range enabledCheck.Collectors {
			mergedCollectors, err := collect.MergeCollectors(checkCollector, collectors)
			if err != nil {
				return errors.Wrap(err, "failed to merge collectors")
			}

			collectors = mergedCollectors
		}
	}

	// execute collect phase
	outputs := []*collecttypes.Output{}
	for _, collector := range collectors {
		outputs, err := collect.Collect(collector, s.OriginalContent)
		if err != nil {
			return errors.Wrap(err, "failed to collect collector")
		}

		outputs = append(outputs, outputs...)
	}

	fmt.Printf("%#v\n", outputs)

	return nil
}

func applyRemediation(content string, i issue.Issue) (string, error) {
	return content, nil
}

func (s Scanner) getContent() []byte {
	if s.RemediatedContent != nil {
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
