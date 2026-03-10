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
		response = cmdSADD(cmd)
	case "SREM":
		response = cmdSREM(cmd)
	case "SMEMBERS":
		response = cmdSMEMBERS(cmd)
	case "SISMEMBER":
		response = cmdSISMEMBER(cmd)
	case "CMS.INITBYPROB":
		response = cmdCMSINITBYPROB(cmd)
	case "CMS.INCRBY":
		response = cmdCMSINCRBY(cmd)
	case "CMS.INFO":
		response = cmdCMSINFO(cmd)
	case "CMS.QUERY":
		response = cmdCMSQUERY(cmd)
	default:
		response = []byte("-CMD NOT FOUND\r\n")
	}
	_, err := syscall.Write(fd, response)
	return err
}
