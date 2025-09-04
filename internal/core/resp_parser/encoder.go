package resp_parser

import "fmt"

func EncodeSimpleString(s string) []byte {
	return []byte("+" + s + "\r\n")
}

func EncodeBulkString(s string) []byte {
	if s == "" {
		return []byte("$0\r\n\r\n")
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func EncodeError(error string) []byte {
	return []byte("-" + error + "\r\n")
}
