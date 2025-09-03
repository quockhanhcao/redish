package resp_parser

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/quockhanhcao/redish/internal/core/command"
)

const (
	CRLF             = "\r\n"
	NULL_BULK_STRING = "$-1\r\n"
)

func ParseCommand(data []byte) (command.Command, error) {
	args, _, err := decodeCommand(data)
	if err != nil {
		return command.Command{}, err
	}
	argList, ok := args.([]interface{})
	if !ok {
		return command.Command{}, fmt.Errorf("invalid command format")
	}
	cmd, ok := argList[0].(string)
	if !ok {
		return command.Command{}, fmt.Errorf("command name must be a string")
	}
	strArgs := make([]string, 0, len(argList)-1)
	for _, arg := range argList[1:] {
		strArg, ok := arg.(string)
		if !ok {
			return command.Command{}, fmt.Errorf("command arguments must be strings")
		}
		strArgs = append(strArgs, strArg)
	}
	return command.Command{
		Cmd:  strings.ToUpper(cmd),
		Args: strArgs,
	}, nil
}

func decodeCommand(data []byte) (interface{}, int, error) {
	switch data[0] {
	case '+':
		return decodeSimpleString(data)
	case ':':
		return decodeInteger(data)
	case '$':
		return decodeBulkString(data)
	case '*':
		return decodeArray(data)
	default:
		return nil, 0, fmt.Errorf("unknown command type: %c", data[0])
	}
}

func readLen(data []byte) (int, int, error) {
	crlfIdx := bytes.Index(data, []byte(CRLF))
	if crlfIdx == -1 {
		return 0, 0, fmt.Errorf("invalid length: no CRLF found")
	}
	length := 0
	for i := 1; i < crlfIdx; i++ {
		if !unicode.IsDigit(rune(data[i])) {
			return 0, 0, fmt.Errorf("invalid length: non-digit character found")
		}
		length = length*10 + int(data[i]-'0')
	}
	return length, crlfIdx + len(CRLF), nil
}

// Eg: +OK\r\n
func decodeSimpleString(data []byte) (string, int, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		return "", 0, fmt.Errorf("invalid simple string: no CRLF found")
	}
	return string(data[1:idx]), idx + len(CRLF), nil
}

// Eg: :1000\r\n, :-1000\r\n, :+1000\r\n
func decodeInteger(data []byte) (int64, int, error) {
	crlfIdx := bytes.Index(data, []byte(CRLF))
	if crlfIdx == -1 {
		return 0, 0, fmt.Errorf("invalid integer: no CRLF found")
	}
	sign := 1
	pos := 1
	if data[1] == '-' {
		sign = -1
		pos++
	} else if data[1] == '+' {
		pos++
	}
	value := int64(0)
	for i := pos; i < crlfIdx; i++ {
		if !unicode.IsDigit(rune(data[i])) {
			return 0, 0, fmt.Errorf("invalid integer: non-digit character found")
		}
		value = value*10 + int64(data[i]-'0')
	}
	return int64(sign) * value, crlfIdx + len(CRLF), nil
}

// Eg: $6\r\nfoobar\r\n, $0\r\n\r\n, $-1\r\n
func decodeBulkString(data []byte) (string, int, error) {
	// null bulk string
	crlfIdx := bytes.Index(data, []byte(CRLF))
	if bytes.Equal(data[:crlfIdx+len(CRLF)], []byte(NULL_BULK_STRING)) {
		return "", crlfIdx + len(CRLF), nil
	}
	strLen, pos, err := readLen(data)
	if err != nil {
		return "", 0, err
	}
	secondCRLF := bytes.Index(data[pos:], []byte(CRLF))
	if secondCRLF == -1 {
		return "", 0, fmt.Errorf("invalid bulk string: no CRLF found after string data")
	}
	str := string(data[pos : pos+secondCRLF])
	if len(str) != strLen {
		return "", 0, fmt.Errorf("invalid bulk string: expected length %d, got %d", strLen, len(str))
	}
	return str, pos + strLen + len(CRLF), nil
}

// *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n
func decodeArray(data []byte) ([]interface{}, int, error) {
	numArgs, pos, err := readLen(data)
	if err != nil {
		return nil, 0, err
	}
	args := make([]interface{}, 0, numArgs)
	for i := 0; i < numArgs; i++ {
		arg, readBytes, err := decodeCommand(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		pos += readBytes
		args = append(args, arg)
	}
	return args, pos, nil
}
