package resp_parser

import "fmt"

func encodeStringArray(value []string) []byte {
	length := len(value)
	res := fmt.Appendf(nil, "*%d\r\n", length)
	for i := range value {
		res = fmt.Appendf(res, "$%d\r\n%s\r\n", len(value[i]), value[i])
	}
	return res
}

func encodeError(err error) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", err.Error()))
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
	case int, int64:
		return fmt.Appendf(nil, ":%d\r\n", v)
	case error:
		return fmt.Appendf(nil, "-%s\r\n", v)
	case []string:
		return encodeStringArray(value.([]string))
	default:
		return encodeError(fmt.Errorf("unsupported type: %T", v))
	}
}
