package command

import (
	"fmt"
	"testing"
)

func TestDecodeInteger(t *testing.T) {
	tests := []struct {
		data           []byte
		expectedResult struct {
			value int64
			pos   int
			err   error
		}
	}{
		{data: []byte(":1000\r\n"), expectedResult: struct {
			value int64
			pos   int
			err   error
		}{value: 1000, pos: 7, err: nil}},
		{data: []byte(":-1000\r\n"), expectedResult: struct {
			value int64
			pos   int
			err   error
		}{value: -1000, pos: 8, err: nil}},
		{data: []byte(":+1000\r\n"), expectedResult: struct {
			value int64
			pos   int
			err   error
		}{value: 1000, pos: 8, err: nil}},
		{data: []byte(":0\r\n"), expectedResult: struct {
			value int64
			pos   int
			err   error
		}{value: 0, pos: 4, err: nil}},
		{data: []byte(":1234567890\r\n"), expectedResult: struct {
			value int64
			pos   int
			err   error
		}{value: 1234567890, pos: 13, err: nil}},
		{
			data: []byte(":12345"), expectedResult: struct {
				value int64
				pos   int
				err   error
			}{value: 0, pos: 0, err: fmt.Errorf("invalid integer: no CRLF found")},
		},
		{
			data: []byte(":+*12345"), expectedResult: struct {
				value int64
				pos   int
				err   error
			}{value: 0, pos: 0, err: fmt.Errorf("invalid integer: no CRLF found")},
		},
		{
			data: []byte(":+*12345\r\n"), expectedResult: struct {
				value int64
				pos   int
				err   error
			}{value: 0, pos: 0, err: fmt.Errorf("invalid integer: non-digit character found")},
		},
	}

	for _, test := range tests {
		value, pos, err := decodeInteger(test.data)
		if value != test.expectedResult.value || pos != test.expectedResult.pos || (err != nil && test.expectedResult.err != nil && err.Error() != test.expectedResult.err.Error()) {
			t.Errorf("decodeInteger(%q) = (%d, %d, %v); want (%d, %d, %v)", test.data, value, pos, err, test.expectedResult.value, test.expectedResult.pos, test.expectedResult.err)
		}
	}
}

// func TestDecode
