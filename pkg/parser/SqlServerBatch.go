package parser

import (
	"fmt"
	"os"
)

//
//https://learn.microsoft.com/en-us/sql/t-sql/language-elements/sql-server-utilities-statements-go
//
// SQL Server provides commands that are not Transact-SQL statements,
// but are recognized by the sqlcmd and osql utilities and SQL Server
// Management Studio Code Editor. These commands can be used to
// facilitate the readability and execution of batches and scripts.

// GO signals the end of a batch of Transact-SQL statements to the SQL Server utilities.
// A Transact-SQL statement cannot occupy the same line as a GO command. However, the line can contain comments.

// Opens the specified file and parses its content in
// small chunks of valid utf8 runes.
// Opens the specified file and parses its content in
// small chunks of valid utf8 runes.
func ParseFilepathTsqlBatches(
	filepath string,
	processBatch func(batch []rune, goLine []rune) bool,
) bool {
	p := Parser{}
	//
	var readBuff [64]byte
	goLine := make([]rune, 0, 32)   //Go statement line
	batch := make([]rune, 0, 10240) //TSQL batch
	curLineSz := 0

	//functions
	var processGoLineFunc func(v rune) bool
	var processBatchFunc func(v rune) bool
	var processCurFunc func(v rune) bool

	//functions
	processGoLineFunc = func(v rune) bool {
		const nl = rune('\n')
		//
		goLine = append(goLine, v)
		curLineSz++
		//GO [number][\r]\n
		if (curLineSz == 3 && v != rune(' ') && v != rune('\r') && v != rune('\n')) ||
			(curLineSz > 3 && (v < rune('0') || v > rune('9')) && v != rune('\r') && v != rune('\n')) {
			// The current 'GO...' line was not a GO statement
			// Return its content to the batch
			batch = append(batch, goLine...)
			goLine = goLine[:0]
			processCurFunc = processBatchFunc
			return true
		}
		//
		if v == nl {
			//goLine's end
			if !processBatch(batch, goLine) {
				fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathTsqlBatches'.\n")
				return false
			}
			//reset
			goLine = goLine[:0]
			batch = batch[:0]
			curLineSz = 0
			processCurFunc = processBatchFunc
		}
		return true
	}

	//functions
	processBatchFunc = func(v rune) bool {
		const nl = rune('\n')
		//
		batch = append(batch, v)
		curLineSz++
		//analyze 'GO line'
		if curLineSz == 2 {
			const suffixLen = 2
			len := len(batch)
			r0 := batch[len-2]
			r1 := batch[len-1]
			if (r0 == rune('G') || r0 == rune('g')) &&
				(r1 == rune('O') || r1 == rune('o')) {
				//current line is a 'GO' statement
				//move the 'GO' text from batch-buffer to go-buffer
				goLine = goLine[:0]
				goLine = append(goLine, r0)
				goLine = append(goLine, r1)
				processCurFunc = processGoLineFunc
				batch = batch[:len-2]
			}
		}
		//
		if v == nl {
			curLineSz = 0
		}
		return true
	}

	//functions
	processCurFunc = processBatchFunc
	processString := func(str string) bool {
		const nl = rune('\n')
		//fmt.Fprintf(os.Stderr, "str: \n%s\n\n", str)
		runes := []rune(str)
		for _, v := range runes {
			if !processCurFunc(v) {
				fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathTsqlBatches'.\n")
				return false
			}
		}
		return true
	}
	//
	if !p.ParseUtf16FileIntoStrings(filepath, readBuff[:], processString) {
		return false
	}
	//flush trailing batch
	if len(batch) > 0 && !processBatch(batch, goLine) {
		fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathTsqlBatches'.\n")
		return false
	}
	return true
}
