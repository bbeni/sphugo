/* This is the config parser

TODO: complete and actually implement the functianilty in the simulation

It seems a big bloat for what it is doing. It should not be this big,
but i just wanted to experiment with parsing files in general and see
how to implement useful error messages. I hope the error messages help
to debug config files.

*/

package sim

import (
	"unicode"
	"fmt"
	"os"
	"math/rand"
	"time"
	"strconv"
)

// For now these Titles and Subtitles are valid
var validTitleSubtitles = map[string][]string{
	"Simulation": {"Constants", "Viewport"},
	"Start": 	  {"UniformRect"},
	"Boundaries": {"Periodic", "Reflection"},
	"Sources": 	  {"Point"},
}

/* Valid params:
  // title,        subtitle,     paramName
	{"Simulation", "Constants",  "NSteps"},
	{"Simulation", "Constants",  "Gamma"},
	{"Simulation", "Constants",  "ParticleMass"},
	{"Simulation", "Constants",  "DeltaTHalf"},
	{"Simulation", "Constants",  "Acceleration"},
	{"Simulation", "Viewport",   "UpperLeft"},
	{"Simulation", "Viewport",   "LowerRight"},
	{"Start",      "UniformRect","NParticles"},
	{"Start",      "UniformRect","UpperLeft"},
	{"Start",      "UniformRect","LowerRight"},
	{"Boundaries", "Periodic",   "Left"},
	{"Boundaries", "Periodic",   "Right"},
	{"Boundaries", "Reflection", "ToOrigin"},
	{"Boundaries", "Reflection", "FromOrigin"},
	{"Sources",    "Point",      "Pos"},
	{"Sources",    "Point",      "Rate"},
} */



type ParticleSource interface {
	Spawn(t float64) []Particle
}

type UniformRectSpawner struct {
	UpperLeft  Vec2
	LowerRight Vec2
	NParticles int
}

// sensible defaults
func MakeUniformRectSpawner() UniformRectSpawner{
	return UniformRectSpawner{
		LowerRight: Vec2{1, 1},
		NParticles: 1000,
	}
}

// spawn onece uniformely in this rect
func (spwn UniformRectSpawner) Spawn(t float64) []Particle {

	if USE_RANDOM_SEED {
		rand.Seed(time.Now().UnixNano())
	} else {
	    rand.Seed(12345678)
	}

 	particles := make([]Particle, spwn.NParticles)

	for i := range spwn.NParticles {
		x := spwn.UpperLeft.X + rand.Float64() * (spwn.LowerRight.X - spwn.UpperLeft.X)
		y := spwn.UpperLeft.Y + rand.Float64() * (spwn.LowerRight.Y - spwn.UpperLeft.Y)
		particles[i].Pos = Vec2{x, y}
	}

	for i := range spwn.NParticles {
		particles[i].Z = rand.Int()
		particles[i].E = 0.01
	}

	return particles
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
	Sources 		[]ParticleSource
	Start			[]ParticleSource

	Viewport		[2]Vec2  // upperleft and lower right
}


// default values conifg all valuues are zero or empty arrays except defined in this function:
func MakeConfig() SphConfig {
	return SphConfig{
		Gamma:		  1.66666,
		NSteps:		  10000,
		DeltaTHalf:	  0.001,
		ParticleMass: 1,
		Viewport:  	  [2]Vec2{Vec2{0, 0}, Vec2{1, 1}},
	}
}

func MakeConfigFromFile(configFilePath string) SphConfig {
	tokens := Tokenize(configFilePath)
	config := MakeConfig()
	config.updateFromTokens(tokens)
	return config
}


type Param struct {
	Title	 string
	Subtitle string
	Name	 string
}

