package sim

import (
	"unicode"
	"fmt"
	"os"
	"strconv"
)

// For now these Titles and Subtitles are valid
var validTitleSubtitles = map[string][]string{
	"Simulation": {"Constants", "Viewports"},
	"Start": 	  {"UniformRect"},
	"Boundaries": {"Periodic", "Reflection"},
	"Sources": 	  {"Point"},
}



type Param struct {
	Title	 string
	Subtitle string
	Name	 string
}


type dType int
const (
	floatT dType = iota
	intT
	vec2T
)


// TODO: we are not chcking if we have all for certain needed params
//  maybe just use default values? or list needed ones per subtitle
var paramMap = map[Param]dType {
	{"Simulation", "Constants", "NSteps"}: intT,
	{"Simulation", "Constants", "Gamma"}: floatT,
	{"Simulation", "Constants", "ParticleMass"}: floatT,
	{"Simulation", "Constants", "DeltaTHalf"}: floatT,
	{"Simulation", "Constants", "Acceleration"}: vec2T,
	{"Simulation", "Viewport", "UpperLeft"}: vec2T,
	{"Simulation", "Viewport", "LowerRight"}: vec2T,
	{"Start", "UniformRect", "NParticles"}: intT,
	{"Start", "UniformRect", "UpperLeft"}: vec2T,
	{"Start", "UniformRect", "LowerRight"}: vec2T,
	{"Boundaries", "Periodic", "Left"}: floatT,
	{"Boundaries", "Periodic", "Right"}: floatT,
	{"Boundaries", "Reflection", "ToOrigin"}: vec2T,
	{"Boundaries", "Reflection", "FromOrigin"}: vec2T,
	{"Source", "Point", "Pos"}: vec2T,
	{"Source", "Point", "Rate"}: floatT,
}


type Source interface {
	Spawn(time float64) []Particle
}

type Boundary struct {
	Offset	   Vec2
	IsPeriodic bool
}

type Reflection struct {
	Offset	   Vec2
	FromOrigin bool
}

type SphConfig struct {
	NSteps 		 	int
	DeltaTHalf		float64
	Gamma  		 	float64
	ParticleMass 	float64
	Acceleration	Vec2

	Boundaries  	[]Boundary
	Reflections	    []Reflection
	Sources 		[]Source
	Start			[]Source // ignores time

	Viewport		[2]Vec2  // upperleft and lower right
}


