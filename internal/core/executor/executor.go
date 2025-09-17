package executor

import (
	"errors"
	"strconv"
	"syscall"

	"github.com/quockhanhcao/redish/internal/core"
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

func cmdSet(cmd *command.Command) []byte {
	if len(cmd.Args) < 2 || len(cmd.Args) == 3 || len(cmd.Args) > 4 {
		return resp_parser.Encode(errors.New("(error) syntax error"), false)
	}
	if len(cmd.Args) == 2 {
		core.Dictionary.AddToSet(cmd.Args[0], cmd.Args[1], -1)
	} else {
		expTime, err := strconv.Atoi(cmd.Args[3])
		if err != nil || expTime <= 0 || cmd.Args[2] != "EX" {
			return resp_parser.Encode(errors.New("(error) syntax error"), false)
		}
		core.Dictionary.AddToSet(cmd.Args[0], cmd.Args[1], int64(expTime))
	}
	return []byte("+OK\r\n")
}

func cmdGet(cmd *command.Command) []byte {
	if len(cmd.Args) > 1 {
		return resp_parser.Encode(errors.New("ERR wrong number of arguments for command"), false)
	}
	val, ok := core.Dictionary.GetFromSet(cmd.Args[0])
	if !ok {
		return []byte("$-1\r\n")
	}
	return resp_parser.Encode(val, false)
}

func ExecuteCommand(cmd *command.Command, fd int) error {
	var response []byte
	switch cmd.Cmd {
	case "PING":
		response = cmdPing(cmd)
	case "SET":
		response = cmdSet(cmd)
	case "GET":
		response = cmdGet(cmd)
	default:
		response = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(fd, response)
	return err
}
