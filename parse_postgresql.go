package ddlparse

import (
	"errors"
	"regexp"
	"strings"
)

type postgresqlParser struct {
	ddl string
	tokens []string
	validatedTokens []string
	size int
	i int
	line int
	flg bool
	tables []Table
}

func newPostgreSQLParser(ddl string) parser {
	return &postgresqlParser{ddl: ddl, ddlr: []rune(ddl)}
}


func (p *postgresqlParser) token() string {
	return p.tokens[p.i]
}


func (p *postgresqlParser) appendToken(token string) {
	if (token != "") {
		p.tokens = append(p.tokens, token)
	}
}


func (p *postgresqlParser) isOutOfRange() bool {
	return p.i > p.size - 1
}


/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  VALIDATE: 
    Check the syntax of DDL (tokens). 
    And eliminate unnecessary tokens during parsing.
	Return an ValidateError if validation fails.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

Example:

***** tokens *****
[CREATE TABLE IF NOT users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

***** validatedTokens *****
[CREATE TABLE users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

func (p *postgresqlParser) Validate() error {
	tokens, err := Tokenize(p.ddl, SQLite)
	if err != nil {
		return err
	}
	p.tokens = tokens
	return p.validate()
}


func (p *postgresqlParser) validate() error {
	p.initV()
	return p.validateProc()
}


func (p *postgresqlParser) initV() {
	p.validatedTokens = []string{}
	p.i = -1
	p.line = 1
	p.size = len(p.tokens)
	p.flg = false
	p.next()
}


func (p *postgresqlParser) flgOn() {
	p.flg = true
}


func (p *postgresqlParser) flgOff() {
	p.flg = false
}


func (p *postgresqlParser) next() error {
	if p.flg {
		p.validatedTokens = append(p.validatedTokens, p.token())
	}
	return p.nextAux()
}


func (p *postgresqlParser) nextAux() error {
	p.i += 1
	if (p.isOutOfRange()) {
		return errors.New("out of range")
	}
	if (p.token() == "\n") {
		p.line += 1
		return p.nextAux()
	} else {
		return nil
	}
}


func (p *postgresqlParser) matchKeyword(keywords ...string) bool {
	return contains(
		append(
			mapSlice(keywords, strings.ToLower), 
			mapSlice(keywords, strings.ToUpper)...,
		), p.token())
}


func (p *postgresqlParser) matchSymbol(symbols ...string) bool {
	return contains(symbols, p.token())
}


func (p *postgresqlParser) validateProc() error {
	if (p.isOutOfRange()) {
		return nil
	}
	if err := p.validateCreateTable(); err != nil {
		return err
	}
	return p.validateProc()
}


func (p *postgresqlParser) syntaxError() error {
	if p.isOutOfRange() {
		return NewValidateError(p.line, p.tokens[p.size - 1])
	}
	return NewValidateError(p.line, p.tokens[p.i])
}


func (p *postgresqlParser) validateKeyword(keywords ...string) error {
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if p.matchKeyword(keywords...) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	return p.syntaxError()
}


func (p *postgresqlParser) validateSymbol(symbols ...string) error {
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if p.matchSymbol(symbols...) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	return p.syntaxError()
}

func (p *postgresqlParser) validatePositiveInteger(symbols ...string) error {
	if !isPositiveIntegerToken(p.token()) {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return nil
}


func (p *postgresqlParser) isValidName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!contains(ReservedWords_SQLite, strings.ToUpper(name))
}


func (p *postgresqlParser) isValidQuotedName(name string) bool {
	return true
}


func (p *postgresqlParser) validateName() error {
	if isQuotedToken(p.token()) {
		if !p.isValidQuotedName(p.token()) {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
	} else {
		if !p.isValidName(p.token()) {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
	}

	return nil
}


func (p *postgresqlParser) validateCreateTable() error {
	p.flgOn()
	if err := p.validateKeyword("CREATE"); err != nil {
		return err
	}
	if err := p.validateKeyword("TABLE"); err != nil {
		return err
	}

	p.flgOff()
	if p.validateKeyword("IF") == nil {
		if err := p.validateKeyword("NOT"); err != nil {
			return err
		}
		if err := p.validateKeyword("EXISTS"); err != nil {
			return err
		}
	}

	p.flgOn()
	if err := p.validateTableName(); err != nil {
		return err
	}
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateColumns(); err != nil {
		return err
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	if err := p.validateTableOptions(); err != nil {
		return err
	}

	if p.matchSymbol(";") {
		if p.next() != nil {
			return nil
		}
	} else {
		return p.syntaxError()
	}

	return p.validateCreateTable()
}


func (p *postgresqlParser) validateTableName() error {
	if err := p.validateName(); err != nil {
		return err
	}
	if p.validateSymbol(".") == nil {
		if err := p.validateName(); err != nil {
			return err
		}
	}

	return nil
}


func (p *postgresqlParser) validateColumns() error {
	if err := p.validateColumn(); err != nil {
		return err
	}
	if p.validateSymbol(",") == nil {
		return p.validateColumns()
	}

	return nil
}


func (p *postgresqlParser) validateColumn() error {
	if p.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
		return p.validateTableConstraint()
	}
	if err := p.validateColumnName(); err != nil {
		return err
	}
	if err := p.validateColumnType(); err != nil {
		return err
	}
	if err := p.validateColumnConstraint(); err != nil {
		return err
	}
	
	return nil
}


func (p *postgresqlParser) validateColumnName() error {
	return p.validateName()
}


// Omitting data types is not supported.
func (p *postgresqlParser) validateColumnType() error {
	if p.matchKeyword("BIT", "CHARACTER") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchKeyword("VARYING") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateTypeDigitN(); err != nil {
			return err
		}
		return nil
	}

	if p.matchKeyword("VARBIT", "VARCHAR", "CHAR") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateTypeDigitN(); err != nil {
			return err
		}
		return nil
	}

	if p.matchKeyword("NUMERIC", "DECIMAL") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateTypeDigitPS(); err != nil {
			return err
		}
		return nil
	}

	if p.matchKeyword("DOUBLE") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("PRECISION"); err != nil {
			return err
		}
		return nil
	}

	// TODO
	//if p.matchKeyword("INTERVAL") {
	//}

	if p.matchKeyword("TIME", "TIMESTAMP") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateTypeDigitP(); err != nil {
			return err
		}
		if p.matchKeyword("WITH", "WITHOUT") {
			if p.next() != nil {
				return p.syntaxError()
			}
			if err := p.validateKeyword("TIME"); err != nil {
				return err
			}
			if err := p.validateKeyword("ZONE"); err != nil {
				return err
			}
		}
		return nil
	}

	if p.matchKeyword(...DataType_PostgreSQL) {
		return nil
	}

	return p.syntaxError()
}

