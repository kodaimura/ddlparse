package ddlparse

import (
	"errors"
	"regexp"
	"strings"
)

type mysqlValidator struct {
	tokens []string
	validatedTokens []string
	size int
	i int
	line int
	flg bool
}

func newMySQLValidator(ddl string) validator {
	return &mysqlValidator{tokens: tokens}
}


func (v *mysqlValidator) token() string {
	return v .tokens[v .i]
}


func (v *mysqlValidator) isOutOfRange() bool {
	return v .i > v .size - 1
}


func (v *mysqlValidator) Validate() ([]string, error) {
	v.initV()
	if err := v.validate(); err != nil {
		return nil, err
	}
	return v.validatedTokens, nil
}


func (v *mysqlValidator) initV() {
	v .validatedTokens = []string{}
	v .i = -1
	v .line = 1
	v .size = len(v .tokens)
	v .flg = false
	v .next()
}


func (v *mysqlValidator) flgOn() {
	v .flg = true
}


func (v *mysqlValidator) flgOff() {
	v .flg = false
}


func (v *mysqlValidator) next() error {
	if v .flg {
		v .validatedTokens = append(v .validatedTokens, v .token())
	}
	return v .nextAux()
}


func (v *mysqlValidator) nextAux() error {
	v .i += 1
	if (v .isOutOfRange()) {
		return errors.New("out of range")
	}
	if (v .token() == "\n") {
		v .line += 1
		return v .nextAux()
	} else {
		return nil
	}
}


func (v *mysqlValidator) syntaxError() error {
	if v .isOutOfRange() {
		return NewValidateError(v .line, v .tokens[v .size - 1])
	}
	return NewValidateError(v .line, v .tokens[v .i])
}


func (v *mysqlValidator) matchKeyword(keywords ...string) bool {
	return contains(
		append(
			mapSlice(keywords, strings.ToLower), 
			mapSlice(keywords, strings.ToUpper)...,
		), v .token())
}


func (v *mysqlValidator) matchSymbol(symbols ...string) bool {
	return contains(symbols, v .token())
}


func (v *mysqlValidator) isStringValue(token string) bool {
	tmp := token[0:1]
	return tmp == "\"" || tmp == "'"
}


func (v *mysqlValidator) isIdentifier(token string) bool {
	return token[0:1] == "`"
}


func (v *mysqlValidator) isValidName(name string) bool {
	pattern := regexv .MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!contains(ReservedWords_MySQL, strings.ToUpper(name))
}


func (v *mysqlValidator) isValidQuotedName(name string) bool {
	return true
}


func (v *mysqlValidator) validateKeyword(keywords ...string) error {
	if (v .isOutOfRange()) {
		return v .syntaxError()
	}
	if v .matchKeyword(keywords...) {
		if v .next() != nil {
			return v .syntaxError()
		}
		return nil
	}
	return v .syntaxError()
}


func (v *mysqlValidator) validateSymbol(symbols ...string) error {
	if (v .isOutOfRange()) {
		return v .syntaxError()
	}
	if v .matchSymbol(symbols...) {
		if v .next() != nil {
			return v .syntaxError()
		}
		return nil
	}
	return v .syntaxError()
}


func (v *mysqlValidator) validateName() error {
	if v .isIdentifier(v .token()) {
		if !v .isValidQuotedName(v .token()) {
			return v .syntaxError()
		}
		if v .next() != nil {
			return v .syntaxError()
		}
	} else {
		if !v .isValidName(v .token()) {
			return v .syntaxError()
		}
		if v .next() != nil {
			return v .syntaxError()
		}
	}

	return nil
}


func (v *mysqlValidator) validateTableName() error {
	if err := v .validateName(); err != nil {
		return err
	}
	if v .validateSymbol(".") == nil {
		if err := v .validateName(); err != nil {
			return err
		}
	}

	return nil
}


func (v *mysqlValidator) validateColumnName() error {
	return v .validateName()
}


