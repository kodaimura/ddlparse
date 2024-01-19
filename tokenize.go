package ddlparse

import (
	"errors"
)

/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  TOKENIZE: 
    Transform ddl (string) to tokens([]string). 
	And remove sql comments.
	Return an ValidateError 
	 if the closing part of a multiline comment or string is not found.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

Example:

***** ddl *****
"CREATE TABLE IF NOT users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	password TEXT NOT NULL, --hashing
	created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
	updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
	UNIQUE(name)
);"

***** tokens *****
[CREATE TABLE IF NOT users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

func tokenize (ddl string, rdbms Rdbms) ([]string, error) {
	lex := newLexer(ddl, rdbms)
	if err := lex.tokenize(); err != nil {
		return lex.tokens, err
	}
	return lex.tokens, nil
}

type lexer struct {
	rdbms Rdbms
	ddlr []rune
	tokens []string
	size int
	i int
	line int
}

func newLexer(ddl string, rdbms Rdbms) *lexer {
	return &lexer{ddlr: []rune(ddl), rdbms: rdbms}
}

func (p *lexer) next() error {
	p.i += 1
	if p.isOutOfRange() {
		return errors.New("out of range")
	}
	return nil
}

func (p *lexer) char() string {
	return string(p.ddlr[p.i])
}

func (p *lexer) appendToken(token string) {
	if (token != "") {
		p.tokens = append(p.tokens, token)
	}
}

func (p *lexer) isOutOfRange() bool {
	return p.i > p.size - 1
}

func (p *lexer) tokenizeError() error {
	if p.isOutOfRange() {
		return NewValidateError(p.line, string(p.ddlr[p.size - 1]))
	}
	return NewValidateError(p.line, string(p.ddlr[p.i]))
}

func (p *lexer) tokenize() error {
	p.initT()
	return p.tokenizeProc()
}

func (p *lexer) initT() {
	p.tokens = []string{}
	p.size = len(p.ddlr)
	p.i = 0
	p.line = 1
}

func (p *lexer) tokenizeProc() error {
	token := ""
	cur := ""
	for p.size > p.i {
		cur = p.char()

		if cur == "-" {
			if err := p.tokenizeHyphen(&token); err != nil {
				return err
			}
		} else if cur == "/" {
			if err := p.tokenizeSlash(&token); err != nil {
				return err
			}
		} else if cur == "*" {
			if err := p.tokenizeAsterisk(&token); err != nil {
				return err
			}
		} else if cur == "\"" {
			if err := p.tokenizeDoubleQuote(&token); err != nil {
				return err
			}
		} else if cur == "'" {
			if err := p.tokenizeSingleQuote(&token); err != nil {
				return err
			}
		} else if cur == "`" {
			if err := p.tokenizeBackQuote(&token); err != nil {
				return err
			}
		} else if cur == " " || cur == "\t"{
			p.tokenizeSpace(&token)
		} else if cur == "\n" {
			p.tokenizeEOL(&token)
		} else if cur == "(" || cur == ")" || cur == "," || cur == "." || cur == ";" {
			p.tokenizeSymbol(&token)
		} else if cur == "　" {
			return p.tokenizeError()
		} else {
			token += cur
			p.next()
		}
	}
	p.appendToken(token)
	return nil
}

func (p *lexer) tokenizeHyphen(token *string) error {
	c := p.char()
	if c == "-" {
		if p.next() != nil {
			return p.tokenizeError()
		}
		if p.char() == "-" {
			p.appendToken(*token)
			*token = ""
			p.skipComment()
		} else {
			*token += c
		}
		p.next()
	}

	return nil
}

func (p *lexer) tokenizeSlash(token *string) error {
	c := p.char()
	if c == "/" {
		if p.next() != nil {
			return p.tokenizeError()
		}
		if p.char() == "*" {
			p.appendToken(*token)
			*token = ""
			if err := p.skipMultiLineComment(); err != nil {
				return err
			}
		} else {
			*token += c
		}
		p.next()
	}
	return nil
}

func (p *lexer) tokenizeAsterisk(token *string) error {
	c := p.char()
	if c == "*" {
		if p.next() != nil {
			return p.tokenizeError()
		}
		if p.char() == "/" {
			p.i -= 1
			return p.tokenizeError()
		} else {
			*token += c
		}
		p.next()
	} 
	return nil
}

func (p *lexer) tokenizeDoubleQuote(token *string) error {
	c := p.char()
	if c == "\"" {
		if *token != "" {
			return p.tokenizeError()
		}
		str, err := p.tokenizeStringDoubleQuote()
		if err != nil {
			return err
		}
		p.appendToken(str)
	}
	return nil
}

func (p *lexer) tokenizeSingleQuote(token *string) error {
	c := p.char()
	if c == "'" {
		if *token != "" {
			return p.tokenizeError()
		}
		str, err := p.tokenizeStringSingleQuote()
		if err != nil {
			return err
		}
		p.appendToken(str)
	}
	p.next()
	return nil
}

func (p *lexer) tokenizeBackQuote(token *string) error {
	c := p.char()
	if c == "'" {
		if *token != "" {
			return p.tokenizeError()
		}
		str, err := p.tokenizeStringBackQuote()
		if err != nil {
			return err
		}
		p.appendToken(str)
	}
	p.next()
	return nil
}

func (p *lexer) tokenizeEOL(token *string) {
	c := p.char()
	if c == "\n" {
		p.line += 1
		p.appendToken(*token)
		p.appendToken(c)
		*token = ""
	}
	p.next()
	return
}

func (p *lexer) tokenizeSpace(token *string) {
	c := p.char()
	if c == " " || c == "\t" {
		p.appendToken(*token)
		*token = ""
	}
	p.next()
	return
}

func (p *lexer) tokenizeSymbol(token *string) {
	c := p.char()
	if c == "(" || c == ")" || c == "," || c == "." || c == ";" {
		p.appendToken(*token)
		p.appendToken(c)
		*token = ""
	}
	p.next()
	return
}

func (p *lexer) skipComment() {
	p.next()
	for !p.isOutOfRange() {
		if p.char() == "\n" {
			p.line += 1
			p.appendToken("\n")
			break
		}
		p.next()
	}
	return
}

func (p *lexer) skipMultiLineComment() error {
	p.next()
	c := ""
	for !p.isOutOfRange() {
		c = p.char()
		if c == "\n" {
			p.line += 1
			p.appendToken("\n")
		} else if c == "*" {
			if p.next() != nil {
				return p.tokenizeError()
			}
			if p.char() == "/" {
				p.next()
				return nil
			}
		} else if c == "/" {
			if p.next() != nil {
				return p.tokenizeError()
			}
			if p.char() == "*" {
				return p.skipMultiLineComment()
			}
		}
		p.next()
	}
	return p.tokenizeError()
}

func (p *lexer) tokenizeStringDoubleQuote() (string, error) {
	p.next()
	str := "\""
	c := ""
	for !p.isOutOfRange() {
		c = p.char()
		if c == "\"" {
			p.next()
			return str + c, nil
		} else if c == "'" {
			s, err := p.tokenizeStringSingleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else if c == "`" {
			s, err := p.tokenizeStringBackQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else {
			str += c
		}
		p.next()
	}
	return str, p.tokenizeError()
}

func (p *lexer) tokenizeStringSingleQuote() (string, error) {
	p.next()
	str := "'"
	c := ""
	for !p.isOutOfRange() {
		c = p.char()
		if c == "'" {
			p.next()
			return str + c, nil			
		} else if c == "\"" {
			s, err := p.tokenizeStringDoubleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else if c == "`" {
			s, err := p.tokenizeStringBackQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else {
			str += c
		}
		p.next()
	}
	return str, p.tokenizeError()
}

func (p *lexer) tokenizeStringBackQuote() (string, error) {
	p.next()
	str := "`"
	c := ""
	for !p.isOutOfRange() {
		c = p.char()
		if c == "`" {
			p.next()
			return str + c, nil			
		} else if c == "\"" {
			s, err := p.tokenizeStringDoubleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else if c == "'" {
			s, err := p.tokenizeStringSingleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else {
			str += c
		}
		p.next()
	}
	return str, p.tokenizeError()
}