package executor

import (
	"errors"
	"syscall"

	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
)

func cmdPing(cmd *command.Command) []byte {

	if len(cmd.Args) == 0 {
		return resp_parser.Encode("PONG", true)
	}
	if len(cmd.Args) == 1 {
		return resp_parser.Encode(cmd.Args[0], false)
	}
	return resp_parser.Encode(errors.New("wrong number of arguments for 'ping' command"), false)
}

func ExecuteCommand(cmd *command.Command, fd int) error {
	var response []byte
	switch cmd.Cmd {
	case "PING":
		response = cmdPing(cmd)
	default:
		response = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(fd, response)
	return err
}
