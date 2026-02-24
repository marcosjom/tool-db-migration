package sqlbuilder

import (
	"github.com/ha1tch/tsqlparser"
)

func ToMariaDbStr_UseStatement(st2 *tsqlparser.UseStatement) (string, error) {
	return "USE " + st2.Database.Value + ";", nil
}
