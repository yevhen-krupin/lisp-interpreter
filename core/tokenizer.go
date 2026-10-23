package core

import (
	"encoding/json"
	"fmt"
	"iter"
	"log"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type TokenKind string

const (
	Identifier    TokenKind = "Identifier"
	Operator      TokenKind = "Operator"
	Punctuator    TokenKind = "Punctuator"
	LiteralInt    TokenKind = "LiteralInt"
	LiteralString TokenKind = "LiteralString"
	LiteralBool   TokenKind = "LiteralBool"
	Keyword       TokenKind = "Keyword"
	Comment       TokenKind = "Comment"
)

type TokenStream []*Token

func (t Token) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"kind":  t.TokenKind,
		"value": value_to_any(t.Value),
	})
}

func (s TokenStream) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, e := range s {
		sb.WriteString(fmt.Sprintf("(%v,'%v')", e.TokenKind, value_to_any(e.Value)))
		if i < len(s)-1 {
			sb.WriteRune(',')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

func (s TokenStream) Diagnostic() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, e := range s {
		sb.WriteString(fmt.Sprintf("[%v, '%v', %v]", e.TokenKind, value_to_any(e.Value), e.Location))
		if i < len(s)-1 {
			sb.WriteRune(',')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

type Token struct {
	TokenKind TokenKind
	Value     []byte
	Location  Location
}

type Tokenizer struct {
	Input     string
	Lexic     *Lexic
	Position  int
	Line      int
	LineStart int
}

type Location struct {
	Line     int
	Position int
}

func (l Location) String() string {
	return fmt.Sprintf("%v:%v", l.Line, l.Position)
}

func tokenize(input string, lexic *Lexic) iter.Seq[*Token] {
	tokenizer := &Tokenizer{input, lexic, 0, 0, 0}
	return tokenizer.tokenize0()
}

func (tokenizer *Tokenizer) tokenize0() iter.Seq[*Token] {
	return func(yield func(*Token) bool) {
		last_pos := tokenizer.Position
		for tokenizer.in() {
			tokenizer.trim()
			log.Println("input left:", tokenizer.Input[tokenizer.Position:len(tokenizer.Input)])
			t := coalesce(
				tokenizer,
				(*Tokenizer).comment,
				(*Tokenizer).scope,
				(*Tokenizer).keyword,
				(*Tokenizer).operator,
				(*Tokenizer).literal,
				(*Tokenizer).identifier,
			)
			if t != nil {
				if !yield(t) {
					return
				}
			} else {
				if tokenizer.Position == last_pos {
					panic(fmt.Sprintln("tokenizer is stuck, unknown token", tokenizer.Input[tokenizer.Position:len(tokenizer.Input)]))
				}
			}
		}
	}
}

func (p *Tokenizer) scope() *Token {
	for _, s := range p.Lexic.Punctuators.All {
		if p.are(s) {
			return p.token(Punctuator, len(s))
		}
	}
	return nil
}

func (p *Tokenizer) operator() *Token {
	for _, s := range p.Lexic.Operators {
		if p.are(s) {
			return p.token(Operator, len(s))
		}
	}
	return nil
}

func (p *Tokenizer) keyword() *Token {
	for _, s := range p.Lexic.Keywords {
		if p.are(s+" ") || p.are(s+"\n") {
			// end of the word
			return p.token(Keyword, len(s))
		}
		for _, pp := range p.Lexic.Punctuators.All {
			// end of the word with punctuatore like `identifier,` or `identifier(` or `identifier)`
			if p.are(s + pp) {
				return p.token(Keyword, len(s))
			}

		}
	}
	return nil
}

func (p *Tokenizer) literal() *Token {
	for _, s := range p.Lexic.Literals.String {
		if p.are(s) {
			// skip open string marker and then look for close marker
			// skip 1 otherwise we find the same marker and it will be empty string
			// + 1 because we count both open and close marker in the length
			length := p.whilent([]byte{s[0]}, 1) + 1
			token := p.tokenF(LiteralString, length, func(v []byte) []byte {
				// drop the open and close marker
				// [' . . . . ]
				//  '[. . . .]'
				return v[1 : len(v)-1]
			})
			return token
		}
	}
	// can be end of expression
	for _, s := range p.Lexic.Literals.True {
		if p.are(s+" ") || p.are(s+")") {
			return p.tokenV(LiteralBool, len(s), []byte{1})
		}
	}
	// can be end of expression
	for _, s := range p.Lexic.Literals.Null {
		if p.are(s+" ") || p.are(s+")") {
			return p.tokenV(LiteralBool, len(s), []byte{0})
		}
	}
	// are numbers = literal
	digits := p.while(unicode.IsDigit)
	if digits > 0 {
		return p.tokenF(LiteralInt, digits, func(v []byte) []byte {
			// convert bytes -> string -> number
			i, err := strconv.Atoi(string(v))
			if err != nil {
				log.Println("Error: unable to convert", string(v), "to int:", err)
			}
			return int32_to_bytes(i)
		})
	}
	return nil
}

func (p *Tokenizer) identifier() *Token {
	// starts from letter, can contain letters or numbers or underscores
	letters := p.while(unicode.IsLetter)
	if letters > 0 {
		length := p.while(func(r rune) bool {
			return unicode.IsDigit(r) || unicode.IsLetter(r)
		})
		return p.token(Identifier, length)
	}
	return nil
}

func (p *Tokenizer) comment() *Token {
	for _, s := range p.Lexic.Comments.Line {
		if p.are(s) {
			length := p.whilent([]byte{'\n'}, 1)
			return p.token(Comment, length)
		}
	}
	return nil
}

func (p *Tokenizer) token(kind TokenKind, len int) *Token {
	location, value := p.drop(len)
	return &Token{kind, value, location}
}

func (p *Tokenizer) tokenV(kind TokenKind, len int, override []byte) *Token {
	location, _ := p.drop(len)
	return &Token{kind, override, location}
}

func (p *Tokenizer) tokenF(kind TokenKind, len int, convert func(v []byte) []byte) *Token {
	location, value := p.drop(len)
	return &Token{kind, convert(value), location}
}

func (p *Tokenizer) in() bool {
	return p.Position < len(p.Input)
}

func (p *Tokenizer) while(f func(rune) bool) int {
	i := p.Position
	for i < len(p.Input) && f(rune(p.Input[i])) == true {
		i += 1
	}
	return i - p.Position
}

func (p *Tokenizer) are(chars string) bool {
	for i := 0; i < len(chars) && p.Position+i < len(p.Input); i++ {
		if p.Input[p.Position+i] != chars[i] {
			return false
		}
	}
	return true
}

func (p *Tokenizer) whilent(chars []byte, skip int) int {
	i := p.Position + skip
	for i < len(p.Input) && slices.Index(chars, p.Input[i]) == -1 {
		i += 1
	}
	return i - p.Position
}

func (p *Tokenizer) isany(chars []byte) bool {
	return p.in() && slices.Contains(chars, p.Input[p.Position])
}

func (p *Tokenizer) is(char byte) bool {
	return p.in() && p.Input[p.Position] == char
}

func (p *Tokenizer) isnt(char byte) bool {
	return p.in() && !p.is(char)
}

func (p *Tokenizer) drop(count int) (Location, []byte) {
	location := Location{p.Line, p.Position - p.LineStart}
	value := []byte(p.Input[p.Position : p.Position+count])
	// each time when we see newline we increment the line and set line start into the next char after \n
	for i, c := range value {
		if c == '\n' {
			// position where drop started
			// i - position of \n inside of the dropped sequence
			// +1 in order to skip the \n itself otherwise the location within new line is shifted
			p.line(p.Position + i + 1)
		}
	}
	p.Position += count
	return location, value
}

func (p *Tokenizer) trim() {
	// tokens can be split by spaces
	for p.is(' ') {
		p.drop(1)
	}
	// or even new lines
	for p.is('\n') {
		p.drop(1)
	}
}

func (p *Tokenizer) line(start int) {
	p.Line += 1
	p.LineStart = start
}
