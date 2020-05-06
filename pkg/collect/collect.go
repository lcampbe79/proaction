package collect

import (
	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/pkg/collect/types"
)

// Collect will run the collector on the workflowContent and return the outputs
func Collect(collector types.Collector, workflowContent []byte) ([]*types.Output, error) {
	inputs, err := pathsToInput(collector.Path, workflowContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert paths to input")
	}

	combinedOutputs := []*types.Output{}
	for _, input := range inputs {
		outputs, err := parseInput(collector.Parser, input, collector.Collectors)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse input")
		}

		combinedOutputs = append(combinedOutputs, outputs...)
	}
	return nil, nil
}

func parseInput(parser string, input string, collectors []string) ([]*types.Output, error) {
	if parser == "githubref" {
		return parseGitHubRef(input, collectors)
	}

	return nil, errors.New("unknown parser")
}
