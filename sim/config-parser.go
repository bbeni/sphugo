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
	"math"
	"errors"
)

// For now these Titles and Subtitles are valid
var validTitleSubtitles = map[string][]string{
	"Simulation": {"Config", "Viewport"},
	"Start": 	  {"UniformRect"},
	"Boundaries": {"Periodic", "Reflection"},
	"Sources": 	  {"Point"},
}

type ParticleSource interface {
	Spawn(t float64) []Particle
}

type UniformRectSpawner struct {
	UpperLeft  Vec2
	LowerRight Vec2
	NParticles int
}

type PointSource struct {
	origin     Vec2
	rate       float64
	LastSpwned float64
}

// sensible defaults
func MakeUniformRectSpawner() UniformRectSpawner{
	return UniformRectSpawner {
		LowerRight: Vec2{1, 1},
		NParticles: 1000,
	}
}

// spawn once uniformely in this rect
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

func (spwn *PointSource) Spawn(t float64) []Particle {
	deltaT   := t - spwn.LastSpwned
	cooldown := 1 / spwn.rate

	n := int(deltaT / cooldown)
	spwn.LastSpwned += float64(n) * cooldown

	particles := make([]Particle, n)
	for i := range n {
		dy := 0.01 * (-1 + 2 * rand.Float64())
		dx := 0.01 * (-1 + 2 * rand.Float64())
		pos := Vec2{spwn.origin.X + dx, spwn.origin.Y + dy}

		particles[i].Pos = pos
		particles[i].Rho = 100
		particles[i].Z = rand.Int()
		particles[i].E = 0.002
	}

	return particles
}

type Reflections struct {
	L float64
	R float64
	U float64
	D float64	
}

type SphConfig struct {
	NSteps 		 	int
	DeltaTHalf		float64
	Gamma  		 	float64
	ParticleMass 	float64
	Acceleration	Vec2

	Kernel			Kernel

	HorPeriodicity  [2]float64 // -math.MaxFloat64, math.MaxFloat64 is open
	VertPeriodicity [2]float64 // -math.MaxFloat64, math.MaxFloat64 is open

	Reflections	    Reflections
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
		Kernel:		  Monahan2D,

		VertPeriodicity: [2]float64{-math.MaxFloat64, math.MaxFloat64},
		HorPeriodicity:  [2]float64{-math.MaxFloat64, math.MaxFloat64},

		Reflections: Reflections{
			L: -math.MaxFloat64, R: math.MaxFloat64,
			U: -math.MaxFloat64, D: math.MaxFloat64,
		},

		Viewport:  	  [2]Vec2{Vec2{0, 0}, Vec2{1, 1}},
	}
}

func MakeConfigFromFile(configFilePath string) (error, SphConfig) {

	config := MakeConfig()

	err, tokens := Tokenize(configFilePath)
	if err != nil {
		return err, config
	}

	err = config.updateFromTokens(tokens)

	if err != nil {
		return err, MakeConfig() // new default config
	}

	return nil, config
}

func ConfigMakeError(token Token, msg string) error {
	m := fmt.Sprintf("%v:%v:%v: ConfigMakeError: %v\n", *token.Fname, token.Line, token.Row, msg)
	return errors.New(m)
}

func ConfigParseError(t Tokenizer, msg string) error {
	m := fmt.Sprintf("%v:%v:%v: ConfigParseError: %v\n", t.fname, t.line+1, t.cursor - t.bol + 1, msg)
	return errors.New(m)
}

type Param struct {
	Title	 string
	Subtitle string
	Name	 string
}

