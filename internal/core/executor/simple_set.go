package executor

import (
	"errors"
	"fmt"

	"github.com/quockhanhcao/redish/internal/core"
	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
	"github.com/quockhanhcao/redish/internal/data_structure"
)

func cmdSAdd(cmd *command.Command) []byte {
	if len(cmd.Args) <= 1 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	for _, value := range cmd.Args {
		fmt.Printf("///////////////////// value: %s\n", value)
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
	return []byte{}
}

func cmdSMembers(cmd *command.Command) []byte {
	setName := cmd.Args[0]
	existSet, ok := core.StoredSet[setName]
	if !ok {
		return resp_parser.EncodeEmptyArray()
	}
	members := existSet.GetMembers()
	return resp_parser.Encode(members, false)
}
