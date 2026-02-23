package sqlbuilder

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ha1tch/tsqlparser"
)

func ToMariaDbStr_CreateIndexStatement(st2 *tsqlparser.CreateIndexStatement) (string, error) {
	strBldr := strings.Builder{}
	strBldr.Grow(10240)
	strBldr.WriteString("CREATE")
	if st2.IsUnique {
		strBldr.WriteString(" UNIQUE")
	}
	strBldr.WriteString(" INDEX")
	//MariaDb sintax
	{
		strBldr.WriteString(" IF NOT EXISTS")
	}
	if len(st2.Name.Value) > 0 {
		strBldr.WriteString(" `" + st2.Name.Value + "`")
	}
	strBldr.WriteString(" ON")
	if len(st2.Table.Parts) <= 0 {
		return "", errors.New("Missing table name in index definition.")
	}
	strBldr.WriteString(" `" + st2.Table.Parts[len(st2.Table.Parts)-1].Value + "`")
	if len(st2.Columns) <= 0 {
		return "", errors.New("Missing columns in index definition.")
	}
	colsCount := 0
	strBldr.WriteString("(")
	for _, col := range st2.Columns {
		if colsCount != 0 {
			strBldr.WriteString(", ")
		}
		colsCount++
		strBldr.WriteString("`" + col.Name.Value + "`")
		if col.Descending {
			strBldr.WriteString(" DESC")
		}
	}
	strBldr.WriteString(")")
	if st2.IsClustered != nil && *st2.IsClustered {
		strBldr.WriteString(" CLUSTERING=YES")
	}
	strBldr.WriteString(";\n")
	//
	fmt.Fprintf(os.Stderr, "\n%s\n", strBldr.String())
	//
	return strBldr.String(), nil
}