// This function generates a configuration given a tokenized config file
func (config *SphConfig) updateFromTokens(tokens []Token) error {

	var token Token
	for len(tokens) > 0 {
		token, tokens = tokens[0], tokens[1:]

		if token.Type != title {
			return ConfigMakeError(token, fmt.Sprintf("Expected a [[Title]] but got `%v`", token.Type))
		}

		titleStr := token.AsStr
		validSubtitles, ok := validTitleSubtitles[titleStr]
		if !ok {
			return ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid title", titleStr))
		}

		if len(tokens) == 0 {
			break
		}
		token, tokens = tokens[0], tokens[1:]

		if token.Type != subtitle {
			return ConfigMakeError(token, fmt.Sprintf("Expected a [Subtitle] but got `%v`", token.Type))
		}

		subtitleStr := token.AsStr
		if !inSlice(validSubtitles, subtitleStr) {
			return ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid subtitle under title: `%v`. It's valid subtitles are: %v", subtitleStr, titleStr, validSubtitles))
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
					return ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid subtitle under title: `%v`. It's valid subtitles incude: %v", subtitleStr, titleStr, validSubtitles))
				}
				continue
			}

			// Here we know we should be in the variable definitions
			// this should never happen?
			if token.Type != integer && token.Type != float && token.Type != vec2 && token.Type != word {
				return ConfigMakeError(token, fmt.Sprintf("Expected either an int, float, vec2 or word parameter definition, but got `%v`", token.Type))
			}

			var err error
			p := Param{titleStr, subtitleStr, token.Name}

			// TODO: check if values are set more than once
			switch p {
			case Param{"Simulation", "Config", "NSteps"}:
				config.NSteps, err = checkInt(token, p)
				if err != nil { return err }
			case Param{"Simulation", "Config", "Gamma"}:
				config.Gamma , err = checkFloat(token, p)
				if err != nil { return err }
			case Param{"Simulation", "Config", "ParticleMass"}:
				config.ParticleMass, err = checkFloat(token, p)
				if err != nil { return err }
			case Param{"Simulation", "Config", "DeltaTHalf"}:
				config.DeltaTHalf, err = checkFloat(token, p)
				if err != nil { return err }
			case Param{"Simulation", "Config", "Acceleration"}:
				config.Acceleration, err = checkVec2(token, p)
				if err != nil { return err }
			case Param{"Simulation", "Config", "Kernel"}:
				kernel := token.AsStr
				if kernel == "Monahan" {
					config.Kernel = Monahan2D
				} else if kernel == "Wendtland" {
					config.Kernel = Wendtland2D
				} else {
					return ConfigMakeError(token, fmt.Sprintf("Kernel `%v` is not implemented", kernel))
				}

			case Param{"Simulation", "Viewport", "UpperLeft"}:
					 config.Viewport[0], err = checkVec2(token, p)
			case Param{"Simulation", "Viewport", "LowerRight"}:
					 config.Viewport[1], err = checkVec2(token, p)

			case Param{"Start", "UniformRect", "NParticles"},
			     Param{"Start", "UniformRect", "UpperLeft"},
				 Param{"Start", "UniformRect", "LowerRight"}:

				// Expect exactly these 3 Params
				if len(tokens) < 2 {
					return ConfigMakeError(token, fmt.Sprintf("Expected 3 Params following [`%v`] but got something else or nothing.\n\tOne of each paramas `NParticles, UpperLeft, LowerRight` must be defined", subtitleStr, ))
				}

				startSpawner := UniformRectSpawner{}
				got := make([]string, 3)

				for i := range 3 {

					if inSlice(got, token.Name) {
						return ConfigMakeError(token, fmt.Sprintf("The parameter name `%v` in [`%v`] is already set!", token.Name, subtitleStr))
					}

					if token.Type != integer && token.Type != vec2 {
						return ConfigMakeError(token, fmt.Sprintf("Expected either an integer or vec2 parameter definition in [`%v`], but got `%v`", subtitleStr, token.Type))
					}

					switch token.Name {
						case "NParticles":
							startSpawner.NParticles, err = checkInt(token, p)
							if err != nil { return err }
						case "UpperLeft":
							startSpawner.UpperLeft, err = checkVec2(token, p)
							if err != nil { return err }
						case "LowerRight":
							startSpawner.LowerRight, err = checkVec2(token, p)
							if err != nil { return err }
						default:
							return ConfigMakeError(token, fmt.Sprintf("The parameter name `%v` in [`%v`] is not valid. Needs to be one of `NParticles, UpperLeft, LowerRight`", token.Name, subtitleStr))
					}

					got = append(got, p.Name)

					if i==2 {
						break
					}
					token, tokens = tokens[0], tokens[1:]
				}

				if tokens[0].Type != title && tokens[0].Type != subtitle {
					return ConfigMakeError(token, fmt.Sprintf("The 4th parameter name `%v` in [`%v`] is too much. need to have `NParticles, UpperLeft and LowerRight`", token.Name, subtitleStr))
				}

				config.Start = append(config.Start, startSpawner)


			case Param{"Boundaries", "Periodic", "Vertical"}:
				// TODO: make somehow sure left < right and so on
				x, err := checkVec2(token, p)
				if err != nil { return err }
				config.VertPeriodicity = [2]float64{x.X, x.Y}

			case Param{"Boundaries", "Periodic", "Horizontal"}:
				x, err := checkVec2(token, p)
				if err != nil { return err }
				config.HorPeriodicity = [2]float64{x.X, x.Y}

			case Param{"Boundaries", "Reflection", "Left"}:
				x, err := checkFloat(token, p)
				if err != nil { return err }
				config.Reflections.L = x
			case Param{"Boundaries", "Reflection", "Right"}:
				x, err := checkFloat(token, p)
				if err != nil { return err }
				config.Reflections.R = x
			case Param{"Boundaries", "Reflection", "Up"}:
				x, err := checkFloat(token, p)
				if err != nil { return err }
				config.Reflections.U = x
			case Param{"Boundaries", "Reflection", "Down"}:
				x, err := checkFloat(token, p)
				if err != nil { return err }
				config.Reflections.D = x
			
			case Param{"Sources", "Point", "Pos"},
				Param{"Sources", "Point", "Rate"}:

				var pos  Vec2
				var rate float64
				if token.Name == "Pos" {
					pos, err = checkVec2(token, p)
					if err != nil { return err }
					if len(tokens) == 0 {
						return ConfigMakeError(token, fmt.Sprintf("Expected `rate` but got nothing"))
					}
					token, tokens = tokens[0], tokens[1:]
					rate, err = checkFloat(token, p)
					if err != nil { return err }

				} else {
					rate, err = checkFloat(token, p)
					if err != nil { return err }
					if len(tokens) == 0 {
						return ConfigMakeError(token, fmt.Sprintf("Expected `pos` but got nothing"))
					}
					token, tokens = tokens[0], tokens[1:]
					pos, err = checkVec2(token, p)
					if err != nil { return err }
				}

				pointSource := &PointSource{
					rate:   rate,
					origin: pos,
				}

				config.Sources = append(config.Sources, pointSource)

			default:
				return ConfigMakeError(token, fmt.Sprintf("`%v` is not a valid parameter in `[[%v]]` - `[%v]`", token.Name, titleStr, subtitleStr ))
			}
		}
	}

	return nil
}

