package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ha1tch/tsqlparser"
	"github.com/marcosjom/db_migrate/pkg/parser"
	"github.com/marcosjom/db_migrate/pkg/sqlbuilder"
)

type runConfig struct {
	InputFile  string //tsql source file
	OutputFile string //sql destination file (stdout as default)
}

func main() {

	//
	cfg := runConfig{}

	//arguments
	args := os.Args
	argsLen := len(args)

	//program's location
	/*if argsLen > 0 {
		fmt.Fprintf(os.Stderr, "Progam '%s'\n", args[0])
	}*/

	//parse arguments
	iArg := 1
	iArgLastRun := iArg
	for iArg < argsLen {
		v := args[iArg]
		switch v {
		case "-i", "-inputFile":
			if iArg+1 >= argsLen {
				fmt.Fprintf(os.Stderr, "Missing value for argument '%s'\n", v)
				os.Exit(-1)
			}
			iArg++
			cfg.InputFile = args[iArg]
		case "-o", "-outputFile":
			if iArg+1 >= argsLen {
				fmt.Fprintf(os.Stderr, "Missing value for argument '%s'\n", v)
				os.Exit(-1)
			}
			iArg++
			cfg.OutputFile = args[iArg]
		case "-r", "-run":
			//Execute with current params-state
			iArgLastRun = iArg + 1
			if !run(&cfg) {
				fmt.Fprintf(os.Stderr, "Run failed.\n")
				os.Exit(-1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown argument '%s'\n", v)
			os.Exit(-1)
		}
		iArg++
	}

	//flush config changes
	if iArgLastRun != iArg {
		iArgLastRun = iArg
		if !run(&cfg) {
			fmt.Fprintf(os.Stderr, "Run failed.\n")
			os.Exit(-1)
		}
	}

	//Help
	if iArg == 1 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "This program parses a MS SQL Server tsql batches file and outputs its equivalent for Mysql/MariaDb.\n")
		fmt.Fprintf(os.Stderr, "Useful for rapid database migration.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "tsql2sql [args]\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "-i | -inputFile file , sets the current input file.\n")
		fmt.Fprintf(os.Stderr, "-o | -ouputFile file , sets the current output file (def: stdout).\n")
		fmt.Fprintf(os.Stderr, "-r | -run            , runs the parser with the current configuration state (optional for last cfg changes).\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "tsql2sql -i myFile.tsql\n")
		fmt.Fprintf(os.Stderr, "tsql2sql -i myFile.tsql -o myFile.sql -run -i myFile2.tsql -o myFile2.sql -run\n")
	}

	//

	//parser.ParseFilepathTsqlBatches("D:\\NIBSA_Proyectos\\CltTecnolite\\sys-caproy\\db-backups\\2026-02-09-CAPROY.full.sql", processBatch)
	//parser.ParseFilepathTsqlBatches("D:\\NIBSA_Proyectos\\CltTecnolite\\sys-caproy\\db-backups\\2026-02-09-CAWEB_Sincronizaciones.full.sql", processBatch)
	//parser.ParseFilepathTsqlBatches("D:\\NIBSA_Proyectos\\CltTecnolite\\sys-caproy\\db-backups\\2026-02-09-CAPROY_Adjuntos2.full.sql", processBatch)

	/*batchStsGrpStr := "USE [CAPROY]\r\n"
	program, errors := tsqlparser.Parse(batchStsGrpStr)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		}
	}
	fmt.Fprintf(os.Stderr, "Program (%d statements):\n%s\n", len(program.Statements), program.String())*/
	//
	//fmt.Fprintf(os.Stderr, "Done!\n")
}

func run(cfg *runConfig) bool {
	processBatch := func(batch []rune, goLine []rune) bool {
		const startComment = "/*"
		const endComment = "*/"
		//Batch statements
		batchStr := string(batch)
		//fmt.Fprintf(os.Stderr, "Batch %d bytes: '%s'\n", len(batchStr), batchStr)
		//TMP Patches
		//Issue: https://github.com/ha1tch/tsqlparser/issues/1
		//From: CREATE TABLE (... CONSTRAINT ... ( ... ) ON [PRIMARY] ) ON [PRIMARY]
		//To:   CREATE TABLE (... CONSTRAINT ... ( ... ) ) ON [PRIMARY]
		if strings.Contains(batchStr, ") ON [PRIMARY]\r\n) ON [PRIMARY]") {
			batchStr = strings.ReplaceAll(batchStr, ") ON [PRIMARY]\r\n) ON [PRIMARY]", ")) ON [PRIMARY]")
		} else if strings.Contains(batchStr, "ALTER TABLE") && strings.Contains(batchStr, ") ON [PRIMARY]") {
			batchStr = strings.ReplaceAll(batchStr, ") ON [PRIMARY]", ")")
		}
		//Issue: https://github.com/ha1tch/tsqlparser/issues/2
		//From: ALTER TABLE [dbo].[myTable] ADD  DEFAULT ('') FOR [myField]
		//To:   ALTER TABLE [dbo].[myTable] ADD  CONSTRAINT [DF_myTable_myField]  DEFAULT (getdate()) FOR [myField]
		if strings.Contains(batchStr, "ALTER TABLE [dbo].[") &&
			strings.Contains(batchStr, "] ADD  DEFAULT (") &&
			strings.Contains(batchStr, ") FOR [") {
			len0 := len("ALTER TABLE [dbo].[")
			len1 := len("] ADD  DEFAULT (")
			len2 := len(") FOR [")
			pos0 := strings.Index(batchStr, "ALTER TABLE [dbo].[")
			pos1 := pos0 + len0 + strings.Index(batchStr[pos0+len0:], "] ADD  DEFAULT (")
			pos2 := pos1 + len1 + strings.Index(batchStr[pos1+len1:], ") FOR [")
			pos3 := pos2 + len2 + strings.Index(batchStr[pos2+len2:], "]")
			tblName := batchStr[pos0+len0 : pos1]
			fldName := batchStr[pos2+len2 : pos3]
			//
			batchStr = batchStr[:pos1] + "] ADD  CONSTRAINT [DF_" + tblName + "_" + fldName + "]  DEFAULT (" + batchStr[pos1+len1:]
			//fmt.Fprintf(os.Stderr, "Fix: Table '%s' Field '%s'.\n", tblName, fldName)
			//fmt.Fprintf(os.Stderr, "Fix: '%s'.\n", batchStr)
		}
		//Ignore "CREATE PROCEDURE" (some fail)
		if strings.Contains(batchStr, "CREATE PROCEDURE") {
			fmt.Fprintf(os.Stderr, "Ignoring 'CREATE PROCEDURE'.\n")
			fmt.Println(startComment + " Create Procedure (IGNORED)" + endComment)
			return true
		}
		//Ignore "CREATE FUNCTION" (some fail)
		if strings.Contains(batchStr, "CREATE FUNCTION") {
			fmt.Fprintf(os.Stderr, "Ignoring 'CREATE FUNCTION'.\n")
			fmt.Println(startComment + " Create Function (IGNORED)" + endComment)
			return true
		}
		//Ignore "EXEC sys."
		if strings.Contains(batchStr, "EXEC sys.") {
			fmt.Fprintf(os.Stderr, "Ignoring 'EXEC sys.' (Diagrams).\n")
			fmt.Println(startComment + " 'EXEC sys.' (Diagram?) (IGNORED)" + endComment)
			return true
		}
		//parse
		program, errors := tsqlparser.Parse(batchStr)
		if len(errors) > 0 {
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
			}
			fmt.Fprintf(os.Stderr, "Failed batch: ---->\n%s\n<----\n", batchStr)
			return false
		}
		if program == nil {
			return false
		}
		if len(program.Statements) <= 0 {
			fmt.Fprintf(os.Stderr, "No statements parsed for batch: ---->\n%s\n<----\n", batchStr)
			return false
		}
		//statements
		for _, st := range program.Statements {
			switch st2 := st.(type) {
			case *tsqlparser.UseStatement:
				sql, error := sqlbuilder.ToMariaDbStr_UseStatement(st2)
				if error != nil {
					fmt.Fprintf(os.Stderr, "\n%s\n", st2.String())
					fmt.Fprintf(os.Stderr, "ERROR UseStatement, : %s.\n", error.Error())
					return false
				}
				fmt.Println(sql)
			case *tsqlparser.CreateTableStatement:
				sql, error := sqlbuilder.ToMariaDbStr(st2)
				if error != nil {
					fmt.Fprintf(os.Stderr, "\n%s\n", st2.String())
					fmt.Fprintf(os.Stderr, "ERROR CreateTableStatement, : %s.\n", error.Error())
					return false
				}
				fmt.Println(sql)
			case *tsqlparser.AlterTableStatement:
				sql, error := sqlbuilder.ToMariaDbStr_AlterTableStatement(st2)
				if error != nil {
					fmt.Fprintf(os.Stderr, "\n%s\n", st2.String())
					fmt.Fprintf(os.Stderr, "ERROR AlterTableStatement, : %s.\n", error.Error())
					return false
				}
				fmt.Println(sql)
			case *tsqlparser.CreateIndexStatement:
				sql, error := sqlbuilder.ToMariaDbStr_CreateIndexStatement(st2)
				if error != nil {
					fmt.Fprintf(os.Stderr, "\n%s\n", st2.String())
					fmt.Fprintf(os.Stderr, "ERROR CreateIndexStatement, : %s.\n", error.Error())
					return false
				}
				fmt.Println(sql)
			case *tsqlparser.InsertStatement:
				sql, error := sqlbuilder.ToMariaDbStr_InsertStatement(st2)
				if error != nil {
					fmt.Fprintf(os.Stderr, "\n%s\n", st2.String())
					fmt.Fprintf(os.Stderr, "ERROR InsertStatement, : %s.\n", error.Error())
					return false
				}
				fmt.Println(sql)
			case *tsqlparser.SetStatement:
				//ignoring
				lineStr := st2.String()
				if lineStr != "" {
					if lineStr[len(lineStr)-1:] == "\n" {
						lineStr = lineStr[:len(lineStr)-1]
					}
					if lineStr[len(lineStr)-1:] == "\r" {
						lineStr = lineStr[:len(lineStr)-1]
					}
					fmt.Println(startComment + lineStr + endComment)
				}
			case *tsqlparser.CreateFunctionStatement:
				fmt.Fprintf(os.Stderr, "Ignoring Create Function '%s'.\n", st2.Name.Parts[len(st2.Name.Parts)-1])
				fmt.Println(startComment + " Create Function '" + st2.Name.Parts[len(st2.Name.Parts)-1].Value + "' (IGNORED)" + endComment)
			case *tsqlparser.CreateProcedureStatement:
				fmt.Fprintf(os.Stderr, "Ignoring Create Procedure '%s'.\n", st2.Name.Parts[len(st2.Name.Parts)-1])
				fmt.Println(startComment + "Create Procedure '" + st2.Name.Parts[len(st2.Name.Parts)-1].Value + "' (IGNORED)" + endComment)
			case *tsqlparser.SetOptionStatement:
				switch st2.Option {
				case "IDENTITY_INSERT":
					//fmt.Fprintf(os.Stderr, "Ignoring '%s', '%s' = '%s'\n", st2.Table.Parts[len(st2.Table.Parts)-1], st2.Option, st2.Value)
				default:
					fmt.Fprintf(os.Stderr, "Unsupported SetOptionStatement: %s\n", st.String())
					return false
				}
				lineStr := st2.String()
				if lineStr != "" {
					if lineStr[len(lineStr)-1:] == "\n" {
						lineStr = lineStr[:len(lineStr)-1]
					}
					if lineStr[len(lineStr)-1:] == "\r" {
						lineStr = lineStr[:len(lineStr)-1]
					}
					fmt.Println(startComment + lineStr + endComment)
				}
			case *tsqlparser.CreateViewStatement:
				fmt.Fprintf(os.Stderr, "Ignoring Create View '%s'.\n", st2.Name.Parts[len(st2.Name.Parts)-1])
				//ignoring
			default:
				fmt.Fprintf(os.Stderr, "Unsupported statement: %s\n", st.String())
				return false
			}
		}
		//Go statement
		goLineStr := string(goLine)
		if goLineStr != "" {
			if goLineStr[len(goLineStr)-1:] == "\n" {
				goLineStr = goLineStr[:len(goLineStr)-1]
			}
			if goLineStr[len(goLineStr)-1:] == "\r" {
				goLineStr = goLineStr[:len(goLineStr)-1]
			}
			goLineStrU := strings.ToUpper(goLineStr)
			if goLineStrU != "GO" {
				fmt.Fprintf(os.Stderr, "Go: \n%s\n", goLineStr)
			}
			fmt.Println(startComment + goLineStr + endComment)
		}
		//fmt.Fprintf(os.Stderr, "Program: %d bytes\n", len(program.String()))
		//
		return true
	}
	//action
	return parser.ParseFilepathTsqlBatches(cfg.InputFile, processBatch)
}
