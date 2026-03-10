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

func cmdCMSINITBYPROB(cmd *command.Command) []byte {
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

type increaseCommand struct {
	key   string
	value uint64
}

func cmdCMSINCRBY(cmd *command.Command) []byte {
	if len(cmd.Args) < 3 || len(cmd.Args)%2 == 0 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	sketchName := cmd.Args[0]
	sketch, ok := core.StoredCountMinSketch[sketchName]
	if !ok {
		return resp_parser.Encode("ERR CMS: key does not exist", false)
	}
	capacity := (len(cmd.Args) - 1) / 2
	commands := make([]increaseCommand, 0, capacity)
	for idx := 1; idx < len(cmd.Args); idx += 2 {
		value, err := strconv.ParseUint(cmd.Args[idx+1], 10, 64)
		if err != nil {
			return resp_parser.Encode(fmt.Errorf("invalid increment %s", cmd.Args[idx+1]), false)
		}
		command := increaseCommand{
			key:   cmd.Args[idx],
			value: value,
		}
		commands = append(commands, command)
	}
	res := make([]string, 0, capacity)
	for _, command := range commands {
		count := sketch.Increase(command.key, command.value)
		res = append(res, fmt.Sprintf("%d", count))
	}

	return resp_parser.Encode(res, false)
}

func cmdCMSINFO(cmd *command.Command) []byte {
	if len(cmd.Args) != 1 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	sketch, ok := core.StoredCountMinSketch[cmd.Args[0]]
	if !ok {
		return resp_parser.Encode("ERR CMS: key does not exist", false)
	}
	// 6 for 3 params in cms: width, depth, and count
	res := make([]string, 0, 6)
	res = append(res, "width")
	res = append(res, fmt.Sprintf("%d", sketch.GetWidth()))
	res = append(res, "depth")
	res = append(res, fmt.Sprintf("%d", sketch.GetDepth()))
	res = append(res, "count")
	res = append(res, fmt.Sprintf("%d", sketch.GetTotalCount()))
	return resp_parser.Encode(res, false)
}

func cmdCMSQUERY(cmd *command.Command) []byte {
	if len(cmd.Args) < 1 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	sketch, ok := core.StoredCountMinSketch[cmd.Args[0]]
	if !ok {
		return resp_parser.Encode("ERR CMS: key does not exist", false)
	}
	res := make([]string, 0, len(cmd.Args)-1)
	for i := 1; i < len(cmd.Args); i++ {
		res = append(res, fmt.Sprintf("%d", sketch.GetMember(cmd.Args[i])))
	}
	return resp_parser.Encode(res, false)
}
