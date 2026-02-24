package parser

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// From utf16.go
const (
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000
)

// Opens the specified file and parses its content in
// small chunks of valid strings-tokens.
func (p Parser) ParseUtf16FileIntoStrings(
	filepath string,
	buff []byte, //read buffer
	processString func(str string) bool,
) bool {

	// Validate buffer's capacity
	if cap(buff) < 4 {
		fmt.Fprintf(os.Stderr, "Buffer error, 4 bytes are required to parse UTF16.\n")
		return false
	}

	//open
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s.\n", err)
		return false
	}
	defer file.Close()

	// read config
	isBOMPossible := true
	endienPositions := [...]int{0, 1, 2, 3}

	// read buffer
	buffAvailStart := 0
	buffAvail := buff[buffAvailStart:]
	strBldr := strings.Builder{}
	strBldr.Grow((cap(buff) / 2) + 1)

	// stats
	bytesTotal := 0

	// read
	for {

		// Validate buffer's space
		if len(buffAvail) == 0 {
			fmt.Fprintf(os.Stderr, "Buffer or format error. Expected valid UTF8 (buffer's full).\n")
			return false
		}

		// Read a chunk
		bytesRead, err := file.Read(buffAvail)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading file: %s.\n", err)
			return false
		}

		// Stop loop at EOF
		if bytesRead == 0 {
			break
		}

		i := 0
		iAfterEnd := (buffAvailStart + bytesRead) / 2 * 2

		//Analyze BOM (byte-order mark)
		if isBOMPossible && (buffAvailStart+bytesRead) > 1 {
			if buff[i] == 0xFF && buff[i+1] == 0xFE {
				//cpu-inverted endien
				fmt.Fprintf(os.Stderr, "cpu-inverted endien BOM (byte-order mark) detected.\n")
				endienPositions[0] = 1
				endienPositions[1] = 0
				endienPositions[2] = 3
				endienPositions[3] = 2
				i += 2
			} else if buff[i+1] == 0xFF && buff[i] == 0xFE {
				//cpu-compatible endien
				fmt.Fprintf(os.Stderr, "cpu-compatible endien BOM (byte-order mark) detected.\n")
				i += 2
			}
			// stop analyzing BOM
			isBOMPossible = false
		}

		// Consume buffer
		var u0 rune
		var u1 rune
		for i < iAfterEnd {
			// read two bytes
			u0 = (rune(buff[i+endienPositions[0]]) << 8) + rune(buff[i+endienPositions[1]])
			// analyze normal rune
			if u0 < surr1 || surr3 <= u0 {
				strBldr.WriteRune(u0)
				i += 2
				continue
			}
			// analyze  pair
			if surr1 <= u0 && u0 < surr2 {
				if i+4 >= iAfterEnd {
					//more bytes are required
					break
				}
				// read another two bytes
				u1 = (rune(buff[i+endienPositions[2]]) << 8) + rune(buff[i+endienPositions[3]])
				if surr2 <= u1 && u1 < surr3 {
					//valid pair
					u01 := (u0-surr1)<<10 | (u1 - surr2) + surrSelf
					strBldr.WriteRune(u01)
					i += 4
					continue
				}
				fmt.Fprintf(os.Stderr, "Buffer or format error. Expected valid UTF16 bytes.")
				return false
			}
			//
			break
		}

		// Validate unconsumed bytes
		if i+1 < iAfterEnd {
			fmt.Fprintf(os.Stderr, "Buffer or format error. Expected valid UTF16 (no more than 1 byte should left for next utf16-read, buffLen=%d, between[%d and %d bytes]).\n", len(buff), p.bytesCount, p.bytesCount+uint64(bytesRead))
			return false
		}

		// Load valid utf8 string
		str := strBldr.String()
		if !processString(str) {
			fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathIntoStrings'.\n")
			return false
		}

		//
		strBldr.Reset()
		p.bytesCount += uint64(bytesRead)

		//Stats
		bytesTotal += bytesRead

		// Update read-slice
		buffAvailStartBefore := buffAvailStart //avoid creating a new slice when posible
		if i == iAfterEnd {
			// Just reset read-slice
			buffAvailStart = 0
		} else {
			// Move suffix bytes to the start of the array
			buffAvailStart = iAfterEnd - i
			buffAll := buff[:]
			for i2 := 0; i2 < buffAvailStart; i2++ {
				buffAll[i2] = buffAll[i2+i]
			}
			fmt.Fprintf(os.Stderr, "Left %d bytes for next read\n", buffAvailStart)
		}
		//avoid creating a new slice when posible
		if buffAvailStartBefore != buffAvailStart {
			buffAvail = buff[buffAvailStart:]
		}
	}
	//
	if buffAvailStart > 0 {
		fmt.Fprintf(os.Stderr, "Trailing data is not valid UTF16.\n")
		return false
	}
	//
	fmt.Fprintf(os.Stderr, "%d bytes total\n", bytesTotal)
	//
	return true
}
