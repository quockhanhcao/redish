package command

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type Command struct {
	Name           string
	NumberOfArgs   int
	Args           []string
	currentArgSize int
	state          CommandParserState
}

type CommandParserState int

const (
	commandParserStateInitialize CommandParserState = iota
	commandParserStateDone
)

const bufferSize = 8
const CRLF = "\r\n"

func CommandFromReader(reader io.Reader) (*Command, error) {
	cmd := &Command{
		state: commandParserStateInitialize,
	}
	buf := make([]byte, bufferSize)
	// index to keep track of the current position in the buffer
	readToIndex := 0

	for cmd.state != commandParserStateDone {
		// increase buffer size if needed
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		readBytes, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				// think about the state of the command here
				// what is the condition for stop parsing?
				break
			}
			return nil, err
		}
		readToIndex += readBytes
		parsedBytes, err := cmd.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		// remove the parsed bytes from the buffer
		copy(buf, buf[parsedBytes:])
		readToIndex -= parsedBytes
	}
	return cmd, nil
}

func (c *Command) parse(data []byte) (int, error) {
	switch c.state {
	case commandParserStateInitialize:
		numberOfArgs, parsedBytes, err := getNumberOfArgs(data)
		if err != nil {
			return 0, err
		}
		if parsedBytes == 0 {
			// Need more data to parse the command
			return 0, nil
		}
		c.NumberOfArgs = numberOfArgs
		c.Args = make([]string, numberOfArgs)
		c.state = commandParserStateDone
		return parsedBytes, nil
	default:
		return 0, fmt.Errorf("unknown command parser state: %d", c.state)
	}
}

func getNumberOfArgs(data []byte) (numberOfArgs, parsedBytes int, err error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		// Need more data
		return 0, 0, nil
	}
	line := data[:idx]
	if line[0] != '*' {
		return 0, 0, fmt.Errorf("wrong format of command's first line")
	}
	numberOfArgs, err = strconv.Atoi(string(line[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse number of params: %w", err)
	}
	return numberOfArgs, idx + len(CRLF), nil
}
