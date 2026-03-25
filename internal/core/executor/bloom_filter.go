package executor

import (
	"errors"
	"fmt"

	"github.com/quockhanhcao/redish/internal/core"
	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
	"github.com/quockhanhcao/redish/internal/data_structure"
)

const BfDefaultInitCapacity = 100
const BfDefaultErrRate = 0.01

func cmdBFADD(cmd *command.Command) []byte {
	if len(cmd.Args) != 2 {
		return resp_parser.Encode(errors.New("wrong number of arguments for command"), false)
	}
	var bloom data_structure.Bloom
	if _, ok := core.StoredBloomFilter[cmd.Args[0]]; !ok {
		bloom = *data_structure.NewBloomFilter(BfDefaultInitCapacity, BfDefaultErrRate)
	}
	bloom.Add(cmd.Args[1])
	return resp_parser.Encode(fmt.Sprintf("%d", 1), false)
}