func MakeConfig(tokens []Token) SphConfig{

	config := SphConfig{
		// default values are zero or nil except:
		Gamma:		  1.66666,
		NSteps:		  10000,
		DeltaTHalf:	  0.001,
		ParticleMass: 1,
		Viewport:  	  [2]Vec2{Vec2{0, 0}, Vec2{1, 1}},
	}

	for len(tokens) > 0 {
		token := tokens[0]
		tokens := tokens[1:]
		if token.Type != title {
			ConfigMakeError(token, fmt.Sprintf("Expected a [[Title]] but got `%v`", token.Type))
		}

		titleStr := token.AsStr
		validSubtitles, ok := validTitleSubtitles[titleStr]
		if !ok {
			ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid title", titleStr))
		}

		if len(tokens) == 0 {
			break
		}
		token, tokens = tokens[0], tokens[1:]

		if token.Type != subtitle {
			ConfigMakeError(token, fmt.Sprintf("Expected a [Subtitle] but got `%v`", token.Type))
		}

		subtitleStr := token.AsStr
		if inSlice(validSubtitles, subtitleStr) {
			ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid subtitle under title: `%v`", subtitleStr, titleStr))
		}

		if len(tokens) == 0 {
			break
		}

		// Here we know we are in the variable definitions
		for len(tokens) > 0 {
			token = tokens[0]
			tokens = tokens[1:]
			if !(token.Type == integer || token.Type == float || token.Type == vec2) {
				ConfigMakeError(token, fmt.Sprintf("Expected either an int, float or Vec2, but got `%v`", token.Type))
			}

			p := Param{titleStr, subtitleStr, token.AsStr}
			_, ok := paramMap[p]
			if !ok {
				ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid parameter in `[[%v]]` - `[%v]`", token.AsStr, titleStr, subtitleStr ))
			}



			// TODO: check if values are set more than once
			switch p{
			case Param{"Simulation", "Constants", "NSteps"}:
				 config.NSteps = checkInt(token, p)
			case Param{"Simulation", "Constants", "Gamma"}:
					 config.Gamma  = checkFloat(token, p)
			case Param{"Simulation", "Constants", "ParticleMass"}:
					 config.ParticleMass  = checkFloat(token, p)
			case Param{"Simulation", "Constants", "DeltaTHalf"}:
					 config.DeltaTHalf   = checkFloat(token, p)
			case Param{"Simulation", "Constants", "Acceleration"}:
					 config.Acceleration = checkVec2(token, p)
			case Param{"Simulation", "Viewport", "UpperLeft"}:
					 config.Viewport[0] = checkVec2(token, p)
			case Param{"Simulation", "Viewport", "LowerRight"}:
					 config.Viewport[1] = checkVec2(token, p)

			case Param{"Start", "UniformRect", "NParticles"}:
				 panic("TOOD: implement [UniformRect]")
			case Param{"Start", "UniformRect", "UpperLeft"},
				 Param{"Start", "UniformRect", "LowerRight"}:
				 panic("TOOD: implement [UniformRect]")


			case Param{"Boundaries", "Periodic", "Left"}:
				// TODO: make somehow sure left < right and so on
					 x := checkFloat(token, p)
				 config.Boundaries = append(config.Boundaries,
				 	Boundary{Offset: Vec2{x, 0}, IsPeriodic: true,})

			case Param{"Boundaries", "Periodic", "Right"}:
					x := checkFloat(token, p)
				 config.Boundaries = append(config.Boundaries,
				 	Boundary{Offset: Vec2{x, 0}, IsPeriodic: true,})

				case Param{"Boundaries", "Periodic", "Top"}:
					 x := checkFloat(token, p)
				 config.Boundaries = append(config.Boundaries,
				 	Boundary{Offset: Vec2{0, x}, IsPeriodic: true,})

			case Param{"Boundaries", "Periodic", "Bottom"}:
					 x := checkFloat(token, p)
				 config.Boundaries = append(config.Boundaries,
				 	Boundary{Offset: Vec2{0, x}, IsPeriodic: true,})

			case Param{"Boundaries", "Reflection", "ToOrigin"}:
				 x := checkVec2(token, p)
				 config.Reflections = append(config.Reflections,
				 	Reflection{Offset: x, FromOrigin: false,})
			case Param{"Boundaries", "Reflection", "FromOrigin"}:
				 x := checkVec2(token, p)
				 config.Reflections = append(config.Reflections,
				 	Reflection{Offset: x, FromOrigin: true,})

			case Param{"Source", "Point", "Pos"}:
				 panic("TOOD: implement [Point]")
			case Param{"Source", "Point", "Rate"}:
				 panic("TOOD: implement [Point]")
			default:
				panic("exhaustive check..")
			}
		}
	}

	return config

}

func checkInt(t Token, p Param) int {
	if t.Type != integer {
		ConfigMakeError(t, fmt.Sprintf("expected an integer but got something else"))
	}
	return int(t.AsInt)
}

func checkFloat(t Token, p Param) float64 {
	if t.Type != integer && t.Type != float  {
		ConfigMakeError(t, fmt.Sprintf("expected an integer or float but got something else"))
	}
	if t.Type == integer {
		return float64(t.AsInt)
	}
	return t.AsFloat
}

func checkVec2(t Token, p Param) Vec2 {
	if t.Type != vec2 {
		ConfigMakeError(t, fmt.Sprintf("expected an integer but got something else"))
	}
	return t.AsVec2
}

func ConfigMakeError(token Token, msg string) {
	fmt.Printf("%v:%v:%v: ConfigMakeError: %v\n", *token.Fname, token.Line, token.Row, msg)
	os.Exit(1)
}



type TokenType int
const (
	title TokenType = iota
	subtitle
	integer
	float
	vec2
)


type Token struct {
	Type    TokenType
	AsStr 	string
	AsInt 	int64
	AsFloat float64
	AsVec2  Vec2

	// if it's a parameter
	Name	string

	// info for error reporting
	Line	int
	Row		int
	Fname	*string
}

