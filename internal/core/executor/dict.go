package executor

import (
	"errors"
	"strconv"
	"strings"
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
	return resp_parser.Encode(errors.New("ERR wrong number of arguments for command"), false)
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
	if len(cmd.Args) != 1 {
		return resp_parser.Encode(errors.New("ERR wrong number of arguments for command"), false)
	}
	obj := core.Dictionary.Get(cmd.Args[0])
	if obj == nil {
		return []byte("$-1\r\n")
	}
	return resp_parser.Encode(obj.Value, false)
}

func cmdTTL(cmd *command.Command) []byte {
	if len(cmd.Args) != 1 {
		return resp_parser.Encode(errors.New("ERR wrong number of arguments for command"), false)
	}
	obj := core.Dictionary.Get(cmd.Args[0])
	if obj == nil {
		return resp_parser.Encode(-2, true)
	}
	expireTime, expExist := core.Dictionary.GetExpiry(cmd.Args[0])
	if !expExist {
		return resp_parser.Encode(-1, true)
	}
	nowMs := time.Now().UnixMilli()
	ttlMs := expireTime - nowMs
	ttlSec := ttlMs / 1000
	return resp_parser.Encode(ttlSec, true)
}

func cmdExpire(cmd *command.Command) []byte {
	if len(cmd.Args) != 2 {
		return resp_parser.Encode(errors.New("ERR wrong number of arguments for command"), false)
	}
	obj := core.Dictionary.Get(cmd.Args[0])
	if obj == nil {
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
		obj := core.Dictionary.Get(key)
		if obj != nil {
			core.Dictionary.Del(key)
			deleted++
		}
	}
	return resp_parser.Encode(deleted, true)
}