func (v *mysqlValidator) validateBrackets() error {
	if err := v .validateSymbol("("); err != nil {
		return err
	}
	if err := v .validateBracketsAux(); err != nil {
		return err
	}
	if err := v .validateSymbol(")"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validatePositiveInteger() error {
	if !isPositiveIntegerToken(v .token()) {
		return v .syntaxError()
	}
	if v .next() != nil {
		return v .syntaxError()
	}
	return nil
}


func (v *mysqlValidator) validateBracketsAux() error {
	if v .matchSymbol(")") {
		return nil
	}
	if v .matchSymbol("(") {
		if err := v .validateBrackets(); err != nil {
			return err
		}
		return v .validateBracketsAux()
	}
	if v .next() != nil {
		return v .syntaxError()
	}
	return v .validateBracketsAux()
}


func (v *mysqlValidator) validateStringValue() error {
	if !v .isStringValue(v .token()) {
		return v .syntaxError()
	}
	if v .next() != nil {
		return v .syntaxError()
	}
	return nil
}


// (number)
func (v *mysqlValidator) validateTypeDigitN() error {
	if v .matchSymbol("(") {
		if v .next() != nil {
			return v .syntaxError()
		}
	} else {
		return nil
	}

	if err := v .validatePositiveInteger(); err != nil {
		return err
	}
	if err := v .validateSymbol(")"); err != nil {
		return err
	}
	return nil
}

// (presision)
func (v *mysqlValidator) validateTypeDigitP() error {
	return v .validateTypeDigitN()
}

// (presision. scale)
func (v *mysqlValidator) validateTypeDigitPS() error {
	if v .matchSymbol("(") {
		if v .next() != nil {
			return v .syntaxError()
		}
	} else {
		return nil
	}

	if err := v .validatePositiveInteger(); err != nil {
		return err
	}
	if (v .matchSymbol(",")) {
		if err := v .validateSymbol(","); err != nil {
			return err
		}
		if err := v .validatePositiveInteger(); err != nil {
			return err
		}
	}
	if err := v .validateSymbol(")"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validate() error {
	if (v .isOutOfRange()) {
		return nil
	}
	if err := v .validateCreateTable(); err != nil {
		return err
	}
	return v .validate()
}


func (v *mysqlValidator) validateCreateTable() error {
	v .flgOn()
	if err := v .validateKeyword("CREATE"); err != nil {
		return err
	}
	if err := v .validateKeyword("TABLE"); err != nil {
		return err
	}

	v .flgOff()
	if v .validateKeyword("IF") == nil {
		if err := v .validateKeyword("NOT"); err != nil {
			return err
		}
		if err := v .validateKeyword("EXISTS"); err != nil {
			return err
		}
	}

	v .flgOn()
	if err := v .validateTableName(); err != nil {
		return err
	}
	if err := v .validateSymbol("("); err != nil {
		return err
	}
	if err := v .validateColumns(); err != nil {
		return err
	}
	if err := v .validateSymbol(")"); err != nil {
		return err
	}
	if err := v .validateTableOptions(); err != nil {
		return err
	}
	if (v .token() == ";") {
		if v .next() != nil {
			return nil
		}
	}

	return v .validateCreateTable()
}


func (v *mysqlValidator) validateColumns() error {
	if err := v .validateColumn(); err != nil {
		return err
	}
	if v .validateSymbol(",") == nil {
		return v .validateColumns()
	}

	return nil
}


func (v *mysqlValidator) validateColumn() error {
	if v .matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "INDEX", "KEY", "FULLTEXT", "SPATIAL", "CHECK") {
		return v .validateTableConstraint()
	}
	if err := v .validateColumnName(); err != nil {
		return err
	}
	if err := v .validateColumnType(); err != nil {
		return err
	}
	if err := v .validateColumnConstraint(); err != nil {
		return err
	}
	
	return nil
}


// Omitting data types is not supported.
func (v *mysqlValidator) validateColumnType() error {
	v .flgOn()
	if v .matchKeyword("VARCHAR", "CHAR", "BINARY", "VARBINARY", "BLOB", "TEXT") {
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateTypeDigitN(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("NUMERIC", "DECIMAL", "FLOAT", "REAL", "DOUBLE") {
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateTypeDigitPS(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("BIT", "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT") {
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateTypeDigitP(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	// TODO
	//if v .matchKeyword("ENUM") {
	//}

	// TODO
	//if v .matchKeyword("SET") {
	//}

	if v .matchKeyword("TIME", "DATETIME", "TIMESTAMP", "YEAR") {
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateTypeDigitP(); err != nil {
			return err
		}
		if v .matchKeyword("WITH", "WITHOUT") {
			if v .next() != nil {
				return v .syntaxError()
			}
			if err := v .validateKeyword("TIME"); err != nil {
				return err
			}
			if err := v .validateKeyword("ZONE"); err != nil {
				return err
			}
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword(DataType_MySQL...) {
		if v .next() != nil {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}

	return v .syntaxError()
}


func (v *mysqlValidator) validateColumnConstraint() error {
	v .flgOff()
	if v .validateKeyword("CONSTRAINT") == nil {
		if !v .matchKeyword("CHECK") {
			if err := v .validateName(); err != nil {
				return err
			}
		}
	}
	v .flgOn()
	return v .validateColumnConstraintAux([]string{})
}


func (v *mysqlValidator) validateColumnConstraintAux(ls []string) error {
	if v .matchKeyword("PRIMARY", "KEY") {
		if contains(ls, "PRIMARY") {
			return v .syntaxError()
		}
		if err := v .validateConstraintPrimaryKey(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "PRIMARY"))
	}

	if v .matchKeyword("NOT") {
		if contains(ls, "NOTNULL") || contains(ls, "NULL") {
			return v .syntaxError()
		}
		if err := v .validateConstraintNotNull(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "NOTNULL"))
	}

	if v .matchKeyword("NULL") {
		if contains(ls, "NOTNULL") || contains(ls, "NULL") {
			return v .syntaxError()
		}
		v .flgOff()
		if err := v .validateConstraintNull(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "NULL"))
	}

	if v .matchKeyword("UNIQUE") {
		if contains(ls, "UNIQUE") {
			return v .syntaxError()
		}
		if err := v .validateConstraintUnique(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "UNIQUE"))
	}

	if v .matchKeyword("CHECK") {
		if contains(ls, "CHECK") {
			return v .syntaxError()
		}
		if err := v .validateConstraintCheck(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "CHECK"))
	}

	if v .matchKeyword("DEFAULT") {
		if contains(ls, "DEFAULT") {
			return v .syntaxError()
		}
		if err := v .validateConstraintDefault(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "DEFAULT"))
	}

	if v .matchKeyword("COMMENT") {
		if contains(ls, "COMMENT") {
			return v .syntaxError()
		}
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateStringValue(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "COMMENT"))
	}

	if v .matchKeyword("COLLATE") {
		if contains(ls, "COLLATE") {
			return v .syntaxError()
		}
		if err := v .validateConstraintCollate(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "COLLATE"))
	}

	if v .matchKeyword("REFERENCES") {
		if contains(ls, "REFERENCES") {
			return v .syntaxError()
		}
		if err := v .validateConstraintForeignKey(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "REFERENCES"))
	}

	if v .matchKeyword("GENERATED", "AS") {
		if contains(ls, "GENERATED") {
			return v .syntaxError()
		}
		if err := v .validateConstraintGenerated(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, "GENERATED"))
	}

	if v .matchKeyword("COLUMN_FORMAT") {
		if contains(ls, v .token()) {
			return v .syntaxError()
		}
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateKeyword("FIXED", "DYNAMIC", "DEFAULT"); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, v .token()))
	}

	if v .matchKeyword("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		if contains(ls, v .token()) {
			return v .syntaxError()
		}
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateStringValue(); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, v .token()))
	}

	if v .matchKeyword("STORAGE") {
		if contains(ls, v .token()) {
			return v .syntaxError()
		}
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateKeyword("DISK", "MEMORY"); err != nil {
			return err
		}
		return v .validateColumnConstraintAux(append(ls, v .token()))
	}

	if v .matchKeyword("VISIBLE", "INVISIBLE", "VIRTUAL", "STORED") {
		if contains(ls, v .token()) {
			return v .syntaxError()
		}
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		return v .validateColumnConstraintAux(append(ls, v .token()))
	}

	if v .matchKeyword("AUTO_INCREMENT") {
		if contains(ls, "AUTO_INCREMENT") {
			return v .syntaxError()
		}
		v .flgOn()
		if v .next() != nil {
			return v .syntaxError()
		}
		return v .validateColumnConstraintAux(append(ls, "AUTO_INCREMENT"))
	}

	return nil
}


