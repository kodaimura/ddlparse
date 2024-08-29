package lexer

import (
	"errors"

	"github.com/kodaimura/ddlparse/internal/common"
)


type Lexer interface {
	Lex(ddl string) ([]string, error)
}

/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  Lex(): 
    Transform ddl (string) to tokens([]string). 
	And Remove sql comments.
	Return an ValidateError 
	 if the closing part of a multiline comment or string is not found.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

Example:

***** DDL *****
"CREATE TABLE IF NOT users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	password TEXT NOT NULL, --hashing
	created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
	updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
	UNIQUE(name)
);"

***** Tokens *****
[CREATE TABLE IF NOT users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

type lexer struct {
	Rdbms common.Rdbms
	ddlr []rune
	size int
	i int
	line int
	tokens []string
}


func NewLexer(rdbms common.Rdbms) Lexer {
	return &lexer{Rdbms: rdbms}
}


func (l *lexer) Lex(ddl string) ([]string, error) {
	l.init(ddl)
	if err := l.lex(); err != nil {
		return []string{}, err
	}
	return l.tokens, nil
}


func (l *lexer) init(ddl string) {
	l.ddlr = []rune(ddl)
	l.size = len(l.ddlr)
	l.i = 0
	l.line = 1
	l.tokens = []string{}
}


func (l *lexer) next() error {
	l.i += 1
	if l.isOutOfRange() {
		return errors.New("out of range")
	}
	return nil
}


func (l *lexer) char() string {
	return string(l.ddlr[l.i])
}


func (l *lexer) appendToken(token string) {
	if (token != "") {
		l.tokens = append(l.tokens, token)
	}
}


func (l *lexer) isOutOfRange() bool {
	return l.i > l.size - 1
}


func (l *lexer) lexError() error {
	if l.isOutOfRange() {
		return common.NewValidateError(l.line, string(l.ddlr[l.size - 1]))
	}
	return common.NewValidateError(l.line, string(l.ddlr[l.i]))
}


func (l *lexer) lex() error {
	token := ""
	for l.size > l.i {
		c := l.char()

		if c == "-" {
			if err := l.lexHyphen(&token); err != nil {
				return err
			}
		} else if c == "/" {
			if err := l.lexSlash(&token); err != nil {
				return err
			}
		} else if c == "*" {
			if err := l.lexAsterisk(&token); err != nil {
				return err
			}
		}

		if l.isOutOfRange() {
			break
		}

		c = l.char()
		if c == "\"" {
			if err := l.lexDoubleQuote(&token); err != nil {
				return err
			}
		} else if c == "'" {
			if err := l.lexSingleQuote(&token); err != nil {
				return err
			}
		} else if c == "`" {
			if err := l.lexBackQuote(&token); err != nil {
				return err
			}
		} else if c == "#" {
			l.lexSharp(&token)

		} else if c == " " || c == "\t"{
			l.lexSpace(&token)

		} else if c == "\n" {
			l.lexEOL(&token)

		} else if c == "(" || c == ")" || c == "," || c == "." || c == ";" {
			l.lexSymbol(&token)

		} else if c == "　" {
			return l.lexError()

		} else {
			token += c
			l.next()
		}
	}
	l.appendToken(token)
	return nil
}


func (l *lexer) lexHyphen(token *string) error {
	c := l.char()
	if c == "-" {
		if l.next() != nil {
			return l.lexError()
		}
		if l.char() == "-" {
			l.appendToken(*token)
			*token = ""
			l.skipComment()
		} else {
			*token += c
		}
	}

	return nil
}


func (l *lexer) lexSlash(token *string) error {
	c := l.char()
	if c == "/" {
		if l.next() != nil {
			return l.lexError()
		}
		if l.char() == "*" {
			l.appendToken(*token)
			*token = ""
			if err := l.skipMultiLineComment(); err != nil {
				return err
			}
		} else {
			*token += c
		}
	}
	return nil
}


func (l *lexer) lexAsterisk(token *string) error {
	c := l.char()
	if c == "*" {
		if l.next() != nil {
			return l.lexError()
		}
		if l.char() == "/" {
			l.i -= 1
			return l.lexError()
		} else {
			*token += c
		}
	} 
	return nil
}


func (l *lexer) lexDoubleQuote(token *string) error {
	c := l.char()
	if c == "\"" {
		l.appendToken(*token)
		*token = ""
		str, err := l.lexStringDoubleQuote()
		if err != nil {
			return err
		}
		l.appendToken(str)
	}
	return nil
}


func (l *lexer) lexSingleQuote(token *string) error {
	c := l.char()
	if c == "'" {
		l.appendToken(*token)
		*token = ""
		str, err := l.lexStringSingleQuote()
		if err != nil {
			return err
		}
		l.appendToken(str)
	}
	return nil
}


func (l *lexer) lexBackQuote(token *string) error {
	if l.Rdbms == common.PostgreSQL {
		return l.lexError()
	}
	c := l.char()
	if c == "`" {
		l.appendToken(*token)
		*token = ""
		str, err := l.lexStringBackQuote()
		if err != nil {
			return err
		}
		l.appendToken(str)
	}
	return nil
}


func (l *lexer) lexSharp(token *string) {
	c := l.char()
	if c == "#" {
		if l.Rdbms == common.MySQL {
			l.appendToken(*token)
			*token = ""
			l.skipComment()
		} else {
			*token += c
			l.next()
		}
	}
	return
}


func (l *lexer) lexEOL(token *string) {
	c := l.char()
	if c == "\n" {
		l.line += 1
		l.appendToken(*token)
		l.appendToken(c)
		*token = ""
	}
	l.next()
	return
}


func (l *lexer) lexSpace(token *string) {
	c := l.char()
	if c == " " || c == "\t" {
		l.appendToken(*token)
		*token = ""
	}
	l.next()
	return
}


func (l *lexer) lexSymbol(token *string) {
	c := l.char()
	if c == "(" || c == ")" || c == "," || c == "." || c == ";" {
		l.appendToken(*token)
		l.appendToken(c)
		*token = ""
	}
	l.next()
	return
}


func (l *lexer) skipComment() {
	l.next()
	for !l.isOutOfRange() {
		if l.char() == "\n" {
			l.line += 1
			l.appendToken("\n")
			break
		}
		l.next()
	}
	l.next()
	return
}


func (l *lexer) skipMultiLineComment() error {
	l.next()
	c := ""
	for !l.isOutOfRange() {
		c = l.char()
		if c == "\n" {
			l.line += 1
			l.appendToken("\n")
		} else if c == "*" {
			if l.next() != nil {
				return l.lexError()
			}
			if l.char() == "/" {
				l.next()
				return nil
			}
		} else if c == "/" {
			if l.next() != nil {
				return l.lexError()
			}
			if l.char() == "*" {
				return l.skipMultiLineComment()
			}
		}
		l.next()
	}
	return l.lexError()
}


func (l *lexer) lexStringDoubleQuote() (string, error) {
	l.next()
	str := "\""
	c := ""
	for !l.isOutOfRange() {
		c = l.char()
		if c == "\n" {
			l.line += 1
			l.appendToken("\n")
		} else if c == "\"" {
			l.next()
			return str + c, nil
		} else if c == "'" {
			s, err := l.lexStringSingleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else if c == "`" {
			s, err := l.lexStringBackQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else {
			str += c
		}
		l.next()
	}
	return str, l.lexError()
}


func (l *lexer) lexStringSingleQuote() (string, error) {
	l.next()
	str := "'"
	c := ""
	for !l.isOutOfRange() {
		c = l.char()
		if c == "\n" {
			l.line += 1
			l.appendToken("\n")
		} else if c == "'" {
			l.next()
			return str + c, nil			
		} else if c == "\"" {
			s, err := l.lexStringDoubleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else if c == "`" {
			s, err := l.lexStringBackQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else {
			str += c
		}
		l.next()
	}
	return str, l.lexError()
}


func (l *lexer) lexStringBackQuote() (string, error) {
	l.next()
	str := "`"
	c := ""
	for !l.isOutOfRange() {
		c = l.char()
		if c == "\n" {
			l.line += 1
			l.appendToken("\n")
		} else if c == "`" {
			l.next()
			return str + c, nil			
		} else if c == "\"" {
			s, err := l.lexStringDoubleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else if c == "'" {
			s, err := l.lexStringSingleQuote()
			str += s
			if err != nil {
				return str, err
			}
		} else {
			str += c
		}
		l.next()
	}
	return str, l.lexError()
}