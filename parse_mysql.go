package ddlparse

import (
	"errors"
	"regexp"
	"strings"
)

type mysqlParser struct {
	ddl string
	tokens []string
	validatedTokens []string
	size int
	i int
	line int
	flg bool
	tables []Table
}

func newMySQLParser(ddl string) parser {
	return &mysqlParser{ddl: ddl}
}


func (p *mysqlParser) token() string {
	return p.tokens[p.i]
}


func (p *mysqlParser) isOutOfRange() bool {
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

func (p *mysqlParser) Validate() error {
	tokens, err := Tokenize(p.ddl, MySQL)
	if err != nil {
		return err
	}
	p.tokens = tokens
	return p.validate()
}


func (p *mysqlParser) validate() error {
	p.initV()
	return p.validateProc()
}


func (p *mysqlParser) initV() {
	p.validatedTokens = []string{}
	p.i = -1
	p.line = 1
	p.size = len(p.tokens)
	p.flg = false
	p.next()
}


func (p *mysqlParser) flgOn() {
	p.flg = true
}


func (p *mysqlParser) flgOff() {
	p.flg = false
}


func (p *mysqlParser) next() error {
	if p.flg {
		p.validatedTokens = append(p.validatedTokens, p.token())
	}
	return p.nextAux()
}


func (p *mysqlParser) nextAux() error {
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


func (p *mysqlParser) syntaxError() error {
	if p.isOutOfRange() {
		return NewValidateError(p.line, p.tokens[p.size - 1])
	}
	return NewValidateError(p.line, p.tokens[p.i])
}


func (p *mysqlParser) matchKeyword(keywords ...string) bool {
	return contains(
		append(
			mapSlice(keywords, strings.ToLower), 
			mapSlice(keywords, strings.ToUpper)...,
		), p.token())
}


func (p *mysqlParser) matchSymbol(symbols ...string) bool {
	return contains(symbols, p.token())
}


func (p *mysqlParser) isStringValue(token string) bool {
	tmp := token[0:1]
	return tmp == "\"" || tmp == "'"
}


func (p *mysqlParser) isIdentifier(token string) bool {
	return token[0:1] == "`"
}


func (p *mysqlParser) isValidName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!contains(ReservedWords_MySQL, strings.ToUpper(name))
}


func (p *mysqlParser) isValidQuotedName(name string) bool {
	return true
}


func (p *mysqlParser) validateKeyword(keywords ...string) error {
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


func (p *mysqlParser) validateSymbol(symbols ...string) error {
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


func (p *mysqlParser) validateName() error {
	if p.isIdentifier(p.token()) {
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


func (p *mysqlParser) validateTableName() error {
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


func (p *mysqlParser) validateColumnName() error {
	return p.validateName()
}


func (p *mysqlParser) validateBrackets() error {
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateBracketsAux(); err != nil {
		return err
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}


func (p *mysqlParser) validatePositiveInteger() error {
	if !isPositiveIntegerToken(p.token()) {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return nil
}


func (p *mysqlParser) validateBracketsAux() error {
	if p.matchSymbol(")") {
		return nil
	}
	if p.matchSymbol("(") {
		if err := p.validateBrackets(); err != nil {
			return err
		}
		return p.validateBracketsAux()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return p.validateBracketsAux()
}


func (p *mysqlParser) validateStringValue() error {
	if !p.isStringValue(p.token()) {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return nil
}


// (number)
func (p *mysqlParser) validateTypeDigitN() error {
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
func (p *mysqlParser) validateTypeDigitP() error {
	return p.validateTypeDigitN()
}

// (presision. scale)
func (p *mysqlParser) validateTypeDigitPS() error {
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
	if (p.matchSymbol(",")) {
		if err := p.validateSymbol(","); err != nil {
			return err
		}
		if err := p.validatePositiveInteger(); err != nil {
			return err
		}
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}


func (p *mysqlParser) validateProc() error {
	if (p.isOutOfRange()) {
		return nil
	}
	if err := p.validateCreateTable(); err != nil {
		return err
	}
	return p.validateProc()
}


func (p *mysqlParser) validateCreateTable() error {
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
	if (p.token() == ";") {
		if p.next() != nil {
			return nil
		}
	}

	return p.validateCreateTable()
}


func (p *mysqlParser) validateColumns() error {
	if err := p.validateColumn(); err != nil {
		return err
	}
	if p.validateSymbol(",") == nil {
		return p.validateColumns()
	}

	return nil
}


func (p *mysqlParser) validateColumn() error {
	if p.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "INDEX", "KEY", "FULLTEXT", "SPATIAL", "CHECK") {
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


// Omitting data types is not supported.
func (p *mysqlParser) validateColumnType() error {
	p.flgOn()
	if p.matchKeyword("VARCHAR", "CHAR", "BINARY", "VARBINARY", "BLOB", "TEXT") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateTypeDigitN(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("NUMERIC", "DECIMAL", "FLOAT", "REAL", "DOUBLE") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateTypeDigitPS(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("BIT", "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateTypeDigitP(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	// TODO
	//if p.matchKeyword("ENUM") {
	//}

	// TODO
	//if p.matchKeyword("SET") {
	//}

	if p.matchKeyword("TIME", "DATETIME", "TIMESTAMP", "YEAR") {
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
		p.flgOff()
		return nil
	}

	if p.matchKeyword(DataType_MySQL...) {
		if p.next() != nil {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}

	return p.syntaxError()
}


func (p *mysqlParser) validateColumnConstraint() error {
	p.flgOff()
	if p.validateKeyword("CONSTRAINT") == nil {
		if !p.matchKeyword("CHECK") {
			if err := p.validateName(); err != nil {
				return err
			}
		}
	}
	p.flgOn()
	return p.validateColumnConstraintAux([]string{})
}


func (p *mysqlParser) validateColumnConstraintAux(ls []string) error {
	if p.matchKeyword("PRIMARY", "KEY") {
		if contains(ls, "PRIMARY") {
			return p.syntaxError()
		}
		if err := p.validateConstraintPrimaryKey(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "PRIMARY"))
	}

	if p.matchKeyword("NOT") {
		if contains(ls, "NOT") {
			return p.syntaxError()
		}
		if err := p.validateConstraintNotNull(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "NOT"))
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

	if p.matchKeyword("COMMENT") {
		if contains(ls, "COMMENT") {
			return p.syntaxError()
		}
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateStringValue(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "COMMENT"))
	}

	if p.matchKeyword("COLLATE") {
		if contains(ls, "COLLATE") {
			return p.syntaxError()
		}
		if err := p.validateConstraintCollate(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "COLLATE"))
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

	if p.matchKeyword("COLUMN_FORMAT") {
		if contains(ls, p.token()) {
			return p.syntaxError()
		}
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("FIXED", "DYNAMIC", "DEFAULT"); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, p.token()))
	}

	if p.matchKeyword("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		if contains(ls, p.token()) {
			return p.syntaxError()
		}
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateStringValue(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, p.token()))
	}

	if p.matchKeyword("STORAGE") {
		if contains(ls, p.token()) {
			return p.syntaxError()
		}
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("DISK", "MEMORY"); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, p.token()))
	}

	if p.matchKeyword("VISIBLE", "INVISIBLE", "VIRTUAL", "STORED") {
		if contains(ls, p.token()) {
			return p.syntaxError()
		}
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		return p.validateColumnConstraintAux(append(ls, p.token()))
	}

	if p.matchKeyword("AUTO_INCREMENT") {
		if contains(ls, "AUTO_INCREMENT") {
			return p.syntaxError()
		}
		p.flgOn()
		if p.next() != nil {
			return p.syntaxError()
		}
		return p.validateColumnConstraintAux(append(ls, "AUTO_INCREMENT"))
	}

	return nil
}


func (p *mysqlParser) validateConstraintPrimaryKey() error {
	p.flgOn()
	if p.matchKeyword("KEY") {
		p.validatedTokens = append(p.validatedTokens, "PRIMARY")
		if p.next() != nil {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}
	if err := p.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateConstraintNotNull() error {
	p.flgOn()
	if err := p.validateKeyword("NOT"); err != nil {
		return err
	}
	if err := p.validateKeyword("NULL"); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateConstraintUnique() error {
	p.flgOn()
	if err := p.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	p.flgOff()
	if p.matchKeyword("KEY") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	return nil
}


func (p *mysqlParser) validateConstraintCheck() error {
	p.flgOff()
	if err := p.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	if p.matchKeyword("NOT") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	if p.matchKeyword("ENFORCED") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	p.flgOn()
	return nil
}


func (p *mysqlParser) validateConstraintDefault() error {
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


func (p *mysqlParser) validateConstraintCollate() error {
	p.flgOff()
	if err := p.validateKeyword("COLLATE"); err != nil {
		return err
	}
	if err := p.validateName(); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateConstraintColumnFormat() error {
	p.flgOff()
	if err := p.validateKeyword("COLUMN_FORMAT"); err != nil {
		return err
	}
	if err := p.validateKeyword("FIXED", "DYNAMIC", "DEFAULT"); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateConstraintForeignKey() error {
	p.flgOff()
	if err := p.validateKeyword("REFERENCES"); err != nil {
		return err
	}
	if err := p.validateTableName(); err != nil {
		return err
	}
	if err := p.validateIndexKeysOff(); err != nil {
		return err
	}
	if err := p.validateConstraintForeignKeyAux(); err != nil {
		return p.syntaxError()
	}
	p.flgOn()
	return nil
}


func (p *mysqlParser) validateConstraintForeignKeyAux() error {
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

	p.flgOn()
	return nil
}


func (p *mysqlParser) validateConstraintGenerated() error {
	p.flgOff()
	if p.validateKeyword("GENERATED") == nil {
		if err := p.validateKeyword("ALWAYS"); err != nil {
			return err
		}
	}
	if err := p.validateKeyword("AS"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}


func (p *mysqlParser) validateExpr() error {
	return p.validateBrackets()
}


func (p *mysqlParser) validateLiteralValue() error {
	if isNumericToken(p.token()) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	if p.isStringValue(p.token()) {
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


func (p *mysqlParser) validateTableConstraint() error {
	p.flgOff()
	if p.validateKeyword("CONSTRAINT") == nil{
		if !p.matchKeyword("PRIMARY", "UNIQUE", "FOREIGN", "CHECK") {
			if err := p.validateName(); err != nil {
				return err
			}
		}
	}
	p.flgOn()
	return p.validateTableConstraintAux()
}


func (p *mysqlParser) validateTableConstraintAux() error {
	if p.matchKeyword("PRIMARY") {
		return p.validateTablePrimaryKey()
	}

	if p.matchKeyword("UNIQUE") {
		return p.validateTableUnique()
	}

	if p.matchKeyword("FOREIGN") {
		return p.validateTableForeignKey()
	}

	if p.matchKeyword("CHECK") {
		return p.validateTableCheck()
	}

	if p.matchKeyword("FULLTEXT", "SPATIAL") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchKeyword("INDEX", "KEY") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if !p.matchSymbol("(") {
			if err := p.validateName(); err != nil {
				return err
			}
		}
		if err := p.validateIndexKeysOff(); err != nil {
			return err
		}
		if err := p.validateIndexOption(); err != nil {
			return err
		}
		return nil
	}

	if p.matchKeyword("INDEX", "KEY") {
		return p.validateTableIndex()
	}

	return p.syntaxError()
}


func (p *mysqlParser) validateTablePrimaryKey() error {
	p.flgOn()
	if err := p.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	p.flgOff()
	if p.matchKeyword("USING") {
		if err := p.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := p.validateIndexKeysOn(); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateIndexOption(); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateTableUnique() error {
	p.flgOn()
	if err := p.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	if p.matchKeyword("INDEX", "KEY") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	if !p.matchSymbol("(") {
		if err := p.validateName(); err != nil {
			return err
		}
	}
	p.flgOff()
	if p.matchKeyword("USING") {
		if err := p.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := p.validateIndexKeysOn(); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateIndexOption(); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateTableForeignKey() error {
	p.flgOff()
	if err := p.validateKeyword("FOREIGN"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	if !p.matchSymbol("(") {
		if err := p.validateName(); err != nil {
			return err
		}
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
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateTableCheck() error {
	return p.validateConstraintCheck()
}


func (p *mysqlParser) validateTableIndex() error {
	p.flgOff()
	if err := p.validateKeyword("INDEX", "KEY"); err != nil {
		return err
	}
	if !p.matchKeyword("USING") && !p.matchSymbol("(") {
		if err := p.validateName(); err != nil {
			return err
		}
	}
	if p.matchKeyword("USING") {
		if err := p.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := p.validateIndexKeysOff(); err != nil {
		return err
	}
	if err := p.validateIndexOption(); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateIndexKeysOn() error {
	p.flgOn()
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateIndexKeysOffAux(); err != nil {
		return p.syntaxError()
	}
	p.flgOn()
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	p.flgOff()
	return nil
}

func (p *mysqlParser) validateIndexKeysOnAux() error {
	p.flgOff()
	if p.matchSymbol("(") {
		if err := p.validateExpr(); err != nil {
			return err
		}
	} else {
		p.flgOn()
		if err := p.validateName(); err != nil {
			return err
		}
		p.flgOff()
		if err := p.validateTypeDigitN(); err != nil {
			return p.syntaxError()
		}
	}
	if p.matchKeyword("ASC", "DESC") {
		if err := p.next(); err != nil {
			return err
		}
	}
	if p.matchSymbol(",") {
		p.flgOn()
		if err := p.next(); err != nil {
			return err
		}
		p.flgOff()
		return p.validateIndexKeysOnAux()
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateIndexKeysOff() error {
	p.flgOff()
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateIndexKeysOffAux(); err != nil {
		return p.syntaxError()
	}
	p.flgOff()
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateIndexKeysOffAux() error {
	p.flgOff()
	if p.matchSymbol("(") {
		if err := p.validateExpr(); err != nil {
			return err
		}
	} else {
		if err := p.validateName(); err != nil {
			return err
		}
		if err := p.validateTypeDigitN(); err != nil {
			return p.syntaxError()
		}
	}
	if p.matchKeyword("ASC", "DESC") {
		if err := p.next(); err != nil {
			return err
		}
	}
	if p.matchSymbol(",") {
		if err := p.next(); err != nil {
			return err
		}
		return p.validateIndexKeysOffAux()
	}
	p.flgOff()
	return nil
}


func (p *mysqlParser) validateIndexType() error {
	p.flgOff()
	if err := p.validateKeyword("USING"); err != nil {
		return err
	}
	if err := p.validateKeyword("BTREE", "HASH"); err != nil {
		return err
	}
	p.flgOff()
	return nil
}

func (p *mysqlParser) validateIndexOption() error {
	p.flgOff()
	if p.matchKeyword("KEY_BLOCK_SIZE") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateLiteralValue(); err != nil {
			return err
		}
		return p.validateIndexOption()

	} else if p.matchKeyword("USING") {
		p.flgOff()
		if err := p.validateIndexType(); err != nil {
			return err
		}
		return p.validateIndexOption()
		
	} else if p.matchKeyword("WITH") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("PARSER"); err != nil {
			return err
		}
		if err := p.validateName(); err != nil {
			return err
		}
		return p.validateIndexOption()

	} else if p.matchKeyword("COMMENT") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateStringValue(); err != nil {
			return err
		}
		return p.validateIndexOption()

	} else if p.matchKeyword("VISIBLE", "INVISIBLE") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		return p.validateIndexOption()

	} else if p.matchKeyword("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateStringValue(); err != nil {
			return err
		}
		
		return p.validateIndexOption()

	}

	p.flgOff()
	return nil
}


func (p *mysqlParser) validateCommaSeparatedColumnNames() error {
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


func (p *mysqlParser) validateCommaSeparatedTableNames() error {
	if err := p.validateTableName(); err != nil {
		return err
	}
	if p.matchSymbol(",") {
		if p.next() != nil {
			return p.syntaxError()
		}
		return p.validateCommaSeparatedTableNames()
	}
	return nil
}


func (p *mysqlParser) validateTableOptions() error {
	p.flgOff()
	if p.matchKeyword(";") {
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


func (p *mysqlParser) validateTableOptionsAux() error {
	p.flgOff()
	if p.matchKeyword(
		"AUTOEXTEND_SIZE", "AUTO_INCREMENT", "AVG_ROW_LENGTH", 
		"KEY_BLOCK_SIZE", "MAX_ROWS", "MIN_ROWS", "STATS_SAMPLE_PAGES",
	) {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateLiteralValue(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword(
		"COMMENT", "ENGINE_ATTRIBUTE", "PASSWORD", 
		"SECONDARY_ENGINE_ATTRIBUTE", "CONNECTION",
	) {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateStringValue(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("DATA", "INDEX") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("DIRECTORY"); err != nil {
			return err
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateStringValue(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("TABLESPACE") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateName(); err != nil {
			return err
		}
		if p.matchKeyword("STORAGE") {
			if p.next() != nil {
				return p.syntaxError()
			}
			if p.validateKeyword("DISK", "MEMORY") != nil {
				return p.syntaxError()
			}
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("DEFAULT", "CHARACTER", "COLLATE") {
		p.flgOff()
		if p.matchKeyword("DEFAULT") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if p.matchKeyword("CHARACTER") {
			if p.next() != nil {
				return p.syntaxError()
			}
			if p.validateKeyword("SET") != nil {
				return p.syntaxError()
			}
		} else if p.matchKeyword("COLLATE") {
			if p.next() != nil {
				return p.syntaxError()
			}
		} else {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateName(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("ENGINE") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateName(); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("CHECKSUM", "DELAY_KEY_WRITE") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if (p.matchSymbol("0", "1")) {
			if p.next() != nil {
				return p.syntaxError()
			}
		} else {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("PACK_KEYS", "STATS_AUTO_RECALC", "STATS_PERSISTENT") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if (p.matchSymbol("0", "1") || p.matchKeyword("DEFAULT")) {
			if p.next() != nil {
				return p.syntaxError()
			}
		} else {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("COMPRESSION") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if (p.matchKeyword("'ZLIB'", "'LZ4'", "'NONE'")) {
			if p.next() != nil {
				return p.syntaxError()
			}
		} else {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("ENCRYPTION") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if (p.matchKeyword("'Y'", "'N'")) {
			if p.next() != nil {
				return p.syntaxError()
			}
		} else {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("INSERT_METHOD") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if (p.matchKeyword("NO", "FIRST", "LAST")) {
			if p.next() != nil {
				return p.syntaxError()
			}
		} else {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("ROW_FORMAT") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if (p.matchKeyword("DEFAULT", "DYNAMIC", "FIXED", "COMPRESSED", "REDUNDANT", "COMPACT")) {
			if p.next() != nil {
				return p.syntaxError()
			}
		} else {
			return p.syntaxError()
		}
		p.flgOff()
		return nil
	}

	if p.matchKeyword("UNION") {
		p.flgOff()
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol("=") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateSymbol("("); err != nil {
			return err
		}
		if err := p.validateCommaSeparatedTableNames(); err != nil {
			return p.syntaxError()
		}
		if err := p.validateSymbol(")"); err != nil {
			return err
		}
		p.flgOff()
		return nil
	}
	
	return p.syntaxError()
}


func (p *mysqlParser) Parse() ([]Table, error) {
	var tables []Table
	return tables, nil
}

var DataType_MySQL = []string{
	"SERIAL",
	"BOOL",
	"BOOLEAN",
	"INTEGER",
	"INT",
	"SMALLINT",
	"TINYINT",
	"MEDIUMINT",
	"BIGINT",
	"DECIMAL",
	"NUMERIC",
	"FLOAT",
	"DOUBLE",
	"BIT",
	"DATE",
	"DATETIME",
	"TIMESTAMP",
	"TIME",
	"YEAR",
	"CHAR",
	"VARCHAR",
	"BINARY",
	"VARBINARY",
	"BLOB",
	"TEXT",
	//"ENUM",
	//"SET",
	"GEOMETRY",
	"POINT",
	"LINESTRING",
	"POLYGON",
	"MULTIPOINT",
	"MULTILINESTRING",
	"MULTIPOLYGON",
	"GEOMETRYCOLLECTION",
	"JSON",
}

var ReservedWords_MySQL = []string{
	"AUTO_INCREMENT",
	"ACCESSIBLE",
	"ADD",
	"ALL",
	"ALTER",
	"ANALYZE",
	"AND",
	"ARRAY",
	"AS",
	"ASC",
	"ASENSITIVE",
	"BEFORE",
	"BETWEEN",
	"BIGINT",
	"BINARY",
	"BLOB",
	"BOTH",
	"BY",
	"CALL",
	"CASCADE",
	"CASE",
	"CHANGE",
	"CHAR",
	"CHARACTER",
	"CHECK",
	"COLLATE",
	"COLUMN",
	"CONDITION",
	"CONSTRAINT",
	"CONTINUE",
	"CONVERT",
	"CREATE",
	"CROSS",
	"CUBE",
	"CUME_DIST",
	"CURRENT_DATE",
	"CURRENT_TIME",
	"CURRENT_TIMESTAMP",
	"CURRENT_USER",
	"CURSOR",
	"DATABASE",
	"DATABASES",
	"DAY_HOUR",
	"DAY_MICROSECOND",
	"DAY_MINUTE",
	"DAY_SECOND",
	"DEC",
	"DECIMAL",
	"DECLARE",
	"DEFAULT",
	"DELAYED",
	"DELETE",
	"DENSE_RANK",
	"DESC",
	"DESCRIBE",
	"DETERMINISTIC",
	"DISTINCT",
	"DISTINCTROW",
	"DIV",
	"DOUBLE",
	"DROP",
	"DUAL",
	"EACH",
	"ELSE",
	"ELSEIF",
	"EMPTY",
	"ENCLOSED",
	"ESCAPED",
	"EXCEPT",
	"EXISTS",
	"EXIT",
	"EXPLAIN",
	"FALSE",
	"FETCH",
	"FIRST_VALUE",
	"FLOAT",
	"FLOAT4",
	"FLOAT8",
	"FOR",
	"FORCE",
	"FOREIGN",
	"FROM",
	"FULLTEXT",
	"FUNCTION",
	"GENERATED",
	"GET",
	"GRANT",
	"GROUP",
	"GROUPING",
	"GROUPS",
	"HAVING",
	"HIGH_PRIORITY",
	"HOUR_MICROSECOND",
	"HOUR_MINUTE",
	"HOUR_SECOND",
	"IF",
	"IGNORE",
	"IN",
	"INDEX",
	"INFILEx",
	"INNER",
	"INOUT",
	"INSENSITIVE",
	"INSERT",
	"INT",
	"INT1",
	"INT2",
	"INT3",
	"INT4",
	"INT8",
	"INTEGER",
	"INTERVAL",
	"INTO",
	"IO_AFTER_GTIDS",
	"IO_BEFORE_GTIDS",
	"IS",
	"ITERATE",
	"JOIN",
	"JSON_TABLE",
	"KEY",
	"KEYS",
	"KILL",
	"LAG",
	"LAST_VALUE",
	"LATERAL",
	"LEAD",
	"LEADING",
	"LEAVE",
	"LEFT",
	"LIKE",
	"LIMIT",
	"LINEAR",
	"LINES",
	"LOAD",
	"LOCALTIME",
	"LOCALTIMESTAMP",
	"LOCK",
	"LONG",
	"LONGBLOB",
	"LONGTEXT",
	"LOOP",
	"LOW_PRIORITY",
	"MASTER",
	"MASTER_BIND",
	"MASTER_SSL_VERIFY_SERVER_CERT",
	"MATCH",
	"MAXVALUE",
	"MEDIUMBLOB",
	"MEDIUMINT",
	"MEDIUMTEXT",
	"MEMBER",
	"MIDDLEINT",
	"MINUTE_MICROSECOND",
	"MINUTE_SECOND",
	"MOD",
	"MODIFIES",
	"NATURAL",
	"NOT",
	"NO_WRITE_TO_BINLOG",
	"NTH_VALUE",
	"NTILE",
	"NULL",
	"NUMERIC",
	"OF",
	"ON",
	"OPTIMIZE",
	"OPTIMIZER_COSTS",
	"OPTION",
	"OPTIONALLY",
	"OR",
	"ORDER",
	"OUT",
	"OUTER",
	"OUTFILE",
	"OVER",
	"PARTITION",
	"PERCENT_RANK",
	"PRECISION",
	"PRIMARY",
	"PROCEDURE",
	"PURGE",
	"RANGE",
	"RANK",
	"READ",
	"READS",
	"READ_WRITE",
	"REAL",
	"RECURSIVE",
	"REFERENCES",
	"REGEXP",
	"RELEASE",
	"RENAME",
	"REPEAT",
	"REPLACE",
	"REQUIRE",
	"RESIGNAL",
	"RESTRICT",
	"RETURN",
	"REVOKE",
	"RIGHT",
	"RLIKE",
	"ROW",
	"ROWS",
	"ROW_NUMBER",
	"SCHEMA",
	"SCHEMAS",
	"SECOND_MICROSECOND",
	"SELECT",
	"SENSITIVE",
	"SEPARATOR",
	"SET",
	"SHOW",
	"SIGNAL",
	"SMALLINT",
	"SPATIAL",
	"SPECIFIC",
	"SQL",
	"SQLEXCEPTION",
	"SQLSTATE",
	"SQLWARNING",
	"SQL_BIG_RESULT",
	"SQL_CALC_FOUND_ROWS",
	"SQL_SMALL_RESULT",
	"SSL",
	"STARTING",
	"STORED",
	"STRAIGHT_JOIN",
	"SYSTEM",
	"TABLE",
	"TERMINATED",
	"THEN",
	"TINYBLOB",
	"TINYINT",
	"TINYTEXT",
	"TO",
	"TRAILING",
	"TRIGGER",
	"TRUE",
	"UNDO",
	"UNION",
	"UNIQUE",
	"UNLOCK",
	"UNSIGNED",
	"UPDATE",
	"USAGE",
	"USE",
	"USING",
	"UTC_DATE",
	"UTC_TIME",
	"UTC_TIMESTAMP",
	"VALUES",
	"VARBINARY",
	"VARCHAR",
	"VARCHARACTER",
	"VARYING",
	"VIRTUAL",
	"WHEN",
	"WHERE",
	"WHILE",
	"WINDOW",
	"WITH",
	"WRITE",
	"XOR",
	"YEAR_MONTH",
	"ZEROFILL",
}