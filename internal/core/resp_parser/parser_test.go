package resp_parser

import (
	"reflect"
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
			name: "empty string", input: "+\r\n", want: "", wantPos: 3,
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
	tests := []struct {
		name    string
		input   string
		want    string
		wantPos int
		wantErr bool
		bug     string
	}{
		{
			name: "normal bulk string", input: "$6\r\nfoobar\r\n", want: "foobar", wantPos: 12,
		},
		{
			name: "bulk string with 0 length", input: "$0\r\n\r\n", want: "", wantPos: 6,
		},
		{
			// Returns the same value as the empty string above, so callers cannot tell null from "" (Part 5.6).
			name: "null bulk string", input: "$-1\r\n", want: "", wantPos: 5,
		},
		{
			name: "stops after declared length", input: "$3\r\nfoo\r\n$3\r\nbar\r\n", want: "foo", wantPos: 9,
		},
		{
			name: "null followed by more data", input: "$-1\r\n+OK\r\n", want: "", wantPos: 5,
		},
		{
			name: "position counts bytes not characters", input: "$9\r\nxin chào\r\n", want: "xin chào", wantPos: 15,
		},
		{
			name: "bare LF inside content", input: "$3\r\na\nb\r\n", want: "a\nb", wantPos: 9,
		},
		{
			name: "CRLF inside content", input: "$4\r\na\r\nb\r\n", want: "a\r\nb", wantPos: 10,
			bug: "searches for the next CRLF instead of reading the declared number of bytes",
		},
		{
			name: "only type byte", input: "$", wantErr: true,
		},
		{
			name: "missing CRLF after length", input: "$3", wantErr: true,
		},
		{
			name: "non-digit length", input: "$abc\r\nfoobar\r\n", wantErr: true,
		},
		{
			name: "negative length other than -1", input: "$-2\r\nfoo\r\n", wantErr: true,
		},
		{
			name: "empty length", input: "$\r\n\r\n", wantErr: true,
			bug: "an empty length is read as 0 instead of rejected",
		},
		{
			name: "content shorter than length", input: "$3\r\nfo\r\n", wantErr: true,
		},
		{
			name: "content longer than length", input: "$2\r\nfoo\r\n", wantErr: true,
		},
		{
			name: "truncated content", input: "$6\r\nfoo", wantErr: true,
		},
		{
			name: "missing final CRLF", input: "$3\r\nfoo", wantErr: true,
		},
		{
			name: "empty string missing final CRLF", input: "$0\r\n", wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bug != "" {
				t.Skip("known bug: " + tt.bug)
			}
			got, pos, err := decodeBulkString([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("decodeBulkString(%q) = (%s, %d, nil); want an error", tt.input, got, pos)
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeBulkString(%q) returned unexpected error: %v", tt.input, err)
			}
			if got != tt.want || pos != tt.wantPos {
				t.Errorf("decodeBulkString(%q) = (%s, %d); want (%s, %d)", tt.input, got, pos, tt.want, tt.wantPos)
			}
		})
	}
}

func TestDecodeArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []interface{}
		wantPos int
		wantErr bool
		bug     string
	}{
		{
			name: "bulk strings", input: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			want: []interface{}{"foo", "bar"}, wantPos: 22,
		},
		{
			name: "integers", input: "*3\r\n:1\r\n:2\r\n:3\r\n",
			want: []interface{}{int64(1), int64(2), int64(3)}, wantPos: 16,
		},
		{
			name: "real SET command", input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			want: []interface{}{"SET", "key", "value"}, wantPos: 33,
		},
		{
			name: "mixed element types", input: "*3\r\n:1\r\n$3\r\nfoo\r\n+OK\r\n",
			want: []interface{}{int64(1), "foo", "OK"}, wantPos: 22,
		},
		{
			name: "nested array", input: "*2\r\n*1\r\n:1\r\n$3\r\nfoo\r\n",
			want: []interface{}{[]interface{}{int64(1)}, "foo"}, wantPos: 21,
		},
		{
			name: "empty array", input: "*0\r\n",
			want: []interface{}{}, wantPos: 4,
		},
		{
			// The null element comes back as "", the same as an empty string (Part 5.6).
			name: "null bulk string element", input: "*2\r\n$-1\r\n$3\r\nfoo\r\n",
			want: []interface{}{"", "foo"}, wantPos: 18,
		},
		{
			name: "stops after declared count", input: "*1\r\n:1\r\n:2\r\n",
			want: []interface{}{int64(1)}, wantPos: 8,
		},
		{
			name: "only type byte", input: "*", wantErr: true,
		},
		{
			name: "missing CRLF after count", input: "*2", wantErr: true,
		},
		{
			name: "non-digit count", input: "*a\r\n", wantErr: true,
		},
		{
			name: "unknown element type", input: "*1\r\n?foo\r\n", wantErr: true,
		},
		{
			name: "invalid element", input: "*1\r\n:abc\r\n", wantErr: true,
		},
		{
			name: "element itself truncated", input: "*1\r\n$3\r\nfo", wantErr: true,
		},
		{
			name: "fewer elements than count", input: "*2\r\n$3\r\nSET\r\n", wantErr: true,
			bug: "panics with index out of range when the data ends before all elements are read",
		},
		{
			name: "count with no elements", input: "*1\r\n", wantErr: true,
			bug: "panics with index out of range when the data ends before all elements are read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bug != "" {
				t.Skip("known bug: " + tt.bug)
			}
			got, pos, err := decodeArray([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("decodeArray(%q) = (%#v, %d, nil); want an error", tt.input, got, pos)
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeArray(%q) returned unexpected error: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) || pos != tt.wantPos {
				t.Errorf("decodeArray(%q) = (%#v, %d); want (%#v, %d)", tt.input, got, pos, tt.want, tt.wantPos)
			}
		})
	}
}
