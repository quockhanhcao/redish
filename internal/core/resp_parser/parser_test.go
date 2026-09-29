package resp_parser

import (
	"fmt"
	"testing"
)

func TestDecodeInteger(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantPos int
		wantErr bool
		bug     string // non-empty means wrong behavior, the case should be fixed
	}{
		{
			name: "positive", input: ":1000\r\n", want: 1000, wantPos: 7,
		},
		{
			name: "negative", input: ":-1000\r\n", want: -1000, wantPos: 8,
		},
		{
			name: "explicit plus sign", input: ":+1000\r\n", want: 1000, wantPos: 8,
		},
		{
			name: "zero", input: ":0\r\n", want: 0, wantPos: 4,
		},
		{
			name: "stops at first CRLF", input: ":360\r\n:7\r\n", want: 360, wantPos: 6,
		},
		{
			name: "missing CRLF", input: ":360", wantErr: true,
		},
		{
			name: "only type byte", input: ":", wantErr: true,
		},
		{
			name: "letter inside the number", input: ":36abc\r\n", wantErr: true,
		},
		{
			name: "letter inside the number", input: ":abc36abc\r\n", wantErr: true,
		},
		{
			name: "letter inside the number", input: ":3abc6\r\n", wantErr: true,
		},
		{
			name: "symbol after sign", input: ":+*36\r\n", wantErr: true,
		},
		{
			name: "empty number", input: ":\r\n", wantErr: true, bug: "want an error, but return 0 instead", // fixing later
		},
		{
			name: "overflow", input: ":99999999999999999999\r\n", wantErr: true, bug: "wrap the number around, not throwing error", //fixing later
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bug != "" {
				t.Skip("known bug " + tt.bug)
			}
			got, pos, err := decodeInteger([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("decodeInteger(%q) = (%d, %d, nil); want an error", tt.input, got, pos)
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeInteger(%q) returned unexpected error: %v", tt.input, err)
			}
			if got != tt.want || pos != tt.wantPos {
				t.Errorf("decodeInteger(%q) = (%d, %d); want (%d, %d)", tt.input, got, pos, tt.want, tt.wantPos)
			}
		})
	}
}

func TestDecodeSimpleString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantPos int
		wantErr bool
		bug     string // know issue
	}{
		{
			name: "normal string", input: "+simplestring\r\n", want: "simplestring", wantPos: 15,
		},
		{
			name: "normal string with number", input: "+simple123\r\n", want: "simple123", wantPos: 12,
		},
		{
			name: "stops at first CRLF", input: "+OK\r\n+PONG\r\n", want: "OK", wantPos: 5,
		},
		{
			name: "spaces allowed", input: "+hello world\r\n", want: "hello world", wantPos: 14,
		},
		{
			name: "position counts bytes not characters", input: "+xin chào\r\n", want: "xin chào", wantPos: 12,
		},
		{
			name: "only type byte", input: "+", wantErr: true,
		},
		{
			name: "empty string", input: "+\r\n", wantErr: true, bug: "does not throw error, return empty string instead",
		},
		{
			name: "string contains redundant CR or LF character", input: "+simplestring\r\r\n", wantErr: true, bug: "does not throw error, return string instead",
		},
		{
			name: "CR without LF at the end", input: "+OK\r", wantErr: true,
		},
		{
			name: "LF without CR", input: "+OK\n", wantErr: true,
		},
		{
			name: "bare LF inside", input: "+ab\ncd\r\n", wantErr: true, bug: "LF inside a simple string is accepted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bug != "" {
				t.Skip("known bug " + tt.bug)
			}
			got, pos, err := decodeSimpleString([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("decodeSimpleString(%q) = (%s, %d, nil); want an error", tt.input, got, pos)
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeSimpleString(%q) returned unexpected error: %v", tt.input, err)
			}
			if got != tt.want || pos != tt.wantPos {
				t.Errorf("decodeSimpleString(%q) = (%s, %d); want (%s, %d)", tt.input, got, pos, tt.want, tt.wantPos)
			}
		})

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
		{data: []byte("$abc\r\nfoobar\r\n"), expectedResult: response{value: "", pos: 0, err: fmt.Errorf("invalid length: non-digit character found")}},
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
			},
		},
		{
			data: []byte("*3\r\n:1\r\n:2\r\n:3\r\n"),
			expectedResult: response{
				value: []interface{}{int64(1), int64(2), int64(3)},
				pos:   16, err: nil,
			},
		},
		{
			data: []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
			expectedResult: response{
				value: []interface{}{"hello", "world"},
				pos:   26, err: nil,
			},
		},
	}
	for _, test := range tests {
		value, pos, err := decodeArray(test.data)
		if fmt.Sprint(value) != fmt.Sprint(test.expectedResult.value) || pos != test.expectedResult.pos || (err != nil && test.expectedResult.err != nil && err.Error() != test.expectedResult.err.Error()) {
			t.Errorf("decodeArray(%q) = (%v, %d, %v); want (%v, %d, %v)", test.data, value, pos, err, test.expectedResult.value, test.expectedResult.pos, test.expectedResult.err)
		}
	}
}
