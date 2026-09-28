package executor

import (
	"errors"

	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/config"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
)

func cmdCONFIG(cmd *command.Command) []byte {
	if len(cmd.Args) != 2 {
		return resp_parser.Encode(errors.New("syntax error"), false)
	}
	resp := []string{
		config.Persistence.Directory,
		config.Persistence.DBFileName,
	}
	return resp_parser.Encode(resp, false)
}
