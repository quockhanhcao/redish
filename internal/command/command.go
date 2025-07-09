package command

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
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
	commandParserStateArgSize
	commandParserStateArgBody
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
				if cmd.state == commandParserStateArgBody {
					return nil, fmt.Errorf("incomplete command data, need more data to parse the command parameters")
				} else if cmd.state == commandParserStateArgSize {
					if cmd.NumberOfArgs == len(cmd.Args) {
						// cmd is fully parsed
						cmd.Name = strings.ToLower(cmd.Args[0])
						cmd.state = commandParserStateDone
						return cmd, nil
					}
					return nil, fmt.Errorf("incomplete command data, need more params")
				}
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
		copy(buf, buf[parsedBytes:readToIndex])
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
		c.state = commandParserStateArgSize
		return parsedBytes, nil
	case commandParserStateArgSize:
		parsedBytes, err := c.parseCommandArgSize(data)
		if err != nil {
			return 0, err
		}
		if parsedBytes == 0 {
			// need more data
			return 0, nil
		}
		return parsedBytes, nil
	case commandParserStateArgBody:
		parsedBytes, err := c.parseCommandArgBody(data)
		if err != nil {
			return 0, err
		}
		if parsedBytes == 0 {
			// need more data
			return 0, nil
		}
		return parsedBytes, err
	case commandParserStateDone:
		return 0, fmt.Errorf("command is already parsed")
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

// return number of bytes parsed and error if any
func (c *Command) parseCommandArgSize(data []byte) (int, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		// need more data
		return 0, nil
	}
	part := data[:idx]
	if part[0] != '$' {
		return 0, fmt.Errorf("wrong format of command's param size line")
	}
	size, err := strconv.Atoi(string(part[1:]))
	if err != nil {
		return 0, fmt.Errorf("failed to parse command's param size: %w", err)
	}
	c.currentArgSize = size
	// if parsed successfully
	c.state = commandParserStateArgBody
	return idx + 2, nil
}

func (c *Command) parseCommandArgBody(data []byte) (int, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		// need more data
		return 0, nil
	}
	param := string(data[:idx])
	if len(param) != c.currentArgSize {
		return 0, fmt.Errorf("command's param body size mismatch: expected %d, got %d", c.currentArgSize, len(param))
	}
	c.Args = append(c.Args, param)
	if len(c.Args) == c.NumberOfArgs {
		// not really ideal
		// TODO: refactor this state machine
		c.Name = strings.ToLower(c.Args[0])
		c.state = commandParserStateDone
	} else {
		// if parsed successfully
		c.state = commandParserStateArgSize
	}
	// c.state = commandParserStateParamSize
	return idx + 2, nil
}
