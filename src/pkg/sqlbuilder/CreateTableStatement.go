package sqlbuilder

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ha1tch/tsqlparser"
	"github.com/ha1tch/tsqlparser/ast"
)

func ToMariaDbStr(st2 *tsqlparser.CreateTableStatement) (string, error) {
	strBldr := strings.Builder{}
	strBldr.Grow(10240)
	tableName := st2.Name.Parts[len(st2.Name.Parts)-1].Value
	strBldr.WriteString("CREATE TABLE `" + tableName + "` (\n")
	defsCount := 0
	autoExists := false
	autoField := ""
	autoFieldIsKey := false
	autoSeed := int64(0)
	for _, c := range st2.Columns {
		fieldName := c.Name.Value
		typeStr := c.DataType.Name
		typeSuffix := ""
		// Fix TEXT to LONGTEXT: in SQLServer TEXT can be 2^31 - 1 long, but in MySQL just 64K long.
		if len(typeStr) == 4 && strings.ToUpper(typeStr) == "TEXT" {
			fmt.Fprintf(os.Stderr, "Changing SqlServer(TEXT) field type to MySQL(LONGTEXT) to fit data.\n")
			typeStr = "LONGTEXT"
		}
		// Fix TINYINT to TINYINT UNSIGNED: in SQLServer TINYINT is unsigned, but in MySQL is signed.
		if strings.EqualFold(typeStr, "TINYINT") {
			fmt.Fprintf(os.Stderr, "Changing `%s`.`%s` from %s to TINYINT UNSIGNED.\n", tableName, fieldName, typeStr)
			typeSuffix = "UNSIGNED"
		}
		//
		typeSz := ""
		if c.DataType.Max {
			typeSz = "(MAX)"
		} else if c.DataType.Precision != nil && c.DataType.Scale != nil {
			typeSz = "(" + strconv.Itoa(*c.DataType.Precision) + ", " + strconv.Itoa(*c.DataType.Scale) + ")"
		} else if c.DataType.Length != nil {
			typeSz = "(" + strconv.Itoa(*c.DataType.Length) + ")"
		} else if c.DataType.Precision != nil {
			typeSz = "(" + strconv.Itoa(*c.DataType.Precision) + ")"
		}
		/*identity := ""
		if c.Identity != nil {
			identity = " identity(" + strconv.Itoa(int(c.Identity.Seed)) + ", +" + strconv.Itoa(int(c.Identity.Increment)) + ")"
		}*/
		nullable := " NOT NULL"
		if c.Nullable != nil && *c.Nullable {
			nullable = " NULL"
		}
		//def := ""
		if c.Default != nil {
			//def = " DEFAULT(" + c.Default.String() + ")"
			return "", errors.New("Default in columns definition are currently unsupported.")
		}
		//constrains
		if c.Constraints != nil || len(c.Constraints) > 0 {
			return "", errors.New("Constraints in columns definition are currently unsupported.")
		}
		//column
		if defsCount != 0 {
			strBldr.WriteString(", ")
		}
		defsCount++
		strBldr.WriteString("`" + c.Name.Value + "` " + typeStr + typeSz + " " + typeSuffix + " " + nullable)
		if c.Identity != nil {
			if autoExists {
				return "", errors.New("Only one AUTO_INCREMENT field is allowed per Table.")
			}
			if c.Identity.Increment != 1 {
				return "", errors.New("Identity increment not (+1) in columns definition are currently unsupported.")
			}
			autoExists = true
			autoField = fieldName
			autoSeed = c.Identity.Seed
			strBldr.WriteString(" AUTO_INCREMENT")
		}
		strBldr.WriteString("\n")
	}
	// Constraints
	for _, cc := range st2.Constraints {
		cType := ""
		switch cc.Type {
		case ast.ConstraintPrimaryKey:
			cType = "PRIMARY"
		default:
			//unimplemented
			//ConstraintForeignKey
			//ConstraintUnique:
			//ConstraintCheck
			//ConstraintDefault
			//ConstraintPeriod
			//ConstraintIndex
			return "", errors.New("Unimplemented constraint type: '" + cc.String() + "'.")
		}
		//fmt.Fprintf(os.Stderr, "    %s '%s'.\n", cType, cc.Name)
		if defsCount != 0 {
			strBldr.WriteString(", ")
		}
		defsCount++
		strBldr.WriteString(cType + " KEY (")
		colsCount := 0
		for _, ccc := range cc.Columns {
			fieldName := ccc.Name.Value
			order := "ASC"
			if ccc.Descending {
				order = "DESC"
			}
			//fmt.Fprintf(os.Stderr, "        '%s' %s.\n", ccc.Name, order)
			if colsCount != 0 {
				strBldr.WriteString(", ")
			}
			colsCount++
			strBldr.WriteString("`" + fieldName + "` " + order)
			//
			if autoExists && !autoFieldIsKey &&
				strings.EqualFold(fieldName, autoField) &&
				(cType == "PRIMARY" || cType == "UNIQUE") {
				autoFieldIsKey = true
			}
		}
		strBldr.WriteString(")\n")
	}
	// Fix: unique key for AUTO_INCREMENT
	if autoExists && !autoFieldIsKey {
		if defsCount != 0 {
			strBldr.WriteString(", ")
		}
		defsCount++
		strBldr.WriteString(" UNIQUE `IX_" + tableName + "`(`" + autoField + "`)")
		fmt.Fprintf(os.Stderr, "Injected UNIQUE INDEX at CREATE TABLE because AUTO_INCREMENT is not a PRIMARY KEY: `IX_%s`(`%s`).\n", tableName, autoField)
	}
	strBldr.WriteString(")")
	if autoSeed > 1 {
		strBldr.WriteString(" AUTO_INCREMENT = " + strconv.FormatInt(autoSeed, 10))
	}
	strBldr.WriteString(";\n")
	//
	return strBldr.String(), nil
}