// This function generates a configuration given a tokenized config file
func (config *SphConfig) updateFromTokens(tokens []Token) {

	var token Token
	for len(tokens) > 0 {
		token, tokens = tokens[0], tokens[1:]

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
		if !inSlice(validSubtitles, subtitleStr) {
			ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid subtitle under title: `%v`. It's valid subtitles are: %v", subtitleStr, titleStr, validSubtitles))
		}

		if len(tokens) == 0 {
			break
		}

		for len(tokens) > 0 {

			if tokens[0].Type == title {
				break
			}

			token, tokens = tokens[0], tokens[1:]

			if token.Type == subtitle {
				subtitleStr = token.AsStr

				if !inSlice(validSubtitles, subtitleStr) {
					ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid subtitle under title: `%v`. It's valid subtitles incude: %v", subtitleStr, titleStr, validSubtitles))
				}
				continue
			}



			// Here we know we should be in the variable definitions
			// this should never happen?
			if token.Type != integer && token.Type != float && token.Type != vec2 {
				ConfigMakeError(token, fmt.Sprintf("Expected either an int, float or Vec2 parameter definition, but got `%v`", token.Type))
			}

			p := Param{titleStr, subtitleStr, token.Name}

			// TODO: check if values are set more than once
			switch p {
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

			case Param{"Start", "UniformRect", "NParticles"},
			     Param{"Start", "UniformRect", "UpperLeft"},
				 Param{"Start", "UniformRect", "LowerRight"}:

				// Expect exactly these 3 Params
				if len(tokens) < 2 {
					ConfigMakeError(token, fmt.Sprintf("Expected 3 Params following [`%v`] but got something else or nothing.\n\tOne of each paramas `NParticles, UpperLeft, LowerRight` must be defined", subtitleStr, ))
				}

				startSpawner := UniformRectSpawner{}
				got := make([]string, 3)

				for i := range 3 {

					if inSlice(got, token.Name) {
						ConfigMakeError(token, fmt.Sprintf("The parameter name `%v` in [`%v`] is already set!", token.Name, subtitleStr))
					}

					if token.Type != integer && token.Type != vec2 {
						ConfigMakeError(token, fmt.Sprintf("Expected either an integer or vec2 parameter definition in [`%v`], but got `%v`", subtitleStr, token.Type))
					}

					switch token.Name {
						case "NParticles":
							startSpawner.NParticles = checkInt(token, p)
						case "UpperLeft":
							startSpawner.UpperLeft = checkVec2(token, p)
						case "LowerRight":
							startSpawner.LowerRight = checkVec2(token, p)
						default:
							ConfigMakeError(token, fmt.Sprintf("The parameter name `%v` in [`%v`] is not valid. Needs to be one of `NParticles, UpperLeft, LowerRight`", token.Name, subtitleStr))
					}

					got = append(got, p.Name)

					if i==2 {
						break
					}
					token, tokens = tokens[0], tokens[1:]
				}

				if tokens[0].Type != title && tokens[0].Type != subtitle {
					ConfigMakeError(token, fmt.Sprintf("The 4th parameter name `%v` in [`%v`] is too much. need to have `NParticles, UpperLeft and LowerRight`", token.Name, subtitleStr))
				}

				config.Start = append(config.Start, startSpawner)


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

			case Param{"Sources", "Point", "Pos"}:
				 panic("TOOD: implement [Point]")
			case Param{"Sources", "Point", "Rate"}:
				 panic("TOOD: implement [Point]")
			default:
				ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid parameter in `[[%v]]` - `[%v]`", token.Name, titleStr, subtitleStr ))
			}
		}
	}
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

//go:generate stringer -type TokenType
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
			t.chopUntilIsNoFail(aNewline)
			if len(t.text) == 0 {
				// we are finished
				break
			}
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

func GenerateDefaultConfigFile(filePath string) {

	var exampleConfigSource = `//  Config file for SPHUGO SPH Simulation
//  This is an example configuration.
//  <- This is a line comment (only allowed at start of line)
//
//	[[Boundaries]]
//  [[Sources]]
//  are not implemented as of now !!
//

[[Simulation]]
[Constants]
NSteps              1000
Gamma               1.666
ParticleMass        1
// A 2-D Vector just has 2 components separated by space(s)
Acceleration        0       1.5
DeltaTHalf          0.004

// Coordinates of viewport for animation
[Viewport]
UpperLeft           0       0
LowerRight          1       1

// Initial configuration
[[Start]]
[UniformRect]
NParticles          2500
UpperLeft           0.1     0.3
LowerRight          0.6     0.7

// Per default boundaries are open
[[Boundaries]]
[Periodic]
Left                0
Right               1

// A reflection reflects particles without losing momentum
[Reflection]
// A refelection line orthognal to origin that refelcts towards origin
// ToOrigin         0       0.9
// A reflection boundary taht excludes origin (top-left corner)
// FromOrigin       0.2     0.2

[[Sources]]
[Point]
//Pos                 0.2     0.2
//Rate                100`


	if _, err := os.Stat("./example.sph-config"); err != nil {
	    if os.IsNotExist(err) {

			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error: couldn't create file %q : %q", filePath, err)
				os.Exit(1)
			}
			defer file.Close()

			file.WriteString(exampleConfigSource)
	    }
	}
}