func checkInt(t Token, p Param) (int, error) {
	if t.Type != integer {
		return 0, ConfigMakeError(t, fmt.Sprintf("expected an integer but got something else"))
	}
	return int(t.AsInt), nil
}

func checkFloat(t Token, p Param) (float64, error) {
	if t.Type != integer && t.Type != float  {
		return 0, ConfigMakeError(t, fmt.Sprintf("expected an integer or float but got something else"))
	}
	if t.Type == integer {
		return float64(t.AsInt), nil
	}
	return t.AsFloat, nil
}

func checkVec2(t Token, p Param) (Vec2, error) {
	if t.Type != vec2 {
		return Vec2{}, ConfigMakeError(t, fmt.Sprintf("expected an integer but got something else"))
	}
	return t.AsVec2, nil
}

//go:generate stringer -type TokenType
type TokenType int
const (
	title TokenType = iota
	subtitle
	integer
	float
	vec2
	word
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

func Tokenize(fname string) (error, []Token) {
	t	   := MakeTokenizer(fname)
	tokens := make([]Token, 0)

	for {
		t.trimLeftAll()
		if len(t.text) == 0 {
			break
		}
		if t.isExactly('/') {
			t.chop(1)
			err := t.expect('/')
			if err != nil {
				return err, tokens
			}

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
				err, content := t.chopUntilIs(notALetter, "expected a letter or `]`")
				if err != nil {
					return err, tokens
				}

				if len(content) == 0 {
					msg := fmt.Sprintf("Expected a letter after `[` but got `%v` \n", string(t.text[0]))
					return ConfigParseError(t, msg), tokens
				}

				err = t.expect(']')
				if err != nil {
					return err, tokens
				}
				t.chop(1)

				t.expectIs(unicode.IsSpace, "a space or newline after `]`")
				tokens = append(tokens, Token{Type: subtitle, AsStr: string(content), Line: t.line+1, Row: startRow, Fname:&fname})
			} else if t.isExactly('[') {
				t.chop(1)
				startRow := t.cursor - t.bol + 1
				err, content := t.chopUntilIs(notALetter, "expected a letter or `]]`")

				if err != nil {
					return err, tokens
				}

				if len(content) == 0 {
					msg := fmt.Sprintf("Expected a letter after `[[` but got `%v`", string(t.text[0]))
					return ConfigParseError(t, msg), tokens
				}

				err = t.expect(']')
				if err != nil {
					return err, tokens
				}
				t.chop(1)

				err = t.expect(']')
				if err != nil {
					return err, tokens
				}
				t.chop(1)

				t.expectIs(unicode.IsSpace, "a space or newline after `]]`")
				tokens = append(tokens, Token{Type: title, AsStr: string(content), Line: t.line+1, Row: startRow, Fname:&fname})
			} else {
				msg := fmt.Sprintf("Expected a Title starting with `[[` or a Subtitle starting with `[` and then at least one letter")
				return ConfigParseError(t, msg), tokens
			}

		} else if t.is(aLetter) {
			err, chopped := t.chopUntilIs(notALetter, "a letter as start of a parmeter name")
			if err != nil {
				return err, tokens
			}

			varName := string(chopped)
			startRow := t.cursor - t.bol + 1
			t.expectIs(aWhitespace, "a whitepsace")
			t.trimLeft()

			if t.is(aDigit) || t.isExactly('-') { // parse a float/int/vec2
				toParse := make([]string, 0, 2)
				nNumbersFound := 0
				isThereAFloat := false

				for {
					t.trimLeft()

					negative := false
					if t.isExactly('-') {
						negative = true
						t.chop(1)
					}

					if nNumbersFound != 0 {
						if len(t.text) == 0 || t.text[0] == '\n' {
							break
						}
						t.expectIs(aDigit, "a digit for parsing as int or float")
					}

					leftPart := t.chopUntilIsNoFail(notADigit)

					if negative {
						x := make([]rune, 1, len(leftPart)+1)
						x[0] = '-'
						leftPart = append(x, leftPart...)
					}

					if len(t.text) != 0 && t.text[0] == '.' {
						//got float
						t.chop(1)
						err, rightPart := t.chopUntilIs(notADigit, "that there is no digit here")
						if err != nil {
							return err, tokens
						}
						t.expectIs(unicode.IsSpace, "a whitespace or newline")
						toParse = append(toParse, string(leftPart) + "." + string(rightPart))
						isThereAFloat = true
					} else if len(t.text) == 0 || unicode.IsSpace(t.text[0]) {
						//got int
						toParse = append(toParse, string(leftPart))
					} else {
						msg := fmt.Sprintf("In number literal expected a `.` or a digit")
						return ConfigParseError(t, msg), tokens
					}
					nNumbersFound += 1
					if nNumbersFound > 2 {
						msg := fmt.Sprintf("Cannot parse a Vector with more than 2 dimensions")
						return ConfigParseError(t, msg), tokens
					}
				}
				if nNumbersFound == 0 {
					msg := fmt.Sprintf("Expected at least one number here")
					return ConfigParseError(t, msg), tokens
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

						// parse as vec2
						x1, err := strconv.ParseFloat(toParse[0], 64)
						check(err)
						x2, err := strconv.ParseFloat(toParse[1], 64)
						check(err)
						tokens = append(tokens, Token{Name: varName, Type: vec2, AsVec2: Vec2{x1, x2}, Line: t.line+1, Row: startRow, Fname:&fname})
				} else {
					panic("unreachable")
				}


			} else if t.is(aLetter) { // parse a word
				supposedWord := string(t.chopUntilIsNoFail(notALetter))
				if len(supposedWord) == 0 || !t.is(unicode.IsSpace) {
					msg := fmt.Sprintf("`%v` Excpected only Letters and a whitespace at the end to form a Word", string(supposedWord))
					return ConfigParseError(t, msg), tokens
				}

				tokens = append(tokens, Token{Name: varName, Type: word, AsStr: supposedWord, Line: t.line+1, Row: startRow, Fname: &fname})

			} else {
				msg := fmt.Sprintf("`%v` Excpected a Number/Vec2 or Word", varName)
				return ConfigParseError(t, msg), tokens
			}

		} else {
			fmt.Println(tokens)
			panic(fmt.Sprintf("%v:%v: unimplemented - starts with: `%v`\n", t.line, t.cursor - t.bol + 1, string(t.text[0])))
		}
	}

	return nil, tokens
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

func (t *Tokenizer) chopUntilIs(pred func(r rune) bool, expectationMsg string) (error, []rune) {
	i 	 := 0
	text := t.text
	for ;len(t.text) > 0 && !pred(t.text[0]); i++ {
		t.chop(1)
	}
	if len(t.text) == 0 {
		msg := fmt.Sprintf("Expected %v, but got to the end of the file", expectationMsg)
		return ConfigParseError(*t, msg), []rune{}
	}
	return nil, text[:i]
}

func (t *Tokenizer) chopUntilIsNoFail(pred func(r rune) bool) ([]rune) {
	i 	 := 0
	text := t.text
	for ;len(t.text) > 0 && !pred(t.text[0]); i++ {
		t.chop(1)
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

func (t *Tokenizer) expect(r rune) error {
	if len(t.text) == 0 {
		panic("expect: no more tokenz (should not happen!)")
	}

	if r != t.text[0] {
		msg := fmt.Sprintf("Expected `%v`, but we got this: `%v`", string(r), string(t.text[0]))
		return ConfigParseError(*t, msg)
	}
	return nil
}

func (t *Tokenizer) expectIs(pred func(rune) bool, expectationMsg string) error {
	if len(t.text) == 0 {
		panic("expect: no more tokenz (should not happen!)")
	}

	if !pred(t.text[0]) {
		msg := fmt.Sprintf("Expected `%v`", expectationMsg)
		return ConfigParseError(*t, msg)
	}
	return nil
}

func inSlice(list []string, a string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}


func GenerateDefaultConfigFiles(filePaths [2]string) {
	GenerateTextFile(filePaths[0], exampleConfigSource1)
	GenerateTextFile(filePaths[1], exampleConfigSource2)
}

func GenerateTextFile(filePath string, source string) {
	if _, err := os.Stat(filePath); err != nil {
	    if os.IsNotExist(err) {
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error: couldn't create file %q : %q", filePath, err)
				os.Exit(1)
			}
			defer file.Close()

			file.WriteString(source)
	    }
	}
}

var exampleConfigSource1 = `//  Config file for SPHUGO SPH Simulation
//  This is an example configuration.
//  <- This is a line comment (only allowed at start of line)
//  The coordinate origin 0 0 is in the top left corner x direction right, y direction down
//

[[Simulation]]
[Config]
NSteps              1000
Gamma               4.666
ParticleMass        1000000.0
// A 2-D Vector just has 2 components separated by space(s)
Acceleration        0       0.55
DeltaTHalf          0.00324
//Kernel            Monahan
Kernel              Wendtland

// Initial setup of particles, for now we can add Uniformely Random distributed Rectangels only
[[Start]]

[UniformRect]
NParticles          260
UpperLeft           0.6     0.2
LowerRight          0.79    0.3

[UniformRect]
NParticles          700
UpperLeft           0.27    0.3
LowerRight          0.4     0.9

// Per default boundaries are open
[[Boundaries]]
[Periodic]
Horizontal          0.2       0.8
Vertical            -100      100

// THIS IS NOT IMPLEMENTED
// A reflection reflects particles without losing momentum
[Reflection]
Down                0.99

//[[Sources]]
//[Point]
//Pos               0.21     0.21
//Rate              10

// THIS IS NOT IMPLEMENTED
// Coordinates of viewport for animation
[[Simulation]]
[Viewport]
UpperLeft           0       0
LowerRight          1       1
`

var exampleConfigSource2 = `//  Config file for SPHUGO SPH Simulation
//  This is an example configuration.
//  Tube that is open to the right

[[Simulation]]
[Config]
NSteps              1000
Gamma               4.666
ParticleMass        100000.0
// A 2-D Vector just has 2 components separated by space(s)
Acceleration        0       0.05
DeltaTHalf          0.00424
//Kernel		    Monahan
Kernel				Wendtland

// Initial setup of particles, for now we can add Uniformely Random distributed Rectangels only
[[Start]]

[UniformRect]
NParticles          4000
UpperLeft           0.3     0.3
LowerRight          0.7     0.4

[UniformRect]
NParticles          700
UpperLeft           0.3    0.3
LowerRight          0.7    0.5

// Per default boundaries are open
[[Boundaries]]
[Reflection]
Left 0.2
Up 0.25
Down 0.5

// Sources with constant rate (unstable feature)
[[Sources]]
[Point]
//Pos               0.2     0.2
//Rate              100

// THIS IS NOT IMPLEMENTED! Has no effect!
// Coordinates of viewport for animation
[[Simulation]]
[Viewport]
UpperLeft           0       0
LowerRight          1       1
`
