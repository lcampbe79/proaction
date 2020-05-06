package checks

import (
	"github.com/proactionhq/proaction/pkg/checks/types"
	collecttypes "github.com/proactionhq/proaction/pkg/collect/types"
)

func UnstableGitHubRef() *types.Check {
	return &types.Check{
		Collectors: []collecttypes.Collector{
			{
				Name:   "uses",
				Path:   "jobs[*].steps[*].uses",
				Parser: "githubref",
				Collectors: []string{
					"repoInfo",
					"branches",
					"tags",
					"commits",
					"recommendations",
				},
			},
		},
	}
}