// (number)
func (p *postgresqlParser) validateTypeDigitN() error {
	if p.matchSymbol("(") {
		if p.next() != nil {
			return p.syntaxError()
		}
	} else {
		return nil
	}

	if err := p.validatePositiveInteger(); err != nil {
		return err
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}

// (presision)
func (p *postgresqlParser) validateTypeDigitP() error {
	return p.validateTypeDigitN()
}

// (presision. scale)
func (p *postgresqlParser) validateTypeDigitPS() error {
	if p.matchSymbol("(") {
		if p.next() != nil {
			return p.syntaxError()
		}
	} else {
		return nil
	}

	if err := p.validatePositiveInteger(); err != nil {
		return err
	}
	if err := p.validateSymbol(","); err != nil {
		return err
	}
	if err := p.validatePositiveInteger(); err != nil {
		return err
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}

func (p *postgresqlParser) validateColumnConstraint() error {
	p.flgOff()
	if p.validateKeyword("CONSTRAINT") == nil {
		if err := p.validateName(); err != nil {
			return err
		}
	}
	p.flgOn()
	return p.validateColumnConstraintAux([]string{})
}


func (p *postgresqlParser) validateColumnConstraintAux(ls []string) error {
	if p.matchKeyword("PRIMARY") {
		if contains(ls, "PRIMARY") {
			return p.syntaxError()
		}
		if err := p.validateConstraintPrimaryKey(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "PRIMARY"))
	}

	if p.matchKeyword("NOT") {
		if contains(ls, "NOTNULL") || contains(ls, "NULL") {
			return p.syntaxError()
		}
		if err := p.validateConstraintNotNull(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "NOTNULL"))
	}

	if p.matchKeyword("NULL") {
		if contains(ls, "NOTNULL") || contains(ls, "NULL") {
			return p.syntaxError()
		}
		if err := p.validateConstraintNull(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "NULL"))
	}

	if p.matchKeyword("UNIQUE") {
		if contains(ls, "UNIQUE") {
			return p.syntaxError()
		}
		if err := p.validateConstraintUnique(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "UNIQUE"))
	}

	if p.matchKeyword("CHECK") {
		if contains(ls, "CHECK") {
			return p.syntaxError()
		}
		if err := p.validateConstraintCheck(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "CHECK"))
	}

	if p.matchKeyword("DEFAULT") {
		if contains(ls, "DEFAULT") {
			return p.syntaxError()
		}
		if err := p.validateConstraintDefault(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "DEFAULT"))
	}

	if p.matchKeyword("REFERENCES") {
		if contains(ls, "REFERENCES") {
			return p.syntaxError()
		}
		if err := p.validateConstraintForeignKey(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "REFERENCES"))
	}

	if p.matchKeyword("GENERATED", "AS") {
		if contains(ls, "GENERATED") {
			return p.syntaxError()
		}
		if err := p.validateConstraintGenerated(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "GENERATED"))
	}

	return nil
}


