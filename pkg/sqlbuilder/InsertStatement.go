package sqlbuilder

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ha1tch/tsqlparser"
)

func ToMariaDbStr_InsertStatement(st2 *tsqlparser.InsertStatement) (string, error) {
	strBldr := strings.Builder{}
	strBldr.Grow(10240)
	strBldr.WriteString("INSERT INTO `" + st2.Table.Parts[len(st2.Table.Parts)-1].Value + "` (")
	colsCount := 0
	for _, col := range st2.Columns {
		if colsCount != 0 {
			strBldr.WriteString(", ")
		}
		colsCount++
		strBldr.WriteString(col.Value)
	}
	strBldr.WriteString(") VALUES")
	rowsCount := 0
	for _, row := range st2.Values {
		if len(row) != colsCount {
			return "", errors.New("Insert columns(" + strconv.Itoa(colsCount) + ") and values(" + strconv.Itoa(len(row)) + ") count misstach.")
		}
		if rowsCount != 0 {
			strBldr.WriteString("\n")
		}
		rowsCount++
		strBldr.WriteString("(")
		valCount := 0
		for _, value := range row {
			if valCount != 0 {
				strBldr.WriteString(", ")
			}
			valCount++
			valueStr := value.String()
			// String ''
			// String N''
			// In T-SQL, the N prefix is used to denote a Unicode string.
			if len(valueStr) > 2 &&
				(valueStr[:1] == "'" || valueStr[:2] == "N'") &&
				valueStr[len(valueStr)-1:] == "'" {
				strOnly := ""
				if valueStr[:2] == "N'" {
					strOnly = valueStr[2 : len(valueStr)-1]
				} else {
					strOnly = valueStr[1 : len(valueStr)-1]
				}
				//Testing external library behavior
				/*
					if strings.Contains(strOnly, "''") {
						return "", errors.New("Found scape combination ('') inside string.")
					}
					if strings.Contains(strOnly, "'") {
						return "", errors.New("Found unscaped char (') inside string.")
					}
					if strings.Contains(strOnly, "\\") {
						return "", errors.New("Found unscaped char (\\) inside string.")
					}
					//It seems the library unscapes the values. ("''" -> "'", "\\" -> "\")
				*/
				//Data Issue, found  DEL (0x7F) character in 4 INSERT statements.
				{
					const delStr = string(rune(0x7F))
					if strings.Contains(strOnly, delStr) {
						strOnly = strings.ReplaceAll(strOnly, delStr, "_")
						fmt.Fprintf(os.Stderr, "\n%s\n", st2.String())
						fmt.Fprintf(os.Stderr, "Found DEL (0x7F) char inside string, using \"_\" as a replacement.\n")
					}
				}
				//scape
				strOnly = strings.ReplaceAll(strOnly, "'", "''")
				strOnly = strings.ReplaceAll(strOnly, "\\", "\\\\")
				//
				valueStr = "'" + strOnly + "'"
			}
			// 0         1         2         3         4
			// 01234567890123456789012345678901234567890123
			// CAST(N'YYYY-MM-DDThh:mm:ss.mmm' AS DATETIME)
			if len(valueStr) == 44 &&
				strings.EqualFold(valueStr[:len("CAST(N'")], "CAST(N'") &&
				strings.EqualFold(valueStr[17:18], "T") &&
				strings.EqualFold(valueStr[30:], "' AS DATETIME)") {
				valueStr = "'" + valueStr[7:17] + " " + valueStr[18:30] + "'"
			}
			//
			strBldr.WriteString(valueStr)

		}
		strBldr.WriteString(")")
	}
	strBldr.WriteString(";\n")
	//
	return strBldr.String(), nil
}
