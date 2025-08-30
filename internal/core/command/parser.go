package command

import (
	"bytes"
	"fmt"
	"strconv"
)

const CRLF = "\r\n"

func DecodeCommand(data []byte) (Command, error) {
	switch data[0] {
	case '+':
		return decodeSimpleString(data)
	case '*':
		return decodeBulkString(data)
	default:
		return Command{}, fmt.Errorf("unknown command type: %c", data[0])
	}
}

func decodeSimpleString(data []byte) (Command, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		return Command{}, fmt.Errorf("invalid simple string: no CRLF found")
	}
	return Command{Cmd: string(data[1:idx])}, nil
}

func decodeBulkString(data []byte) (Command, error) {
	firstCRLF := bytes.Index(data, []byte(CRLF))
	if firstCRLF == -1 {
		return Command{}, fmt.Errorf("invalid bulk string: no CRLF found after array length")
	}
	numberOfElements, err := strconv.ParseInt(string(data[1:firstCRLF]), 10, 16)
	if err != nil {
		return Command{}, fmt.Errorf("invalid bulk string: cannot parse number of elements: %w", err)
	}
	fmt.Println("Number of elements:", numberOfElements)
	command := data[1:firstCRLF]
	args := data[firstCRLF+2:]
	argParts := bytes.Split(args, []byte(CRLF))
	var commandArgs = make([]string, 0, len(argParts))
	for _, arg := range argParts {
		if len(arg) == 0 {
			continue
		}
		commandArgs = append(commandArgs, string(arg))
	}
	return Command{Cmd: string(command), Args: commandArgs}, nil
}
