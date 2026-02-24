package parser

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// https://en.wikipedia.org/wiki/ANSI_character_set
// https://en.wikipedia.org/wiki/Windows-1252

var Ansi1252Runes = [256]rune{
	//ASCII
	rune(0), rune(1), rune(2), rune(3), rune(4), rune(5), rune(6), rune(7),
	rune(8), rune(9), rune(10), rune(11), rune(12), rune(13), rune(14), rune(15),
	rune(16), rune(17), rune(18), rune(19), rune(20), rune(21), rune(22), rune(23),
	rune(24), rune(25), rune(26), rune(27), rune(28), rune(29), rune(30), rune(31),
	rune(32), rune(33), rune(34), rune(35), rune(36), rune(37), rune(38), rune(39),
	rune(40), rune(41), rune(42), rune(43), rune(44), rune(45), rune(46), rune(47),
	rune(48), rune(49), rune(50), rune(51), rune(52), rune(53), rune(54), rune(55),
	rune(56), rune(57), rune(58), rune(59), rune(60), rune(61), rune(62), rune(63),
	rune(64), rune(65), rune(66), rune(67), rune(68), rune(69), rune(70), rune(71),
	rune(72), rune(73), rune(74), rune(75), rune(76), rune(77), rune(78), rune(79),
	rune(80), rune(81), rune(82), rune(83), rune(84), rune(85), rune(86), rune(87),
	rune(88), rune(89), rune(90), rune(91), rune(92), rune(93), rune(94), rune(95),
	rune(96), rune(97), rune(98), rune(99), rune(100), rune(101), rune(102), rune(103),
	rune(104), rune(105), rune(106), rune(107), rune(108), rune(109), rune(110), rune(111),
	rune(112), rune(113), rune(114), rune(115), rune(116), rune(117), rune(118), rune(119),
	rune(120), rune(121), rune(122), rune(123), rune(124), rune(125), rune(126), rune(127),
	//Windows-1252
	rune('€') /*20AC*/, rune(0), rune('‚') /*201A*/, rune('ƒ') /*0192*/, rune('„') /*201E*/, rune('…') /*2026*/, rune('†') /*2020*/, rune('‡'), /*2021*/
	rune('ˆ') /*02C6*/, rune('‰') /*2030*/, rune('Š') /*0160*/, rune('‹') /*2039*/, rune('Œ') /*0152*/, rune(0), rune('Ž') /*017D*/, rune(0),
	//
	rune(0), rune('‘') /*2018*/, rune('’') /*2019*/, rune('“') /*201C*/, rune('”') /*201D*/, rune('•') /*2022*/, rune('–') /*2013*/, rune('—'), /*2014*/
	rune('˜') /*02DC*/, rune('™') /*2122*/, rune('š') /*0161*/, rune('›') /*203A*/, rune('œ') /*0153*/, rune('ž') /*017E*/, rune('Ÿ'), /*0178*/
	//
	rune(0x00A0) /*NBSP*/, rune('¡'), rune('¢'), rune('£'), rune('¤'), rune('¥'), rune('¦'), rune('§'),
	rune('¨'), rune('©'), rune('ª'), rune('«'), rune('¬'), rune(0x00AD) /*SOFT HYPHEN*/, rune('®'), rune('¯'),
	//
	rune('°'), rune('±'), rune('²'), rune('³'), rune('´'), rune('µ'), rune('¶'), rune('·'),
	rune('¸'), rune('¹'), rune('º'), rune('»'), rune('¼'), rune('½'), rune('¾'), rune('¿'),
	//
	rune('À'), rune('Á'), rune('Â'), rune('Ã'), rune('Ä'), rune('Å'), rune('Æ'), rune('Ç'),
	rune('È'), rune('É'), rune('Ê'), rune('Ë'), rune('Ì'), rune('Í'), rune('Î'), rune('Ï'),
	//
	rune('Ð'), rune('Ñ'), rune('Ò'), rune('Ó'), rune('Ô'), rune('Õ'), rune('Ö'), rune('×'),
	rune('Ø'), rune('Ù'), rune('Ú'), rune('Û'), rune('Ü'), rune('Ý'), rune('Þ'), rune('ß'),
	//
	rune('à'), rune('á'), rune('â'), rune('ã'), rune('ä'), rune('å'), rune('æ'), rune('ç'),
	rune('è'), rune('é'), rune('ê'), rune('ë'), rune('ì'), rune('í'), rune('î'), rune('ï'),
	//
	rune('ð'), rune('ñ'), rune('ò'), rune('ó'), rune('ô'), rune('õ'), rune('ö'), rune('÷'),
	rune('ø'), rune('ù'), rune('ú'), rune('û'), rune('ü'), rune('ý'), rune('þ'), rune('ÿ'),
	//
}

// Opens the specified file and parses its content in
// small chunks of valid strings-tokens.
func (p Parser) ParseAnsiFileIntoStrings(
	filepath string,
	buff []byte, //read buffer
	processString func(str string) bool,
) bool {
	strBldr := strings.Builder{}
	strBldr.Grow(cap(buff))
	//open
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
		return false
	}
	defer file.Close()
	//stats
	bytesTotal := 0
	for {
		// Read a chunk
		bytesRead, err := file.Read(buff)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
			return false
		}

		// Stop loop at EOF
		if bytesRead == 0 {
			break
		}

		//convert ANSI to rune
		for i := 0; i < bytesRead; i++ {
			c := Ansi1252Runes[buff[i]]
			strBldr.WriteRune(c)
			if c == 0 {
				fmt.Fprintf(os.Stderr, "ERROR Control-char(%d) found at byte #%d.", buff[i], bytesTotal+i+1)
				return false
			}
		}

		//notify
		str := strBldr.String()
		if !processString(str) {
			fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseAnsiFileIntoStrings'.\n")
			return false
		}
		strBldr.Reset()

		p.bytesCount += uint64(bytesRead)

		//Stats
		bytesTotal += bytesRead
	}
	//
	fmt.Fprintf(os.Stderr, "%d bytes total\n", bytesTotal)
	//
	return true
}
