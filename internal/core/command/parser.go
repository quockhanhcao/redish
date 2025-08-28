package command

import (
	"bytes"
	"fmt"
)

const CRLF = "\r\n"

func ParseCommand(data []byte) (Command, error) {

	return Command{}, nil
}

func DecodeCommand(data []byte) (Command, error) {
	switch data[0] {
	case '+':
		return decodeSimpleString(data)
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
