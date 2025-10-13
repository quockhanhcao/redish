package resp_parser

import "fmt"

func EncodeSimpleString(s string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", s))
}

func EncodeBulkString(s string) []byte {
	if s == "" {
		return []byte("$0\r\n\r\n")
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func EncodeError(err error) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", err.Error()))
}

func Encode(value interface{}, isSimpleString bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimpleString {
			return fmt.Appendf(nil, "+%s\r\n", v)
		}
		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(v), v)
	case int:
		return fmt.Appendf(nil, ":%d\r\n", v)
	case int64:
		return fmt.Appendf(nil, ":%d\r\n", v)
	case error:
		return fmt.Appendf(nil, "-%s\r\n", v)
	default:
		return EncodeError(fmt.Errorf("unsupported type: %T", v))
	}
}