func (p *postgresqlParser) validateConstraintPrimaryKey() error {
	p.flgOn()
	if err := p.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateIndexParameters(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateConstraintNotNull() error {
	p.flgOn()
	if err := p.validateKeyword("NOT"); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateKeyword("NULL"); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateConstraintNull() error {
	p.flg Off()
	if err := p.validateKeyword("NULL"); err != nil {
		return err
	}
	p.flg On()
	return nil
}


func (p *postgresqlParser) validateConstraintUnique() error {
	p.flgOn()
	if err := p.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	if err := p.validateIndexParameters(); err != nil {
		return err
	}
	return nil
}


func (p *postgresqlParser) validateConstraintCheck() error {
	p.flgOff()
	if err := p.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	if p.matchKeyword("NO") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("INHERIT"); err != nil {
			return err
		}
	}

	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateConstraintDefault() error {
	p.flgOn()
	if err := p.validateKeyword("DEFAULT"); err != nil {
		return err
	}
	if p.matchSymbol("(") {
		if err := p.validateExpr(); err != nil {
			return err
		}
	} else {
		if err := p.validateLiteralValue(); err != nil {
			return err
		}
	}
	return nil
}


func (p *postgresqlParser) validateConstraintForeignKey() error {
	p.flgOff()
	if err := p.validateKeyword("REFERENCES"); err != nil {
		return err
	}
	if err := p.validateTableName(); err != nil {
		return err
	}
	if p.validateSymbol("(") == nil {
		if err := p.validateCommaSeparatedColumnNames(); err != nil {
			return err
		}
		if err := p.validateSymbol(")"); err != nil {
			return err
		}
	}
	if err := p.validateConstraintForeignKeyAux(); err != nil {
		return p.syntaxError()
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateConstraintForeignKeyAux() error {
	p.flgOff()
	if p.validateKeyword("ON") == nil {
		if err := p.validateKeyword("DELETE", "UPDATE"); err != nil {
			return err
		}
		if p.validateKeyword("SET") == nil {
			if err := p.validateKeyword("NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if p.validateKeyword("CASCADE", "RESTRICT") == nil {

		} else if p.validateKeyword("NO") == nil {
			if err := p.validateKeyword("ACTION"); err != nil {
				return err
			}
		} else {
			return p.syntaxError()
		}
		return p.validateConstraintForeignKeyAux()
	}

	if p.validateKeyword("MATCH") == nil {
		if err := p.validateKeyword("SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return p.validateConstraintForeignKeyAux()
	}

	if p.matchKeyword("NOT", "DEFERRABLE") {
		if p.matchKeyword("NOT") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateKeyword("DEFERRABLE"); err != nil {
			return err
		}
		if p.validateKeyword("INITIALLY") == nil {
			if err := p.validateKeyword("DEFERRED", "IMMEDIATE"); err != nil {
				return err
			}
		}
		return p.validateConstraintForeignKeyAux()
	}

	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateConstraintGenerated() error {
	p.flgOff()
	if err := p.validateKeyword("GENERATED"); err != nil {
		return err
	}

	if p.matchKeyword("ALWAYS") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("AS"); err != nil {
			return err
		}
		if p.matchKeyword("IDENTITY") {
			if p.next() != nil {
				return p.syntaxError()
			}
			if p.matchSymbol("(") {
				if err := p.validateExpr(); err != nil {
					return err
				}
			}
			p.flgOn()
			return nil
		} else if p.matchSymbol("(") {
			if err := p.validateExpr(); err != nil {
				return err
			}
			if err := p.validateKeyword("STORED"); err != nil {
				return err
			}
			p.flgOn()
			return nil
		} else {
			return p.syntaxError()
		}
	} else if p.matchKeyword("BY") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("DEFAULT"); err != nil {
			return err
		}
		if err := p.validateKeyword("AS"); err != nil {
			return err
		}
		if err := p.validateKeyword("IDENTITY"); err != nil {
			return err
		}
		if p.matchSymbol("(") {
			if err := p.validateExpr(); err != nil {
				return err
			}
		}
		p.flgOn()
		return nil
	} else if p.matchKeyword("AS") {
		if err := p.validateKeyword("AS"); err != nil {
			return err
		}
		if err := p.validateKeyword("IDENTITY"); err != nil {
			return err
		}
		if p.matchSymbol("(") {
			if err := p.validateExpr(); err != nil {
				return err
			}
		}
		p.flgOn()
		return nil
	}

	return p.syntaxError()
}


func (p *postgresqlParser) validateExpr() error {
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateExprAux(); err != nil {
		return err
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}


func (p *postgresqlParser) validateExprAux() error {
	if p.matchSymbol(")") {
		return nil
	}
	if p.matchSymbol("(") {
		if err := p.validateExpr(); err != nil {
			return err
		}
		return p.validateExprAux()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return p.validateExprAux()
}


func (p *postgresqlParser) validateIndexParameters() error {
	p.flgOff()
	if p.matchKeyword("INCLUDE") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateSymbol("("); err != nil {
			return err
		}
		if err := p.validateCommaSeparatedColumnNames(); err != nil {
			return p.syntaxError()
		}
		if err := p.validateSymbol(")"); err != nil {
			return err
		}
	}
	if p.matchKeyword("WITH") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateExpr(); err != nil {
			return err
		}
	}
	if p.matchKeyword("USING") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("INDEX"); err != nil {
			return err
		}
		if err := p.validateKeyword("TABLESPACE"); err != nil {
			return err
		}
		if err := p.validateName(); err != nil {
			return err
		}
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateLiteralValue() error {
	if isNumericToken(p.token()) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	if isQuotedToken(p.token()) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := p.validateKeyword(ls...); err != nil {
		return err
	}
	return nil
}


func (p *postgresqlParser) validateTableConstraint() error {
	p.flgOff()
	if p.validateKeyword("CONSTRAINT") == nil{
		if err := p.validateName(); err != nil {
			return err
		}
	}
	p.flgOn()
	return p.validateTableConstraintAux()
}


func (p *postgresqlParser) validateTableConstraintAux() error {
	if p.matchKeyword("PRIMARY") {
		return p.validateTablePrimaryKey()
	}

	if p.matchKeyword("UNIQUE") {
		return p.validateTableUnique()
	}

	if p.matchKeyword("CHECK") {
		return p.validateTableCheck()
	}

	if p.matchKeyword("FOREIGN") {
		return p.validateTableForeignKey()
	}

	if p.matchKeyword("EXCLUDE") {
		return p.validateTableExclude()
	}

	return p.syntaxError()
}


func (p *postgresqlParser) validateTablePrimaryKey() error {
	p.flgOn()
	if err := p.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateCommaSeparatedColumnNames(); err != nil {
		return p.syntaxError()
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateIndexParameters(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateTableUnique() error {
	p.flgOn()
	if err := p.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateCommaSeparatedColumnNames(); err != nil {
		return p.syntaxError()
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateIndexParameters(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateTableCheck() error {
	p.flgOff()
	if err := p.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateTableForeignKey() error {
	p.flgOff()
	if err := p.validateKeyword("FOREIGN"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateCommaSeparatedColumnNames(); err != nil {
		return p.syntaxError()
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	if err := p.validateConstraintForeignKey(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateTableExclude() error {
	p.flgOff()
	if err := p.validateKeyword("EXCLUDE"); err != nil {
		return err
	}
	if p.validateKeyword("USING") == nil{
		if err := p.validateName(); err != nil {
			return err
		}
	}
	if err := p.validateAux(); err != nil {
		return p.syntaxError()
	}
	if err := p.validateIndexParameters(); err != nil {
		return err
	}
	if p.validateKeyword("WHERE") == nil{
		if err := p.validateAux(); err != nil {
			return err
		}
	}
	p.flgOn()
	return nil
}


func (p *postgresqlParser) validateCommaSeparatedColumnNames() error {
	if err := p.validateColumnName(); err != nil {
		return err
	}
	if p.matchSymbol(",") {
		if p.next() != nil {
			return p.syntaxError()
		}
		return p.validateCommaSeparatedColumnNames()
	}
	return nil
}


func (p *postgresqlParser) validateTableOptions() error {
	p.flgOff()
	if p.matchKeyword(";") {
		p.flgOn()
		return nil
	}
	if p.matchSymbol(",") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	if err := p.validateTableOptionsAux(); err != nil {
		return err
	}
	return p.validateTableOptions()
}


func (p *postgresqlParser) validateTableOptionsAux() error {
	if p.matchKeyword("WITH") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateExpr; err != nil {
			return err
		}
		return nil
	}
	if p.matchKeyword("WITHOUT") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("OID"); err != nil {
			return err
		}
		return nil
	}
	if p.matchKeyword("TABLESPACE") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateName; err != nil {
			return err
		}
		return nil
	}
	p.flgOn()
	return p.syntaxError()
}


func (p *postgresqlParser) Parse() ([]Table, error) {
	var tables []Table
	return tables, nil
}

var DataType_PostgreSQL = []string{
	"BIGINT",
	"INT8",
	"BIGSERIAL",
	"SERIAL8",
	"BIT",
	"VARBIT"
	"BOOLEAN",
	"BOOL",
	"BOX",
	"BYTEA",
	"CHARACTER",
	"CHAR",
	"VARCHAR",
	"CIDR",
	"CIRCLE",
	"DATE",
	"FLOAT8",
	"INET",
	"INTEGER",
	"INT",
	"INT4",
	//"INTERVAL",
	"JSON",
	"JSONB",
	"LINE",
	"LSEG",
	"MACADDR",
	"MACADDR8",
	"MONEY",
	"NUMERIC",
	"DECIMAL",
	"PATH",
	"PG_LSN",
	"POINT",
	"POLYGON",
	"REAL",
	"FLOAT4",
	"SMALLINT",
	"INT2",
	"SMALLSERIAL",
	"SERIAL2",
	"SERIAL",
	"SERIAL4",
	"TEXT",
	"TIME",
	"TIMETZ",
	"TIMESTAMP",
	"TIMESTAMPTZ",
	"TSQUERY",
	"TSVECTOR",
	"TXID_SNAPSHOT",
	"UUID",
	"XML",
}

var ReservedWords_PostgreSQL = []string{
	"A",
	"ABS",
	"ADA",
	"ALIAS",
	"ALL",
	"ALLOCATE",
	"ANALYSE",
	"ANALYZE",
	"AND",
	"ANY",
	"ARE",
	"ARRAY",
	"AS",
	"ASC",
	"ASENSITIVE",
	"ASYMMETRIC",
	"ATOMIC",
	"ATTRIBUTE",
	"ATTRIBUTES",
	"AUTHORIZATION",
	"AVG",
	"BASE64",
	"BERNOULLI",
	"BETWEEN",
	"BINARY",
	"BITVAR",
	"BIT_LENGTH",
	"BLOB",
	"BOTH",
	"BREADTH",
	"C",
	"CALL",
	"CARDINALITY",
	"CASE",
	"CAST",
	"CATALOG",
	"CATALOG_NAME",
	"CEIL",
	"CEILING",
	"CHARACTERS",
	"CHARACTER_LENGTH",
	"CHARACTER_SET_CATALOG",
	"CHARACTER_SET_NAME",
	"CHARACTER_SET_SCHEMA",
	"CHAR_LENGTH",
	"CHECK",
	"CHECKED",
	"CLASS_ORIGIN",
	"CLOB",
	"COBOL",
	"COLLATE",
	"COLLATION",
	"COLLATION_CATALOG",
	"COLLATION_NAME",
	"COLLATION_SCHEMA",
	"COLLECT",
	"COLUMN",
	"COLUMN_NAME",
	"COMMAND_FUNCTION",
	"COMMAND_FUNCTION_CODE",
	"COMPLETION",
	"CONDITION",
	"CONDITION_NUMBER",
	"CONNECT",
	"CONNECTION_NAME",
	"CONSTRAINT",
	"CONSTRAINT_CATALOG",
	"CONSTRAINT_NAME",
	"CONSTRAINT_SCHEMA",
	"CONSTRUCTOR",
	"CONTAINS",
	"CONTINUE",
	"CONVERT",
	"CORR",
	"CORRESPONDING",
	"COUNT",
	"COVAR_POP",
	"COVAR_SAMP",
	"CREATE",
	"CROSS",
	"CUBE",
	"CUME_DIST",
	"CURRENT",
	"CURRENT_DATE",
	"CURRENT_DEFAULT_TRANSFORM_GROUP",
	"CURRENT_PATH",
	"CURRENT_ROLE",
	"CURRENT_TIME",
	"CURRENT_TIMESTAMP",
	"CURRENT_TRANSFORM_GROUP_FOR_TYPE",
	"CURRENT_USER",
	"CURSOR_NAME",
	"DATA",
	"DATE",
	"DATETIME_INTERVAL_CODE",
	"DATETIME_INTERVAL_PRECISION",
	"DEFAULT",
	"DEFERRABLE",
	"DEFINED",
	"DEGREE",
	"DENSE_RANK",
	"DEPTH",
	"DEREF",
	"DERIVED",
	"DESC",
	"DESCRIBE",
	"DESCRIPTOR",
	"DESTROY",
	"DESTRUCTOR",
	"DETERMINISTIC",
	"DIAGNOSTICS",
	"DISCONNECT",
	"DISPATCH",
	"DISTINCT",
	"DO",
	"DYNAMIC",
	"DYNAMIC_FUNCTION",
	"DYNAMIC_FUNCTION_CODE",
	"ELEMENT",
	"ELSE",
	"END",
	"END-EXEC",
	"EQUALS",
	"EVERY",
	"EXCEPT",
	"EXCEPTION",
	"EXCLUDE",
	"EXEC",
	"EXISTING",
	"EXP",
	"FALSE",
	"FILTER",
	"FINAL",
	"FLOOR",
	"FOLLOWING",
	"FOR",
	"FOREIGN",
	"FORTRAN",
	"FOUND",
	"FREE",
	"FREEZE",
	"FROM",
	"FULL",
	"FUSION",
	"G",
	"GENERAL",
	"GENERATED",
	"GET",
	"GO",
	"GOTO",
	"GRANT",
	"GROUP",
	"GROUPING",
	"HAVING",
	"HEX",
	"HIERARCHY",
	"HOST",
	"IDENTITY",
	"IGNORE",
	"ILIKE",
	"IMPLEMENTATION",
	"IN",
	"INDICATOR",
	"INFIX",
	"INITIALIZE",
	"INITIALLY",
	"INNER",
	"INSTANCE",
	"INSTANTIABLE",
	"INTERSECT",
	"INTERSECTION",
	"INTO",
	"IS",
	"ISNULL",
	"ITERATE",
	"JOIN",
	"K",
	"KEY_MEMBER",
	"KEY_TYPE",
	"LATERAL",
	"LEADING",
	"LEFT",
	"LENGTH",
	"LESS",
	"LIKE",
	"LIMIT",
	"LN",
	"LOCALTIME",
	"LOCALTIMESTAMP",
	"LOCATOR",
	"LOWER",
	"M",
	"MAP",
	"MATCHED",
	"MAX",
	"MEMBER",
	"MERGE",
	"MESSAGE_LENGTH",
	"MESSAGE_OCTET_LENGTH",
	"MESSAGE_TEXT",
	"METHOD",
	"MIN",
	"MOD",
	"MODIFIES",
	"MODIFY",
	"MODULE",
	"MORE",
	"MULTISET",
	"MUMPS",
	"NATURAL",
	"NCLOB",
	"NESTING",
	"NEW",
	"NORMALIZE",
	"NORMALIZED",
	"NOT",
	"NOTNULL",
	"NULL",
	"NULLABLE",
	"NUMBER",
	"OCTETS",
	"OCTET_LENGTH",
	"OFF",
	"OFFSET",
	"OLD",
	"ON",
	"ONLY",
	"OPEN",
	"OPERATION",
	"OPTIONS",
	"OR",
	"ORDER",
	"ORDERING",
	"ORDINALITY",
	"OTHERS",
	"OUTER",
	"OUTPUT",
	"OVER",
	"OVERLAPS",
	"OVERRIDING",
	"PAD",
	"PARAMETER",
	"PARAMETERS",
	"PARAMETER_MODE",
	"PARAMETER_NAME",
	"PARAMETER_ORDINAL_POSITION",
	"PARAMETER_SPECIFIC_CATALOG",
	"PARAMETER_SPECIFIC_NAME",
	"PARAMETER_SPECIFIC_SCHEMA",
	"PARTITION",
	"PASCAL",
	"PATH",
	"PERCENTILE_CONT",
	"PERCENTILE_DISC",
	"PERCENT_RANK",
	"PLACING",
	"PLI",
	"POSTFIX",
	"POWER",
	"PRECEDING",
	"PREFIX",
	"PREORDER",
	"PRIMARY",
	"PUBLIC",
	"RANGE",
	"RANK",
	"READS",
	"RECURSIVE",
	"REF",
	"REFERENCES",
	"REFERENCING",
	"REGR_AVGX",
	"REGR_AVGY",
	"REGR_COUNT",
	"REGR_INTERCEPT",
	"REGR_R2",
	"REGR_SLOPE",
	"REGR_SXX",
	"REGR_SXY",
	"REGR_SYY",
	"RESULT",
	"RETURN",
	"RETURNED_CARDINALITY",
	"RETURNED_LENGTH",
	"RETURNED_OCTET_LENGTH",
	"RETURNED_SQLSTATE",
	"RETURNING",
	"RIGHT",
	"ROLLUP",
	"ROUTINE",
	"ROUTINE_CATALOG",
	"ROUTINE_NAME",
	"ROUTINE_SCHEMA",
	"ROW_COUNT",
	"ROW_NUMBER",
	"SCALE",
	"SCHEMA_NAME",
	"SCOPE",
	"SCOPE_CATALOG",
	"SCOPE_NAME",
	"SCOPE_SCHEMA",
	"SECTION",
	"SELECT",
	"SELF",
	"SENSITIVE",
	"SERVER_NAME",
	"SESSION_USER",
	"SETS",
	"SIMILAR",
	"SIZE",
	"SOME",
	"SOURCE",
	"SPACE",
	"SPECIFIC",
	"SPECIFICTYPE",
	"SPECIFIC_NAME",
	"SQL",
	"SQLCODE",
	"SQLERROR",
	"SQLEXCEPTION",
	"SQLSTATE",
	"SQLWARNING",
	"SQRT",
	"STATE",
	"STATIC",
	"STDDEV_POP",
	"STDDEV_SAMP",
	"STRUCTURE",
	"STYLE",
	"SUBCLASS_ORIGIN",
	"SUBLIST",
	"SUBMULTISET",
	"SUM",
	"SYMMETRIC",
	"SYSTEM_USER",
	"TABLE",
	"TABLESAMPLE",
	"TABLE_NAME",
	"TERMINATE",
	"THAN",
	"THEN",
	"TIES",
	"TIMEZONE_HOUR",
	"TIMEZONE_MINUTE",
	"TO",
	"TOP_LEVEL_COUNT",
	"TRAILING",
	"TRANSACTIONS_COMMITTED",
	"TRANSACTIONS_ROLLED_BACK",
	"TRANSACTION_ACTIVE",
	"TRANSFORM",
	"TRANSFORMS",
	"TRANSLATE",
	"TRANSLATION",
	"TRIGGER_CATALOG",
	"TRIGGER_NAME",
	"TRIGGER_SCHEMA",
	"TRUE",
	"UESCAPE",
	"UNBOUNDED",
	"UNDER",
	"UNION",
	"UNIQUE",
	"UNNAMED",
	"UNNEST",
	"UPPER",
	"USAGE",
	"USER",
	"USER_DEFINED_TYPE_CATALOG",
	"USER_DEFINED_TYPE_CODE",
	"USER_DEFINED_TYPE_NAME",
	"USER_DEFINED_TYPE_SCHEMA",
	"USING",
	"VARIABLE",
	"VAR_POP",
	"VAR_SAMP",
	"VERBOSE",
	"WHEN",
	"WHENEVER",
	"WHERE",
	"WIDTH_BUCKET",
	"WINDOW",
	"WITH",
	"WITHIN",
	"XMLAGG",
	"XMLBINARY",
	"XMLCOMMENT",
	"XMLNAMESPACES",
}