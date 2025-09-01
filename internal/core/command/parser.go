package command

import (
	"bytes"
	"fmt"
)

const (
	CRLF             = "\r\n"
	NULL_BULK_STRING = "$-1\r\n"
)

func DecodeCommand(data []byte) (interface{}, int, error) {
	switch data[0] {
	case '+':
		return decodeSimpleString(data)
	case ':':
		return decodeInteger(data)
	case '$':
		return decodeBulkString(data)
	default:
		return nil, 0, fmt.Errorf("unknown command type: %c", data[0])
	}
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
func decodeInteger(data []byte) (value int64, pos int, err error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		return 0, 0, fmt.Errorf("invalid integer: no CRLF found")
	}
	sign := 1
	pos = 1
	if data[1] == '-' {
		sign = -1
		pos++
	} else if data[1] == '+' {
		pos++
	}
	value = int64(0)
	for i := pos; i < idx; i++ {
		value = value*10 + int64(data[i]-'0')
	}
	return int64(sign) * value, idx + len(CRLF), nil
}

// Eg: $6\r\nfoobar\r\n, $0\r\n\r\n, $-1\r\n
func decodeBulkString(data []byte) (string, int, error) {
	firstCRLF := bytes.Index(data, []byte(CRLF))
	if firstCRLF == -1 {
		return "", 0, fmt.Errorf("invalid bulk string: no CRLF found after array length")
	}
	if bytes.Equal(data[1:firstCRLF], []byte(NULL_BULK_STRING)) {
		return "", firstCRLF + len(CRLF), nil
	}
	stringLength := 0
	pos := 1
	for i := 1; i < firstCRLF; i++ {
		stringLength = stringLength*10 + int(data[i]-'0')
		pos++
	}
	secondCRLF := bytes.Index(data[firstCRLF+len(CRLF):], []byte(CRLF))
	if secondCRLF == -1 {
		return "", 0, fmt.Errorf("invalid bulk string: no CRLF found after string data")
	}
	str := string(data[pos+len(CRLF) : secondCRLF])
	if len(str) != stringLength {
		return "", 0, fmt.Errorf("invalid bulk string: expected length %d, got %d", stringLength, len(str))
	}
	return str, secondCRLF + len(CRLF), nil
}
