package parser

// Parser accumules stats and current values.
type Parser struct {
	bytesCount uint64 //ammount of bytes processed
}

/*
// Token is a sequence of non-separators runes or a single separator rune.
// During a first pass, decimal numbers could be represented
// by multiple tokens, like '3' + '.' + '1416'. For those cases,
// a second pass is required to merge tokens according to the
// context of what is being parsed.
type Token struct {
	runes       []rune
	isSeparator bool
}
*/

/*
// Opens the specified file and parses its content in
// small chunks of valid runes-tokens using the specified
// list of tokens-separator runes.
func (p Parser) ParseFilepathIntoTokens(
	filepath string,
	processToken func(runes []rune, isSeparator bool) bool,
	separators map[rune]rune,
) bool {
	//reset
	token := make([]rune, 0, 32)
	//
	cb := func(str string) bool {
		runes := []rune(str)
		for i, v := range runes {
			_, isSeparator := separators[v]
			if isSeparator {
				//is separator,
				//flush accum
				if len(token) > 0 {
					if !processToken(token, false) {
						fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathIntoTokens'.\n")
						return false
					}
					p.tokensCount++
				}
				if !processToken(runes[i:i+1], true) {
					fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathIntoTokens'.\n")
					return false
				}
				p.tokensCount++
				token = token[:0]
			} else {
				//is not separator
				token = append(token, v)
			}
		}
		return true
	}
	//
	if !p.ParseFilepathIntoStrings(filepath, cb) {
		return false
	}
	//flush triling data
	if len(token) > 0 {
		if !processToken(token, false) {
			fmt.Fprintf(os.Stderr, "Cancelation by callback at 'ParseFilepathIntoTokens'.\n")
			return false
		}
		p.tokensCount++
		token = token[:0]
	}
	//
	return true
}
*/

/*
// Appends non-digits and non-letters runes to provided map.
func AddDefaultSeparatorRunes(
	dst map[rune]rune,
) {
	//Typical separators
	var i rune
	for i = 0; i <= 47; i++ { //1 - 47 = 'start of heading' - /
		dst[i] = i
	}
	//48 - 57 = 0 - 9
	for i = 58; i <= 64; i++ { //58 - 64 = : - @
		dst[i] = i
	}
	//65 - 90 = A - Z
	for i = 91; i <= 96; i++ { //91 - 96 = [ - `
		dst[i] = i
	}
	//97 - 122 = a - z
	for i = 123; i <= 191; i++ { //123 - 191 = { - ┐┐
		dst[i] = i
	}
	//192 - 214 =
	for i = 215; i <= 215; i++ { //215 - 215 = %
		dst[i] = i
	}
	//216 - 246 =
	for i = 247; i <= 247; i++ { //247 - 247 = %
		dst[i] = i
	}
	//248 - 255 =
}
*/
