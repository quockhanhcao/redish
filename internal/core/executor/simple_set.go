package executor

import (
	"errors"

	"github.com/quockhanhcao/redish/internal/core"
	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
	"github.com/quockhanhcao/redish/internal/data_structure"
)

func cmdSAdd(cmd *command.Command) []byte {
	if len(cmd.Args) <= 1 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	setName := cmd.Args[0]
	existSet, ok := core.StoredSet[setName]
	if !ok {
		core.StoredSet[setName] = data_structure.NewSimpleSet()
		existSet = core.StoredSet[setName]
	}
	addedMem := existSet.Add(cmd.Args[1:]...)
	return resp_parser.Encode(addedMem, true)
}

func cmdSRem(cmd *command.Command) []byte {
	if len(cmd.Args) <= 1 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	existSet, ok := core.StoredSet[cmd.Args[0]]
	if !ok {
		// no such set exist, return 0
		return resp_parser.Encode(0, true)
	}
	removedMem := existSet.Remove(cmd.Args[1:]...)
	return resp_parser.Encode(removedMem, true)
}

func cmdSMembers(cmd *command.Command) []byte {
	if len(cmd.Args) != 1 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	// 1st args is set name
	existSet, ok := core.StoredSet[cmd.Args[0]]
	if !ok {
		return resp_parser.EncodeEmptyArray()
	}
	members := existSet.GetMembers()
	return resp_parser.Encode(members, false)
}

func cmdSIsmember(cmd *command.Command) []byte {
	if len(cmd.Args) != 2 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	existSet, ok := core.StoredSet[cmd.Args[0]]
	if !ok {
		return resp_parser.Encode(0, true)
	}
	isMember := existSet.Exist(cmd.Args[1])
	return resp_parser.Encode(isMember, true)
}