func (v *mysqlValidator) validateConstraintPrimaryKey() error {
	v .flgOn()
	if v .matchKeyword("KEY") {
		v .validatedTokens = append(v .validatedTokens, "PRIMARY")
		if v .next() != nil {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}
	if err := v .validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := v .validateKeyword("KEY"); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintNotNull() error {
	v .flgOn()
	if err := v .validateKeyword("NOT"); err != nil {
		return err
	}
	if err := v .validateKeyword("NULL"); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintNull() error {
	v .flgOff()
	if err := v .validateKeyword("NULL"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintUnique() error {
	v .flgOn()
	if err := v .validateKeyword("UNIQUE"); err != nil {
		return err
	}
	v .flgOff()
	if v .matchKeyword("KEY") {
		if v .next() != nil {
			return v .syntaxError()
		}
	}
	return nil
}


func (v *mysqlValidator) validateConstraintCheck() error {
	v .flgOff()
	if err := v .validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := v .validateExpr(); err != nil {
		return err
	}
	if v .matchKeyword("NOT") {
		if v .next() != nil {
			return v .syntaxError()
		}
	}
	if v .matchKeyword("ENFORCED") {
		if v .next() != nil {
			return v .syntaxError()
		}
	}
	v .flgOn()
	return nil
}


func (v *mysqlValidator) validateConstraintDefault() error {
	v .flgOn()
	if err := v .validateKeyword("DEFAULT"); err != nil {
		return err
	}
	if v .matchSymbol("(") {
		if err := v .validateExpr(); err != nil {
			return err
		}
	} else {
		if err := v .validateLiteralValue(); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validateConstraintCollate() error {
	v .flgOff()
	if err := v .validateKeyword("COLLATE"); err != nil {
		return err
	}
	if err := v .validateName(); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintColumnFormat() error {
	v .flgOff()
	if err := v .validateKeyword("COLUMN_FORMAT"); err != nil {
		return err
	}
	if err := v .validateKeyword("FIXED", "DYNAMIC", "DEFAULT"); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintForeignKey() error {
	v .flgOff()
	if err := v .validateKeyword("REFERENCES"); err != nil {
		return err
	}
	if err := v .validateTableName(); err != nil {
		return err
	}
	if err := v .validateIndexKeysOff(); err != nil {
		return err
	}
	if err := v .validateConstraintForeignKeyAux(); err != nil {
		return v .syntaxError()
	}
	v .flgOn()
	return nil
}


func (v *mysqlValidator) validateConstraintForeignKeyAux() error {
	v .flgOff()
	if v .validateKeyword("ON") == nil {
		if err := v .validateKeyword("DELETE", "UPDATE"); err != nil {
			return err
		}
		if v .validateKeyword("SET") == nil {
			if err := v .validateKeyword("NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if v .validateKeyword("CASCADE", "RESTRICT") == nil {

		} else if v .validateKeyword("NO") == nil {
			if err := v .validateKeyword("ACTION"); err != nil {
				return err
			}
		} else {
			return v .syntaxError()
		}
		return v .validateConstraintForeignKeyAux()
	}

	if v .validateKeyword("MATCH") == nil {
		if err := v .validateKeyword("SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return v .validateConstraintForeignKeyAux()
	}

	v .flgOn()
	return nil
}


func (v *mysqlValidator) validateConstraintGenerated() error {
	v .flgOff()
	if v .validateKeyword("GENERATED") == nil {
		if err := v .validateKeyword("ALWAYS"); err != nil {
			return err
		}
	}
	if err := v .validateKeyword("AS"); err != nil {
		return err
	}
	if err := v .validateExpr(); err != nil {
		return err
	}
	v .flgOn()
	return nil
}


func (v *mysqlValidator) validateExpr() error {
	return v .validateBrackets()
}


func (v *mysqlValidator) validateLiteralValue() error {
	if isNumericToken(v .token()) {
		if v .next() != nil {
			return v .syntaxError()
		}
		return nil
	}
	if v .isStringValue(v .token()) {
		if v .next() != nil {
			return v .syntaxError()
		}
		return nil
	}
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := v .validateKeyword(ls...); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableConstraint() error {
	v .flgOff()
	if v .validateKeyword("CONSTRAINT") == nil{
		if !v .matchKeyword("PRIMARY", "UNIQUE", "FOREIGN", "CHECK") {
			if err := v .validateName(); err != nil {
				return err
			}
		}
	}
	v .flgOn()
	return v .validateTableConstraintAux()
}


func (v *mysqlValidator) validateTableConstraintAux() error {
	if v .matchKeyword("PRIMARY") {
		return v .validateTablePrimaryKey()
	}

	if v .matchKeyword("UNIQUE") {
		return v .validateTableUnique()
	}

	if v .matchKeyword("FOREIGN") {
		return v .validateTableForeignKey()
	}

	if v .matchKeyword("CHECK") {
		return v .validateTableCheck()
	}

	if v .matchKeyword("FULLTEXT", "SPATIAL") {
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchKeyword("INDEX", "KEY") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if !v .matchSymbol("(") {
			if err := v .validateName(); err != nil {
				return err
			}
		}
		if err := v .validateIndexKeysOff(); err != nil {
			return err
		}
		if err := v .validateIndexOption(); err != nil {
			return err
		}
		return nil
	}

	if v .matchKeyword("INDEX", "KEY") {
		return v .validateTableIndex()
	}

	return v .syntaxError()
}


func (v *mysqlValidator) validateTablePrimaryKey() error {
	v .flgOn()
	if err := v .validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := v .validateKeyword("KEY"); err != nil {
		return err
	}
	v .flgOff()
	if v .matchKeyword("USING") {
		if err := v .validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v .validateIndexKeysOn(); err != nil {
		return err
	}
	v .flgOff()
	if err := v .validateIndexOption(); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateTableUnique() error {
	v .flgOn()
	if err := v .validateKeyword("UNIQUE"); err != nil {
		return err
	}
	if v .matchKeyword("INDEX", "KEY") {
		if v .next() != nil {
			return v .syntaxError()
		}
	}
	if !v .matchSymbol("(") {
		if err := v .validateName(); err != nil {
			return err
		}
	}
	v .flgOff()
	if v .matchKeyword("USING") {
		if err := v .validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v .validateIndexKeysOn(); err != nil {
		return err
	}
	v .flgOff()
	if err := v .validateIndexOption(); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateTableForeignKey() error {
	v .flgOff()
	if err := v .validateKeyword("FOREIGN"); err != nil {
		return err
	}
	if err := v .validateKeyword("KEY"); err != nil {
		return err
	}
	if !v .matchSymbol("(") {
		if err := v .validateName(); err != nil {
			return err
		}
	}
	if err := v .validateSymbol("("); err != nil {
		return err
	}
	if err := v .validateCommaSeparatedColumnNames(); err != nil {
		return v .syntaxError()
	}
	if err := v .validateSymbol(")"); err != nil {
		return err
	}
	if err := v .validateConstraintForeignKey(); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateTableCheck() error {
	return v .validateConstraintCheck()
}


func (v *mysqlValidator) validateTableIndex() error {
	v .flgOff()
	if err := v .validateKeyword("INDEX", "KEY"); err != nil {
		return err
	}
	if !v .matchKeyword("USING") && !v .matchSymbol("(") {
		if err := v .validateName(); err != nil {
			return err
		}
	}
	if v .matchKeyword("USING") {
		if err := v .validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v .validateIndexKeysOff(); err != nil {
		return err
	}
	if err := v .validateIndexOption(); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexKeysOn() error {
	v .flgOn()
	if err := v .validateSymbol("("); err != nil {
		return err
	}
	if err := v .validateIndexKeysOffAux(); err != nil {
		return v .syntaxError()
	}
	v .flgOn()
	if err := v .validateSymbol(")"); err != nil {
		return err
	}
	v .flgOff()
	return nil
}

func (v *mysqlValidator) validateIndexKeysOnAux() error {
	v .flgOff()
	if v .matchSymbol("(") {
		if err := v .validateExpr(); err != nil {
			return err
		}
	} else {
		v .flgOn()
		if err := v .validateName(); err != nil {
			return err
		}
		v .flgOff()
		if err := v .validateTypeDigitN(); err != nil {
			return v .syntaxError()
		}
	}
	if v .matchKeyword("ASC", "DESC") {
		if v .next() != nil {
			return v .syntaxError()
		}
	}
	if v .matchSymbol(",") {
		v .flgOn()
		if v .next() != nil {
			return v .syntaxError()
		}
		v .flgOff()
		return v .validateIndexKeysOnAux()
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexKeysOff() error {
	v .flgOff()
	if err := v .validateSymbol("("); err != nil {
		return err
	}
	if err := v .validateIndexKeysOffAux(); err != nil {
		return v .syntaxError()
	}
	v .flgOff()
	if err := v .validateSymbol(")"); err != nil {
		return err
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexKeysOffAux() error {
	v .flgOff()
	if v .matchSymbol("(") {
		if err := v .validateExpr(); err != nil {
			return err
		}
	} else {
		if err := v .validateName(); err != nil {
			return err
		}
		if err := v .validateTypeDigitN(); err != nil {
			return v .syntaxError()
		}
	}
	if v .matchKeyword("ASC", "DESC") {
		if v .next() != nil {
			return v .syntaxError()
		}
	}
	if v .matchSymbol(",") {
		if v .next() != nil {
			return v .syntaxError()
		}
		return v .validateIndexKeysOffAux()
	}
	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexType() error {
	v .flgOff()
	if err := v .validateKeyword("USING"); err != nil {
		return err
	}
	if err := v .validateKeyword("BTREE", "HASH"); err != nil {
		return err
	}
	v .flgOff()
	return nil
}

func (v *mysqlValidator) validateIndexOption() error {
	v .flgOff()
	if v .matchKeyword("KEY_BLOCK_SIZE") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateLiteralValue(); err != nil {
			return err
		}
		return v .validateIndexOption()

	} else if v .matchKeyword("USING") {
		v .flgOff()
		if err := v .validateIndexType(); err != nil {
			return err
		}
		return v .validateIndexOption()
		
	} else if v .matchKeyword("WITH") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateKeyword("PARSER"); err != nil {
			return err
		}
		if err := v .validateName(); err != nil {
			return err
		}
		return v .validateIndexOption()

	} else if v .matchKeyword("COMMENT") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateStringValue(); err != nil {
			return err
		}
		return v .validateIndexOption()

	} else if v .matchKeyword("VISIBLE", "INVISIBLE") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		return v .validateIndexOption()

	} else if v .matchKeyword("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateStringValue(); err != nil {
			return err
		}
		
		return v .validateIndexOption()

	}

	v .flgOff()
	return nil
}


func (v *mysqlValidator) validateCommaSeparatedColumnNames() error {
	if err := v .validateColumnName(); err != nil {
		return err
	}
	if v .matchSymbol(",") {
		if v .next() != nil {
			return v .syntaxError()
		}
		return v .validateCommaSeparatedColumnNames()
	}
	return nil
}


func (v *mysqlValidator) validateCommaSeparatedTableNames() error {
	if err := v .validateTableName(); err != nil {
		return err
	}
	if v .matchSymbol(",") {
		if v .next() != nil {
			return v .syntaxError()
		}
		return v .validateCommaSeparatedTableNames()
	}
	return nil
}


func (v *mysqlValidator) validateTableOptions() error {
	v .flgOff()
	if v .matchKeyword(";") {
		return nil
	}
	if v .matchSymbol(",") {
		if v .next() != nil {
			return v .syntaxError()
		}
	}
	if err := v .validateTableOptionsAux(); err != nil {
		return err
	}
	return v .validateTableOptions()
}


func (v *mysqlValidator) validateTableOptionsAux() error {
	v .flgOff()
	if v .matchKeyword(
		"AUTOEXTEND_SIZE", "AUTO_INCREMENT", "AVG_ROW_LENGTH", 
		"KEY_BLOCK_SIZE", "MAX_ROWS", "MIN_ROWS", "STATS_SAMPLE_PAGES",
	) {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateLiteralValue(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword(
		"COMMENT", "ENGINE_ATTRIBUTE", "PASSWORD", 
		"SECONDARY_ENGINE_ATTRIBUTE", "CONNECTION",
	) {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateStringValue(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("DATA", "INDEX") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateKeyword("DIRECTORY"); err != nil {
			return err
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateStringValue(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("TABLESPACE") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if err := v .validateName(); err != nil {
			return err
		}
		if v .matchKeyword("STORAGE") {
			if v .next() != nil {
				return v .syntaxError()
			}
			if v .validateKeyword("DISK", "MEMORY") != nil {
				return v .syntaxError()
			}
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("DEFAULT", "CHARACTER", "COLLATE") {
		v .flgOff()
		if v .matchKeyword("DEFAULT") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if v .matchKeyword("CHARACTER") {
			if v .next() != nil {
				return v .syntaxError()
			}
			if v .validateKeyword("SET") != nil {
				return v .syntaxError()
			}
		} else if v .matchKeyword("COLLATE") {
			if v .next() != nil {
				return v .syntaxError()
			}
		} else {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateName(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("ENGINE") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateName(); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("CHECKSUM", "DELAY_KEY_WRITE") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if (v .matchSymbol("0", "1")) {
			if v .next() != nil {
				return v .syntaxError()
			}
		} else {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("PACK_KEYS", "STATS_AUTO_RECALC", "STATS_PERSISTENT") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if (v .matchSymbol("0", "1") || v .matchKeyword("DEFAULT")) {
			if v .next() != nil {
				return v .syntaxError()
			}
		} else {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("COMPRESSION") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if (v .matchKeyword("'ZLIB'", "'LZ4'", "'NONE'")) {
			if v .next() != nil {
				return v .syntaxError()
			}
		} else {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("ENCRYPTION") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if (v .matchKeyword("'Y'", "'N'")) {
			if v .next() != nil {
				return v .syntaxError()
			}
		} else {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("INSERT_METHOD") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if (v .matchKeyword("NO", "FIRST", "LAST")) {
			if v .next() != nil {
				return v .syntaxError()
			}
		} else {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("ROW_FORMAT") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if (v .matchKeyword("DEFAULT", "DYNAMIC", "FIXED", "COMPRESSED", "REDUNDANT", "COMPACT")) {
			if v .next() != nil {
				return v .syntaxError()
			}
		} else {
			return v .syntaxError()
		}
		v .flgOff()
		return nil
	}

	if v .matchKeyword("UNION") {
		v .flgOff()
		if v .next() != nil {
			return v .syntaxError()
		}
		if v .matchSymbol("=") {
			if v .next() != nil {
				return v .syntaxError()
			}
		}
		if err := v .validateSymbol("("); err != nil {
			return err
		}
		if err := v .validateCommaSeparatedTableNames(); err != nil {
			return v .syntaxError()
		}
		if err := v .validateSymbol(")"); err != nil {
			return err
		}
		v .flgOff()
		return nil
	}
	
	return v .syntaxError()
}


func (v *mysqlValidator) Parse() ([]Table, error) {
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