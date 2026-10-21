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

type Token struct {
	TokenKind TokenKind
	Value     []byte
}

type Tokenizer struct {
	Input    string
	Lexic    *Lexic
	Position int
}

func tokenize(input string, lexic *Lexic) iter.Seq[*Token] {
	tokenizer := &Tokenizer{input, lexic, 0}
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
			return &Token{Punctuator, p.drop(len(s))}
		}
	}
	return nil
}

func (p *Tokenizer) operator() *Token {
	for _, s := range p.Lexic.Operators {
		if p.are(s) {
			return &Token{Operator, p.drop(len(s))}
		}
	}
	return nil
}

func (p *Tokenizer) keyword() *Token {
	for _, s := range p.Lexic.Keywords {
		if p.are(s+" ") || p.are(s+"(") {
			return &Token{Keyword, p.drop(len(s))}
		}
	}
	return nil
}

func (p *Tokenizer) literal() *Token {

	for _, s := range p.Lexic.Literals.String {
		if p.are(s) {
			length := p.whilent([]byte{s[0]}, 1)
			// drop the open quote
			p.drop(1)
			token := &Token{LiteralString, p.drop(length - 1)}
			p.drop(1)
			return token
		}
	}
	// can be end of expression
	for _, s := range p.Lexic.Literals.True {
		if p.are(s+" ") || p.are(s+")") {
			_ = p.drop(len(s))
			return &Token{LiteralBool, []byte{1}}
		}
	}
	// can be end of expression
	for _, s := range p.Lexic.Literals.Null {
		if p.are(s+" ") || p.are(s+")") {
			_ = p.drop(len(s))
			return &Token{LiteralBool, []byte{0}}
		}
	}
	// are numbers = literal
	digits := p.while(unicode.IsDigit)
	if digits > 0 {
		// convert bytes -> string -> number
		s := string(p.drop(digits))
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Println("Error: unable to convert", s, "to int:", err)
		}
		return &Token{LiteralInt, int32_to_bytes(i)}
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
		return &Token{Identifier, p.drop(length)}
	}
	return nil
}

func (p *Tokenizer) comment() *Token {
	for _, s := range p.Lexic.Comments.Line {
		if p.are(s) {
			length := p.whilent([]byte{'\n'}, 1)
			return &Token{Comment, p.drop(length)}
		}
	}
	return nil
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

func (p *Tokenizer) drop(count int) []byte {
	s := []byte(p.Input[p.Position : p.Position+count])
	p.Position += count
	return s
}

func (p *Tokenizer) trim() {
	// tokens can be split by spaces or even new lines
	for p.is(' ') || p.is('\n') {
		p.drop(1)
	}
}
