package parser

import (
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// Opens the specified file and parses its content in
// small chunks of valid strings-tokens.
func (p Parser) ParseUtf8FileIntoStrings(
	filepath string,
	buff []byte, //read buffer
	processString func(str string) bool,
) bool {

	// Validate buffer's capacity
	if cap(buff) < 6 {
		fmt.Fprintf(os.Stderr, "Buffer error, 6 bytes are required to parse UTF8.\n")
		return false
	}

	//open
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s\n")
		return false
	}
	defer file.Close()

	//read buffer
	buffAvailStart := 0
	buffAvail := buff[buffAvailStart:]

	//stats
	bytesTotal := 0

	//read
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

		// Detect utf8 boundary
		bytesValid := bytesRead
		for {
			bytesToEval := buffAvailStart + bytesValid
			// Nothing to evaluate
			if bytesToEval == 0 {
				fmt.Fprintf(os.Stderr, "Buffer or format error. Expected valid UTF8 (no valid-runes after read filtering, buffLen=%d, between[%d and %d bytes]).\n", len(buff), p.bytesCount, p.bytesCount+uint64(bytesRead))
				return false
			}
			// Eval utf8 format
			if utf8.Valid(buff[:bytesToEval]) {
				break
			} else {
				if (buffAvailStart+bytesRead)-(buffAvailStart+bytesValid) > 3 {
					fmt.Fprintf(os.Stderr, "Buffer or format error. Expected valid UTF8 (no more than 3 bytes should left for next utf8-read, buffLen=%d, between[%d and %d bytes]).\n", len(buff), p.bytesCount, p.bytesCount+uint64(bytesRead))
					return false
				}
				bytesValid--
			}
		}

		// None of the available bytes is utf8
		if bytesValid <= 0 {
			fmt.Fprintf(os.Stderr, "Buffer or format error. Expected valid UTF8 (no valid-runes after read).\n")
			return false
		}

		p.bytesCount += uint64(bytesRead)

		// Load valid utf8 string
		bytesAvail := buffAvailStart + bytesRead
		bytesToEval := buffAvailStart + bytesValid
		str := string(buff[:bytesToEval])
		if !processString(str) {
			fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathIntoStrings'.\n")
			return false
		}

		//Stats
		bytesTotal += bytesToEval

		// Update read-slice
		buffAvailStartBefore := buffAvailStart //avoid creating a new slice when posible
		if bytesToEval == bytesAvail {
			// Just reset read-slice
			buffAvailStart = 0
		} else {
			// Move suffix bytes to the start of the array
			buffAvailStart := bytesAvail - bytesToEval
			buffAll := buff[:]
			for i := 0; i < buffAvailStart; i++ {
				buffAll[i] = buffAll[i+bytesToEval]
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
		fmt.Fprintf(os.Stderr, "Trailing data is not valid UTF8.\n")
		return false
	}
	//
	fmt.Fprintf(os.Stderr, "%d bytes total\n", bytesTotal)
	//
	return true
}