func Tokenize(fname string) []Token {
	t	  := MakeTokenizer(fname)
	tokens := make([]Token, 0)

	for {
		t.trimLeftAll()
		if len(t.text) == 0 {
			break
		}
		if t.isExactly('/') {
			t.chop(1)
			t.expect('/')
			t.chop(1)
			t.chopUntilIs(aNewline, "expected a newline")
		} else if t.isExactly('[') {
			t.chop(1)

			if t.is(aLetter) {
				startRow := t.cursor - t.bol + 1
				content := t.chopUntilIs(notALetter, "expected a letter or `]`")
				if len(content) == 0 {
					fmt.Printf("%v:%v:%v: Expected a letter after `[` but got `%v` \n", t.fname, t.line+1, t.cursor - t.bol + 1, string(t.text[0]))
					os.Exit(1)
				}
				t.expect(']')
				t.chop(1)
				t.expectIs(unicode.IsSpace, "a space or newline after `]`")
				//fmt.Println("Title:", string(content))
				tokens = append(tokens, Token{Type: subtitle, AsStr: string(content), Line: t.line+1, Row: startRow, Fname:&fname})
			} else if t.isExactly('[') {
				t.chop(1)
				startRow := t.cursor - t.bol + 1
				content := t.chopUntilIs(notALetter, "expected a letter or `]]`")
				if len(content) == 0 {
					fmt.Printf("%v:%v:%v: Expected a letter after `[[` but got `%v` \n", t.fname, t.line+1, t.cursor - t.bol + 1, string(t.text[0]))
					os.Exit(1)
				}
				t.expect(']')
				t.chop(1)
				t.expect(']')
				t.chop(1)
				t.expectIs(unicode.IsSpace, "a space or newline after `]]`")
				//fmt.Println("Title:", string(content))
				tokens = append(tokens, Token{Type: title, AsStr: string(content), Line: t.line+1, Row: startRow, Fname:&fname})
			} else {
				fmt.Printf("%v:%v:%v: Expected a Title starting with `[[` or a Subtitle starting with `[` and then at least one letter\n", t.fname, t.line+1, t.cursor - t.bol + 1)
				os.Exit(1)
			}

		} else if t.is(aLetter) {
			varName := string(t.chopUntilIs(notALetter, "a letter as start of a parmeter name"))
			startRow := t.cursor - t.bol + 1
			t.expectIs(aWhitespace, "a whitepsace")

			toParse := make([]string, 0, 2)
			nNumbersFound := 0
			isThereAFloat := false
			for {
				t.trimLeft()
				if nNumbersFound == 0 {
					t.expectIs(aDigit, "a digit for parsing as int or float")
				} else {
					if len(t.text) == 0 || t.text[0] == '\n' {
						break
					}
					t.expectIs(aDigit, "a digit for parsing as int or float")
				}

				leftPart := t.chopUntilIsNoFail(notADigit)
				if len(t.text) != 0 && t.text[0] == '.' {
					//got float
					t.chop(1)
					rightPart := t.chopUntilIs(notADigit, "that there is no digit here")
					t.expectIs(unicode.IsSpace, "a whitespace or newline")
					toParse = append(toParse, string(leftPart) + "." + string(rightPart))
					isThereAFloat = true
				} else if len(t.text) == 0 || unicode.IsSpace(t.text[0]) {
					//got int
					toParse = append(toParse, string(leftPart))
				} else {
					fmt.Printf("%v:%v:%v: In number literal expected a `.` or a digit \n", t.fname, t.line+1, t.cursor - t.bol + 1)
					os.Exit(1)
				}
				nNumbersFound += 1
				if nNumbersFound > 2 {
					fmt.Printf("%v:%v:%v: Cannot parse a Vector with more than 2 dimensions \n", t.fname, t.line+1, t.cursor - t.bol + 1)
					os.Exit(1)
				}
			}
			if nNumbersFound == 0 {
				fmt.Printf("%v:%v:%v: Expected at least one numbere here \n", t.fname, t.line+1, t.cursor - t.bol + 1)
				os.Exit(1)
			}


			// one numer
			if nNumbersFound == 1 {
				if isThereAFloat {
					// parse as float
					x, err := strconv.ParseFloat(toParse[0], 64)
					check(err)
					tokens = append(tokens, Token{Name: varName, Type: float, AsFloat: x, Line: t.line+1, Row: startRow, Fname:&fname})

				} else {
					// parse as int
					x, err := strconv.ParseInt(toParse[0], 10, 64)
					check(err)
					tokens = append(tokens, Token{Name: varName, Type: integer, AsInt: x, Line: t.line+1, Row: startRow, Fname:&fname})
				}
			} else if nNumbersFound == 2 {

					// parse as float
					x1, err := strconv.ParseFloat(toParse[0], 64)
					check(err)
					x2, err := strconv.ParseFloat(toParse[1], 64)
					check(err)
					tokens = append(tokens, Token{Name: varName, Type: vec2, AsVec2: Vec2{x1, x2}, Line: t.line+1, Row: startRow, Fname:&fname})
			} else {
				panic("unreachable")
			}

		} else {
			fmt.Println(tokens)
			panic(fmt.Sprintf("%v:%v: unimplemented - starts with: `%v`", t.line, t.cursor - t.bol + 1, string(t.text[0])))
		}
	}

	return tokens
}




