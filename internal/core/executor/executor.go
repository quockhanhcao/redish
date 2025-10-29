package executor

import (
	"errors"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/quockhanhcao/redish/internal/core"
	"github.com/quockhanhcao/redish/internal/core/command"
	"github.com/quockhanhcao/redish/internal/core/resp_parser"
)

func cmdPing(cmd *command.Command) []byte {

	if len(cmd.Args) == 0 {
		return resp_parser.Encode("PONG", true)
	}
	if len(cmd.Args) == 1 {
		return resp_parser.Encode(cmd.Args[0], false)
	}
	return resp_parser.Encode(errors.New("wrong number of arguments for 'ping' command"), false)
}

// set abc ex 5
func cmdSet(cmd *command.Command) []byte {
	if len(cmd.Args) < 2 || len(cmd.Args) == 3 || len(cmd.Args) > 4 {
		return resp_parser.Encode(errors.New("syntax error"), false)
	}
	if len(cmd.Args) == 2 {
		core.Dictionary.Set(cmd.Args[0], cmd.Args[1], -1)
	} else {
		expTime, err := strconv.Atoi(cmd.Args[3])
		if err != nil || expTime <= 0 || strings.ToUpper(cmd.Args[2]) != "EX" {
			return resp_parser.Encode(errors.New("syntax error"), false)
		}
		core.Dictionary.Set(cmd.Args[0], cmd.Args[1], int64(expTime))
	}
	return []byte("+OK\r\n")
}

func cmdGet(cmd *command.Command) []byte {
	if len(cmd.Args) > 1 {
		return resp_parser.Encode(errors.New("ERR wrong number of arguments for command"), false)
	}
	val, ok := core.Dictionary.Get(cmd.Args[0])
	if !ok {
		return []byte("$-1\r\n")
	}
	return resp_parser.Encode(val, false)
}

func cmdTTL(cmd *command.Command) []byte {
	expireTime, expExist := core.Dictionary.GetExpiry(cmd.Args[0])
	_, keyExist := core.Dictionary.Get(cmd.Args[0])
	if !expExist {
		if keyExist {
			return resp_parser.Encode(-1, true)
		}
		return resp_parser.Encode(-2, true)
	}
	nowMs := time.Now().UnixMilli()
	ttlMs := expireTime - nowMs
	ttlSec := int64(ttlMs / 1000)
	if ttlSec < 0 {
		return resp_parser.Encode(-2, true)
	}
	return resp_parser.Encode(ttlSec, true)
}

func cmdExpire(cmd *command.Command) []byte {
	if len(cmd.Args) != 2 {
		return resp_parser.Encode(errors.New("ERR wrong number of arguments for command"), false)
	}
	_, ok := core.Dictionary.Get(cmd.Args[0])
	if !ok {
		return resp_parser.Encode(0, true)
	}
	expTime, err := strconv.Atoi(cmd.Args[1])
	if err != nil || expTime <= 0 {
		return resp_parser.Encode(errors.New("ERR invalid expire time"), false)
	}
	core.Dictionary.SetExpiry(cmd.Args[0], int64(expTime))
	return resp_parser.Encode(1, true)
}

func cmdDel(cmd *command.Command) []byte {
	deleted := 0
	for _, key := range cmd.Args {
		_, ok := core.Dictionary.Get(key)
		if ok {
			core.Dictionary.Del(key)
			deleted++
		}
	}
	return resp_parser.Encode(deleted, true)
}

func ExecuteCommand(cmd *command.Command, fd int) error {
	var response []byte
	switch cmd.Cmd {
	case "PING":
		response = cmdPing(cmd)
	case "SET":
		response = cmdSet(cmd)
	case "GET":
		response = cmdGet(cmd)
	case "TTL":
		response = cmdTTL(cmd)
	case "EXPIRE":
		response = cmdExpire(cmd)
	case "DEL":
		response = cmdDel(cmd)
	default:
		response = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(fd, response)
	return err
}
