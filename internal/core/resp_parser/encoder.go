package resp_parser

import (
	"bytes"
	"fmt"
)

func encodeStringArray(value []string) []byte {
	length := len(value)
	res := fmt.Appendf(nil, "*%d\r\n", length)
	for i := range value {
		res = fmt.Appendf(res, "$%d\r\n%s\r\n", len(value[i]), value[i])
	}
	return res
}

func encodeError(err error) []byte {
	return fmt.Appendf(nil, "-%s\r\n", err.Error())
}

func EncodeEmptyArray() []byte {
	return fmt.Appendf(nil, "*0\r\n")
}

func Encode(value interface{}, isSimpleString bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimpleString {
			return fmt.Appendf(nil, "+%s\r\n", v)
		}
		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(v), v)
	case int64, int32, int16, int8, int:
		return fmt.Appendf(nil, ":%d\r\n", v)
	case error:
		return encodeError(v)
	case []string:
		return encodeStringArray(value.([]string))
	case [][]string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, sa := range value.([][]string) {
			buf.Write(encodeStringArray(sa))
		}
		return fmt.Appendf(nil, "*%d\r\n%s", len(value.([][]string)), buf.Bytes())
	case []interface{}:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, x := range value.([]interface{}) {
			buf.Write(Encode(x, false))
		}
		return fmt.Appendf(nil, "*%d\r\n%s", len(value.([]interface{})), buf.Bytes())
	default:
		return []byte("$-1\r\n")
	}
}
