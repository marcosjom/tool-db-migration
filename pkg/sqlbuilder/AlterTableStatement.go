package sqlbuilder

import (
	"errors"
	"strings"

	"github.com/ha1tch/tsqlparser"
	"github.com/ha1tch/tsqlparser/ast"
)

func ToMariaDbStr_AlterTableStatement(st2 *tsqlparser.AlterTableStatement) (string, error) {
	strBldr := strings.Builder{}
	strBldr.Grow(10240)
	// Actions
	for _, act := range st2.Actions {
		switch act.Type {
		case ast.AlterAddConstraint:
			cc := act.Constraint
			if cc == nil {
				return "", errors.New("Missing constraint definition in action : " + act.String() + " in " + st2.String() + ".")
			}
			//cType := ""
			switch cc.Type {
			case ast.ConstraintUnique:
				//cType = "UNIQUE"
				strBldr.WriteString("CREATE UNIQUE INDEX")
				//ToDo: implement as optional
				//MariaDb sintax
				{
					strBldr.WriteString(" IF NOT EXISTS")
				}
				if cc.Name != "" {
					strBldr.WriteString(" `" + cc.Name + "`")
				}
				strBldr.WriteString(" ON `" + st2.Table.Parts[len(st2.Table.Parts)-1].Value + "`(")
				if len(cc.Columns) <= 0 {
					return "", errors.New("Missing columns in unique constraint definition.")
				}
				colsCount := 0
				for _, col := range cc.Columns {
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
				//ToDo: implement as optional
				//MySql sintax
				/*{
					strBldr.WriteString(" IF NOT EXISTS")
				}*/
				//
				strBldr.WriteString(";\n")

			case ast.ConstraintDefault:
				//cType = "DEFAULT"
				if cc.ForColumn == nil {
					return "", errors.New("Missing ForColumn on DEFAULT Constraint.")
				}
				defVal := cc.DefaultExpression.String()
				if defVal == "getdate()" {
					defVal = "CURRENT_TIMESTAMP"
				}
				strBldr.WriteString("ALTER TABLE `" + st2.Table.Parts[len(st2.Table.Parts)-1].Value + "` ALTER COLUMN `" + cc.ForColumn.Value + "` SET DEFAULT " + defVal + ";\n")
			case ast.ConstraintForeignKey:
				//cType = "FOREIGN"
				strBldr.WriteString("ALTER TABLE `" + st2.Table.Parts[len(st2.Table.Parts)-1].Value + "` ADD CONSTRAINT")
				if cc.Name != "" {
					strBldr.WriteString(" `" + cc.Name + "`")
				}
				strBldr.WriteString(" FOREIGN KEY(")
				if len(cc.Columns) <= 0 {
					return "", errors.New("Missing columns in unique constraint definition.")
				}
				colsCount := 0
				for _, col := range cc.Columns {
					if colsCount != 0 {
						strBldr.WriteString(", ")
					}
					colsCount++
					strBldr.WriteString("`" + col.Name.Value + "`")
				}
				strBldr.WriteString(") REFERENCES ")
				if cc.ReferencesTable == nil || len(cc.ReferencesTable.Parts) <= 0 {
					return "", errors.New("Missing referenced table name in foreign constraint definition.")
				}
				strBldr.WriteString("`" + cc.ReferencesTable.Parts[len(cc.ReferencesTable.Parts)-1].Value + "`(")
				colsCount = 0
				for _, col := range cc.ReferencesColumns {
					if colsCount != 0 {
						strBldr.WriteString(", ")
					}

					colsCount++
					strBldr.WriteString("`" + col.Value + "`")
				}
				strBldr.WriteString(");")
			default:
				//unimplemented
				//ConstraintPrimaryKey
				//ConstraintForeignKey
				//ConstraintCheck
				//ConstraintDefault
				//ConstraintPeriod
				//ConstraintIndex
				return "", errors.New("Unimplemented constraint type: '" + cc.String() + "'.")
			}
		case ast.AlterCheckConstraint:
			//
		default:
			//AlterAddColumn = iota
			//AlterDropColumn
			//AlterAlterColumn
			//AlterDropConstraint
			//AlterRenameColumn
			//AlterEnableTrigger
			//AlterDisableTrigger
			//AlterSetOption
			//AlterCheckConstraint   // CHECK CONSTRAINT name
			//AlterNoCheckConstraint // NOCHECK CONSTRAINT name
			//AlterSwitch            // SWITCH [PARTITION n] TO target [PARTITION n]
			//AlterRebuild           // REBUILD
			return "", errors.New("Unimplemented Alter-table action: " + act.String() + " in " + st2.String() + ".")
		}
	}
	//
	return strBldr.String(), nil
}
