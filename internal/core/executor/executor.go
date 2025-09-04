package executor

import (
	"syscall"

	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
)

func ExecuteCommand(cmd command.Command, fd int) {
	var response []byte
	switch cmd.Cmd {
	case "PING":
		if len(cmd.Args) == 0 {
			response = resp_parser.EncodeSimpleString("PONG")
		} else if len(cmd.Args) == 1 {
			response = resp_parser.EncodeBulkString(cmd.Args[0])
		} else {
			response = resp_parser.EncodeError("(error) wrong number of arguments for 'ping' command")
		}
	}
	syscall.Write(fd, response)
}
