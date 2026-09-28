package executor

import (
	"bytes"
	"fmt"

	"github.com/quockhanhcao/redish/internal/core/resp_parser"
	"github.com/quockhanhcao/redish/internal/data_structure"
)

func cmdINFO() []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	buf.WriteString(fmt.Sprintf("db0:keys=%d,expire=0\r\n", data_structure.Stats.Key))
	return resp_parser.Encode(buf.String(), false)
}
