package executor

import (
	"syscall"

	"github.com/quockhanhcao/redish/internal/core/command"
)

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
	case "SADD":
		response = cmdSAdd(cmd)
	case "SREM":
		response = cmdSRem(cmd)
	case "SMEMBERS":
		response = cmdSMembers(cmd)
	case "SISMEMBER":
		response = cmdSIsmember(cmd)
	case "CMS.INITBYPROB":
		response = cmdInitCMS(cmd)
	case "CMS.INCR":
		response = cmdIncrCMS(cmd)
	default:
		response = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(fd, response)
	return err
}
