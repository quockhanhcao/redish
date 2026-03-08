package executor

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/quockhanhcao/redish/internal/core"
	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
	"github.com/quockhanhcao/redish/internal/data_structure"
)

func cmdInitCMS(cmd *command.Command) []byte {
	if len(cmd.Args) != 3 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	cmsName := cmd.Args[0]
	if _, ok := core.StoredCountMinSketch[cmsName]; ok {
		return resp_parser.Encode(errors.New("ERR item exists"), false)
	}

	errorRate, err := strconv.ParseFloat(cmd.Args[1], 64)
	if err != nil || errorRate <= 0 || errorRate >= 1 {
		return resp_parser.Encode(fmt.Errorf("invalid error rate: %s (must be between 0 and 1)", cmd.Args[1]), false)
	}

	probabilityRate, err := strconv.ParseFloat(cmd.Args[2], 64)
	if err != nil || probabilityRate <= 0 || probabilityRate >= 1 {
		return resp_parser.Encode(fmt.Errorf("invalid probability: %s (must be between 0 and 1)", cmd.Args[2]), false)
	}
	newCMS := data_structure.NewCountMinSketch(errorRate, probabilityRate)
	core.StoredCountMinSketch[cmsName] = newCMS

	return resp_parser.Encode("OK", true)
}

func cmdIncrCMS(cmd *command.Command) []byte {
	if len(cmd.Args) < 3 || len(cmd.Args)%2 == 0 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}

	// TODO: finish adding value to key
	return []byte{}
}