func check(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadEntireFile(fname string) []rune {
	content, err := os.ReadFile(fname)
	check(err)
	return []rune(string(content))
}

var expectedTitles = [3]string{"Constants", "Boundaries", "StartGeometry"}

func isIn(sArray []string, s string) bool {
	for _, it := range sArray {
		if s == it {
			return true
		}
	}
	return false
}

func aLetter(c rune) bool {
    return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func notALetter(c rune) bool {
    return !(('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z'))
}

func aDigit(c rune) bool {
    return ('0' <= c && c <= '9')
}

func notADigit(c rune) bool {
    return !('0' <= c && c <= '9')
}

func aWhitespace(r rune) bool {
	return unicode.IsSpace(r) && r != '\n'
}

func notAWhitespace(r rune) bool {
	return !(unicode.IsSpace(r) && r != '\n')
}

func aNewline(r rune) bool {
	return r == '\n'
}

func aDotOrWhitespace(r rune) bool {
	return r == '.' || aWhitespace(r)
}

type Tokenizer struct {
	fname  string
	text   []rune
	bol    int
	line   int
	cursor int
}

func MakeTokenizer(fname string) (tokenizer Tokenizer) {

	runes := ReadEntireFile(fname)
	t := Tokenizer{
		fname: fname,
		text: runes,
	}
	return t
}


func (t *Tokenizer) chop(n int) {
	if n > len(t.text) {
		panic("trying to chop more than it has!")
	}

	for _ = range n {
		r := t.text[0]
		t.cursor += 1
		t.text = t.text[1:]
		if r == '\n' {
			t.line   += 1
			t.bol     = t.cursor
		}
	}

}

func (t *Tokenizer) chopUntilIs(pred func(r rune) bool, expectationMsg string) []rune {
	i 	 := 0
	text := t.text
	for ;len(t.text) > 0 && !pred(t.text[0]); i++ {
		t.chop(1)
	}
	if len(t.text) == 0 {
		fmt.Printf("%v:%v:%v: Expected %v, but got to the end of the file \n", t.fname, t.line+1, t.cursor - t.bol + 1, expectationMsg)
		os.Exit(1)
	}
	return text[:i]
}

func (t *Tokenizer) chopUntilIsNoFail(pred func(r rune) bool) ([]rune) {
	i 	 := 0
	text := t.text
	for ;len(t.text) > 0 && !pred(t.text[0]); i++ {
		t.chop(1)
	}
	return text[:i]
}

func (t *Tokenizer) chopUntilExactly(r rune) []rune {
	i 	 := 0
	text := t.text
	for ;len(t.text) > 0 && r != t.text[0]; i++ {
		t.chop(1)
	}
	if len(t.text) == 0 {
		fmt.Printf("%v:%v:%v: Expected `%v`, but got to the end of the file \n", t.fname, t.line+1, t.cursor - t.bol + 1, string(r))
		os.Exit(1)
	}
	return text[:i]
}

func (t *Tokenizer) trimLeftAll() {
	for len(t.text) > 0 && unicode.IsSpace(t.text[0]) {
		t.chop(1)
	}
}

func (t *Tokenizer) trimLeft() {
	for len(t.text) > 0 && aWhitespace(t.text[0]) {
		t.chop(1)
	}
}

func (t *Tokenizer) is(predicate func(rune) bool) bool {
	if len(t.text) == 0 {
		panic("is: no more tokenz (should not happen!)")
	}
	return predicate(t.text[0])
}

func (t *Tokenizer) isExactly(r rune) bool {
	if len(t.text) == 0 {
		panic("isExactly: no more tokenz (should not happen!)")
	}
	return t.text[0] == r
}

func (t *Tokenizer) expect(r rune) {
	if len(t.text) == 0 {
		panic("expect: no more tokenz (should not happen!)")
	}

	if r != t.text[0] {
		fmt.Printf("%v:%v:%v: Expected `%v`, but we got this: `%v`\n", t.fname, t.line+1, t.cursor - t.bol + 1, string(r), string(t.text[0]))
		os.Exit(1)
	}
}

func (t *Tokenizer) expectIs(pred func(rune) bool, expectationMsg string) {
	if len(t.text) == 0 {
		panic("expect: no more tokenz (should not happen!)")
	}

	if !pred(t.text[0]) {
		fmt.Printf("%v:%v:%v: Expected %v \n", t.fname, t.line+1, t.cursor - t.bol + 1, expectationMsg)
		os.Exit(1)
	}
}

func inSlice(list []string, a string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}