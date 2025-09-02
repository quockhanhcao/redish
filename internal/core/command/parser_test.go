package command

import (
	"fmt"
	"testing"
)

func TestDecodeInteger(t *testing.T) {
	type response struct {
		value int64
		pos   int
		err   error
	}
	tests := []struct {
		data           []byte
		expectedResult response
	}{
		{data: []byte(":1000\r\n"), expectedResult: response{value: 1000, pos: 7, err: nil}},
		{data: []byte(":-1000\r\n"), expectedResult: response{value: -1000, pos: 8, err: nil}},
		{data: []byte(":+1000\r\n"), expectedResult: response{value: 1000, pos: 8, err: nil}},
		{data: []byte(":0\r\n"), expectedResult: response{value: 0, pos: 4, err: nil}},
		{data: []byte(":1234567890\r\n"), expectedResult: response{value: 1234567890, pos: 13, err: nil}},
		{
			data: []byte(":12345"), expectedResult: response{value: 0, pos: 0, err: fmt.Errorf("invalid integer: no CRLF found")},
		},
		{
			data: []byte(":+*12345"), expectedResult: response{value: 0, pos: 0, err: fmt.Errorf("invalid integer: no CRLF found")},
		},
		{
			data: []byte(":+*12345\r\n"), expectedResult: response{value: 0, pos: 0, err: fmt.Errorf("invalid integer: non-digit character found")},
		},
	}

	for _, test := range tests {
		value, pos, err := decodeInteger(test.data)
		if value != test.expectedResult.value || pos != test.expectedResult.pos || (err != nil && test.expectedResult.err != nil && err.Error() != test.expectedResult.err.Error()) {
			t.Errorf("decodeInteger(%q) = (%d, %d, %v); want (%d, %d, %v)", test.data, value, pos, err, test.expectedResult.value, test.expectedResult.pos, test.expectedResult.err)
		}
	}
}

func TestDecodeSimpleString(t *testing.T) {
	type response struct {
		value string
		pos   int
		err   error
	}
	tests := []struct {
		data           []byte
		expectedResult response
	}{
		{data: []byte("+OK\r\n"), expectedResult: response{value: "OK", pos: 5, err: nil}},
		{data: []byte("+PONG\r\n"), expectedResult: response{
			value: "PONG", pos: 7, err: nil,
		}},
		{
			data: []byte("+Hello"), expectedResult: response{value: "", pos: 0, err: fmt.Errorf("invalid simple string: no CRLF found")},
		},
	}
	for _, test := range tests {
		value, pos, err := decodeSimpleString(test.data)
		if value != test.expectedResult.value || pos != test.expectedResult.pos || (err != nil && test.expectedResult.err != nil && err.Error() != test.expectedResult.err.Error()) {
			t.Errorf("decodeSimpleString(%q) = (%s, %d, %v); want (%s, %d, %v)", test.data, value, pos, err, test.expectedResult.value, test.expectedResult.pos, test.expectedResult.err)
		}
	}
}

func TestDecodeBulkString(t *testing.T) {
	type response struct {
		value string
		pos   int
		err   error
	}
	tests := []struct {
		data           []byte
		expectedResult response
	}{
		{data: []byte("$3\r\nfoo"), expectedResult: response{value: "", pos: 0, err: fmt.Errorf("invalid bulk string: no CRLF found after string data")}},
		{data: []byte("$3\r\nfo\r\n"), expectedResult: response{value: "", pos: 0, err: fmt.Errorf("invalid bulk string: expected length 3, got 2")}},
		{data: []byte("$abc\r\nfoobar\r\n"), expectedResult: response{value: "", pos: 0, err: fmt.Errorf("invalid bulk string length: non-digit character found")}},
		{data: []byte("$6\r\nfoobar\r\nabc"), expectedResult: response{value: "foobar", pos: 12, err: nil}},
		{data: []byte("$0\r\n\r\n"), expectedResult: response{value: "", pos: 6, err: nil}},
		{data: []byte("$-1\r\n"), expectedResult: response{value: "", pos: 5, err: nil}},
	}
	for _, test := range tests {
		value, pos, err := decodeBulkString(test.data)
		if value != test.expectedResult.value || pos != test.expectedResult.pos || (err != nil && test.expectedResult.err != nil && err.Error() != test.expectedResult.err.Error()) {
			t.Errorf("decodeBulkString(%q) = (%s, %d, %v); want (%s, %d, %v)", test.data, value, pos, err, test.expectedResult.value, test.expectedResult.pos, test.expectedResult.err)
		}
	}
}

func TestDecodeArray(t *testing.T) {
	type response struct {
		value []interface{}
		pos   int
		err   error
	}
	tests := []struct {
		data           []byte
		expectedResult response
	}{
		{
			data: []byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"),
			expectedResult: response{
				value: []interface{}{"foo", "bar"},
				pos:   22, err: nil,
			}},
	}
	for _, test := range tests {
		value, pos, err := decodeArray(test.data)
		if fmt.Sprint(value) != fmt.Sprint(test.expectedResult.value) || pos != test.expectedResult.pos || (err != nil && test.expectedResult.err != nil && err.Error() != test.expectedResult.err.Error()) {
			t.Errorf("decodeArray(%q) = (%v, %d, %v); want (%v, %d, %v)", test.data, value, pos, err, test.expectedResult.value, test.expectedResult.pos, test.expectedResult.err)
		}
	}
}
