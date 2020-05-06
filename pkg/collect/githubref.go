package collect

import (
	"github.com/pkg/errors"
	"github.com/proactionhq/proaction/pkg/collect/types"
	"github.com/proactionhq/proaction/pkg/logger"
	"go.uber.org/zap"
)

func parseGitHubRef(input string, collectors []string) ([]*types.Output, error) {
	logger.Debug("parseGitHubRef",
		zap.String("input", input),
		zap.Strings("collectors", collectors))

	for _, collector := range collectors {
		if collector == "repoInfo" {

		} else if collector == "branches" {

		} else if collector == "commits" {

		} else if collector == "tags" {

		} else if collector == "recommendations" {

		} else {
			return nil, errors.Errorf("unknown collector %q", collector)
		}
	}
	return nil, nil
}
