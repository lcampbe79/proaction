package types

import (
	collecttypes "github.com/proactionhq/proaction/pkg/collect/types"
)

type Check struct {
	Collectors []collecttypes.Collector `json:"collect"`
}
