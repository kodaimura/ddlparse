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

func newMySQLValidator(tokens []string) validator {
	return &mysqlValidator{tokens: tokens}
}


func (v *mysqlValidator) Validate() ([]string, error) {
	v.init()
	if err := v.validate(); err != nil {
		return nil, err
	}
	return v.validatedTokens, nil
}


func (v *mysqlValidator) init() {
	v.validatedTokens = []string{}
	v.i = -1
	v.line = 1
	v.size = len(v.tokens)
	v.flg = false
	v.next()
}


func (v *mysqlValidator) token() string {
	return v.tokens[v.i]
}


func (v *mysqlValidator) flgOn() {
	v.flg = true
}


func (v *mysqlValidator) flgOff() {
	v.flg = false
}


func (v *mysqlValidator) isOutOfRange() bool {
	return v.i > v.size - 1
}


func (v *mysqlValidator) next() error {
	if v.flg {
		v.validatedTokens = append(v.validatedTokens, v.token())
	}
	return v.nextAux()
}


func (v *mysqlValidator) nextAux() error {
	v.i += 1
	if (v.isOutOfRange()) {
		return errors.New("out of range")
	}
	if (v.token() == "\n") {
		v.line += 1
		return v.nextAux()
	} else {
		return nil
	}
}


func (v *mysqlValidator) syntaxError() error {
	if v.isOutOfRange() {
		return NewValidateError(v.line, v.tokens[v.size - 1])
	}
	return NewValidateError(v.line, v.tokens[v.i])
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
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!contains(ReservedWords_MySQL, strings.ToUpper(name))
}


func (v *mysqlValidator) isValidQuotedName(name string) bool {
	return true
}


func (v *mysqlValidator) validateKeyword(keywords ...string) error {
	if (v.isOutOfRange()) {
		return v.syntaxError()
	}
	if v.matchKeyword(keywords...) {
		if v.next() != nil {
			return v.syntaxError()
		}
		return nil
	}
	return v.syntaxError()
}


func (v *mysqlValidator) validateSymbol(symbols ...string) error {
	if (v.isOutOfRange()) {
		return v.syntaxError()
	}
	if v.matchSymbol(symbols...) {
		if v.next() != nil {
			return v.syntaxError()
		}
		return nil
	}
	return v.syntaxError()
}


func (v *mysqlValidator) validateName() error {
	if v.isIdentifier(v.token()) {
		if !v.isValidQuotedName(v.token()) {
			return v.syntaxError()
		}
		if v.next() != nil {
			return v.syntaxError()
		}
	} else {
		if !v.isValidName(v.token()) {
			return v.syntaxError()
		}
		if v.next() != nil {
			return v.syntaxError()
		}
	}

	return nil
}


func (v *mysqlValidator) validateTableName() error {
	if err := v.validateName(); err != nil {
		return err
	}
	if v.matchSymbol(".") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateName(); err != nil {
			return err
		}
	}

	return nil
}


func (v *mysqlValidator) validateColumnName() error {
	return v.validateName()
}


func (v *mysqlValidator) validateBrackets() error {
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateBracketsAux(); err != nil {
		return err
	}
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateBracketsAux() error {
	if v.matchSymbol(")") {
		return nil
	}
	if v.matchSymbol("(") {
		if err := v.validateBrackets(); err != nil {
			return err
		}
		return v.validateBracketsAux()
	}
	if v.next() != nil {
		return v.syntaxError()
	}
	return v.validateBracketsAux()
}


func (v *mysqlValidator) validatePositiveInteger() error {
	if !isPositiveIntegerToken(v.token()) {
		return v.syntaxError()
	}
	if v.next() != nil {
		return v.syntaxError()
	}
	return nil
}


func (v *mysqlValidator) validateStringValue() error {
	if !v.isStringValue(v.token()) {
		return v.syntaxError()
	}
	if v.next() != nil {
		return v.syntaxError()
	}
	return nil
}


func (v *mysqlValidator) validate() error {
	if (v.isOutOfRange()) {
		return nil
	}
	if err := v.validateCreateTable(); err != nil {
		return err
	}
	return v.validate()
}


func (v *mysqlValidator) validateCreateTable() error {
	v.flgOn()
	if err := v.validateKeyword("CREATE"); err != nil {
		return err
	}
	if err := v.validateKeyword("TABLE"); err != nil {
		return err
	}
	if err := v.validateIfNotExists(); err != nil {
		return err
	}
	if err := v.validateTableName(); err != nil {
		return err
	}
	if err := v.validateTableDefinition(); err != nil {
		return err
	}

	if v.matchSymbol(";") {
		v.flgOn()
		if v.next() != nil {
			return nil
		}
	}

	return v.validateCreateTable()
}


func (v *mysqlValidator) validateIfNotExists() error {
	if v.matchKeyword("IF") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("NOT"); err != nil {
			return err
		}
		if err := v.validateKeyword("EXISTS"); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validateTableDefinition() error {
	v.flgOn()
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateColumnDefinitions(); err != nil {
		return err
	}
	v.flgOn()
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	if err := v.validateTableOptions(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateColumnDefinitions() error {
	if err := v.validateColumnDefinition(); err != nil {
		return err
	}
	if v.matchSymbol(",") {
		v.flgOn()
		if v.next() != nil {
			return v.syntaxError()
		}
		return v.validateColumnDefinitions()
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateColumnDefinition() error {
	v.flgOn()
	if v.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "INDEX", "KEY", "FULLTEXT", "SPATIAL", "CHECK") {
		return v.validateTableConstraint()
	}
	if err := v.validateColumnName(); err != nil {
		return err
	}
	if err := v.validateColumnType(); err != nil {
		return err
	}
	if err := v.validateColumnConstraints(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateColumnType() error {
	v.flgOn()
	if v.matchKeyword("VARCHAR", "CHAR", "BINARY", "VARBINARY", "BLOB", "TEXT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitN(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchKeyword("NUMERIC", "DECIMAL", "FLOAT", "REAL", "DOUBLE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitPS(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchKeyword("BIT", "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitP(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchKeyword("TIME", "DATETIME", "TIMESTAMP", "YEAR") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitP(); err != nil {
			return err
		}
		v.flgOff()
		if v.matchKeyword("WITH", "WITHOUT") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateKeyword("TIME"); err != nil {
				return err
			}
			if err := v.validateKeyword("ZONE"); err != nil {
				return err
			}
		}
		v.flgOff()
		return nil
	}

	// TODO if v.matchKeyword("ENUM") {}
	// TODO if v.matchKeyword("SET") {}

	v.flgOn()
	if err := v.validateKeyword(DataType_MySQL...); err != nil {
		return err
	}

	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateTypeDate() error {
	v.flgOn()
	if err := v.validateKeyword("TIME", "DATETIME", "TIMESTAMP", "YEAR"); err != nil {
		return err
	}
	if err := v.validateTypeDigitP(); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("WITH", "WITHOUT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("TIME"); err != nil {
			return err
		}
		if err := v.validateKeyword("ZONE"); err != nil {
			return err
		}
	}
	if v.matchKeyword("ON") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("UPDATE"); err != nil {
			return err
		}
		if err := v.validateKeyword("CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"); err != nil {
			return err
		}
	}
	v.flgOff()
	return nil
}

// (number)
func (v *mysqlValidator) validateTypeDigitN() error {
	if v.matchSymbol("(") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validatePositiveInteger(); err != nil {
			return err
		}
		if err := v.validateSymbol(")"); err != nil {
			return err
		}
	} 
	return nil
}


// (presision)
func (v *mysqlValidator) validateTypeDigitP() error {
	return v.validateTypeDigitN()
}


// (presision. scale)
func (v *mysqlValidator) validateTypeDigitPS() error {
	if v.matchSymbol("(") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validatePositiveInteger(); err != nil {
			return err
		}
		if v.matchSymbol(",") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validatePositiveInteger(); err != nil {
				return err
			}
		}
		if err := v.validateSymbol(")"); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validateColumnConstraints() error {
	v.flgOn()
	if v.matchKeyword("CONSTRAINT") {
		if v.next() != nil {
			return nil
		}
		if !v.matchKeyword("CHECK") {
			if err := v.validateName(); err != nil {
				return err
			}
		}
	}
	return v.validateColumnConstraintsAux([]string{})
}


func (v *mysqlValidator) isColumnConstraint(token string) bool {
	return v.matchKeyword(
		"PRIMARY", "KEY", "NOT", "NULL", "UNIQUE", "CHECK", "DEFAULT", "COLLATE", "REFERENCES", 
		"GENERATED", "AS", "COMMENT", "COLUMN_FORMAT", "ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE", 
		"STORAGE", "VISIBLE", "INVISIBLE", "VIRTUAL", "STORED", "AUTO_INCREMENT",
	)
}


func (v *mysqlValidator) validateColumnConstraintsAux(ls []string) error {
	if !v.isColumnConstraint(v.token()) {
		v.flgOff()
		return nil
	} 
	if v.matchKeyword("NOT") {
		if contains(ls, "NULL") {
			return v.syntaxError()
		} 
		ls = append(ls, "NULL")
	} else if v.matchKeyword("PRIMARY", "KEY") {
		if contains(ls, "PRIMARY") {
			return v.syntaxError()
		} 
		ls = append(ls, "PRIMARY")
	} else if v.matchKeyword("GENERATED", "AS") {
		if contains(ls, "GENERATED") {
			return v.syntaxError()
		} 
		ls = append(ls, "GENERATED")
	} else {
		if contains(ls, strings.ToUpper(v.token())) {
			return v.syntaxError()
		} 
		ls = append(ls, strings.ToUpper(v.token()))
	}

	if err := v.validateColumnConstraint(); err != nil {
		return err
	}

	return v.validateColumnConstraintsAux(ls)
}


func (v *mysqlValidator) validateColumnConstraint() error {
	if v.matchKeyword("PRIMARY", "KEY") {
		return v.validateConstraintPrimaryKey()
	}
	if v.matchKeyword("NOT") {
		return v.validateConstraintNotNull()
	}
	if v.matchKeyword("NULL") {
		return v.validateConstraintNull()
	}
	if v.matchKeyword("UNIQUE") {
		return v.validateConstraintUnique()
	}
	if v.matchKeyword("CHECK") {
		return v.validateConstraintCheck()
	}
	if v.matchKeyword("DEFAULT") {
		return v.validateConstraintDefault()
	}
	if v.matchKeyword("COLLATE") {
		return v.validateConstraintCollate()
	}
	if v.matchKeyword("REFERENCES") {
		return v.validateConstraintReferences()
	}
	if v.matchKeyword("GENERATED", "AS") {
		return v.validateConstraintGenerated()
	}
	if v.matchKeyword("COMMENT") {
		return v.validateConstraintComment()
	}
	if v.matchKeyword("COLUMN_FORMAT") {
		return v.validateConstraintColumnFormat()
	}
	if v.matchKeyword("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		return v.validateConstraintEngineAttribute()
	}
	if v.matchKeyword("STORAGE") {
		return v.validateConstraintStorage()
	}
	if v.matchKeyword("AUTO_INCREMENT") {
		return v.validateConstraintAutoincrement()
	}
	if v.matchKeyword("VISIBLE", "INVISIBLE", "VIRTUAL", "STORED") {
		v.flgOff()
		if v.next() != nil {
			return v.syntaxError()
		}
		return nil
	}
	
	return v.syntaxError()
}


func (v *mysqlValidator) validateConstraintPrimaryKey() error {
	v.flgOn()
	if v.matchKeyword("KEY") {
		v.validatedTokens = append(v.validatedTokens, "PRIMARY")
		if v.next() != nil {
			return v.syntaxError()
		}
		v.flgOff()
		return nil
	}
	if err := v.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateKeyword("KEY"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintNotNull() error {
	v.flgOn()
	if err := v.validateKeyword("NOT"); err != nil {
		return err
	}
	if err := v.validateKeyword("NULL"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintNull() error {
	v.flgOff()
	if err := v.validateKeyword("NULL"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintUnique() error {
	v.flgOn()
	if err := v.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("KEY") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	return nil
}


func (v *mysqlValidator) validateConstraintCheck() error {
	v.flgOn()
	if err := v.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("NOT") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if v.matchKeyword("ENFORCED") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	return nil
}


func (v *mysqlValidator) validateConstraintDefault() error {
	v.flgOn()
	if err := v.validateKeyword("DEFAULT"); err != nil {
		return err
	}
	if v.matchSymbol("(") {
		if err := v.validateExpr(); err != nil {
			return err
		}
	} else {
		if err := v.validateLiteralValue(); err != nil {
			return err
		}
		v.flgOff()
		if v.matchKeyword("ON") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateKeyword("UPDATE"); err != nil {
				return err
			}
			if err := v.validateKeyword("CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"); err != nil {
				return err
			}
		}
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintCollate() error {
	v.flgOn()
	if err := v.validateKeyword("COLLATE"); err != nil {
		return err
	}
	if err := v.validateName(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintReferences() error {
	v.flgOn()
	if err := v.validateKeyword("REFERENCES"); err != nil {
		return err
	}
	if err := v.validateTableName(); err != nil {
		return err
	}
	if v.matchSymbol("(") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateColumnName(); err != nil {
			return err
		}
		if err := v.validateSymbol(")"); err != nil {
			return err
		}
	}
	if err := v.validateConstraintReferencesAux(); err != nil {
		return v.syntaxError()
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintReferencesAux() error {
	v.flgOff()
	if v.matchKeyword("ON") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("DELETE", "UPDATE"); err != nil {
			return err
		}
		if v.matchKeyword("SET") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateKeyword("NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if v.matchKeyword("CASCADE", "RESTRICT") {
			if v.next() != nil {
				return v.syntaxError()
			}
		} else if v.matchKeyword("NO") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateKeyword("ACTION"); err != nil {
				return err
			}
		} else {
			return v.syntaxError()
		}
		return v.validateConstraintReferencesAux()
	}

	if v.matchKeyword("MATCH") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return v.validateConstraintReferencesAux()
	}

	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintGenerated() error {
	v.flgOff()
	if v.matchKeyword("GENERATED") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("ALWAYS"); err != nil {
			return err
		}
	}
	if err := v.validateKeyword("AS"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintComment() error {
	v.flgOff()
	if err := v.validateKeyword("COMMENT"); err != nil {
		return err
	}
	if err := v.validateStringValue(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintColumnFormat() error {
	v.flgOff()
	if err := v.validateKeyword("COLUMN_FORMAT"); err != nil {
		return err
	}
	if err := v.validateKeyword("FIXED", "DYNAMIC", "DEFAULT"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintEngineAttribute() error {
	v.flgOff()
	if err := v.validateKeyword("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE"); err != nil {
		return err
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateStringValue(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintStorage() error {
	v.flgOff()
	if err := v.validateKeyword("STORAGE"); err != nil {
		return err
	}
	if err := v.validateKeyword("DISK", "MEMORY"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintAutoincrement() error {
	v.flgOn()
	if err := v.validateKeyword("AUTO_INCREMENT"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateExpr() error {
	return v.validateBrackets()
}


func (v *mysqlValidator) validateLiteralValue() error {
	if isNumericToken(v.token()) {
		if v.next() != nil {
			return v.syntaxError()
		}
		return nil
	}
	if v.isStringValue(v.token()) {
		if v.next() != nil {
			return v.syntaxError()
		}
		return nil
	}
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := v.validateKeyword(ls...); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableConstraint() error {
	v.flgOn()
	if v.matchKeyword("CONSTRAINT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if !v.matchKeyword("PRIMARY", "UNIQUE", "FOREIGN", "CHECK") {
			if err := v.validateName(); err != nil {
				return err
			}
		}
	}
	return v.validateTableConstraintAux()
}


func (v *mysqlValidator) validateTableConstraintAux() error {
	if v.matchKeyword("PRIMARY") {
		return v.validateTableConstraintPrimaryKey()
	}
	if v.matchKeyword("UNIQUE") {
		return v.validateTableConstraintUnique()
	}
	if v.matchKeyword("FOREIGN") {
		return v.validateTableConstraintForeignKey()
	}
	if v.matchKeyword("CHECK") {
		return v.validateTableConstraintCheck()
	}
	if v.matchKeyword("INDEX", "KEY") {
		return v.validateTableConstraintIndex()
	}
	if v.matchKeyword("FULLTEXT", "SPATIAL") {
		v.flgOff()
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchKeyword("INDEX", "KEY") {
			if v.next() != nil {
				return v.syntaxError()
			}
		}
		if !v.matchSymbol("(") {
			if err := v.validateName(); err != nil {
				return err
			}
		}
		if err := v.validateIndexKeysOff(); err != nil {
			return err
		}
		if err := v.validateIndexOption(); err != nil {
			return err
		}
		return nil
	}

	return v.syntaxError()
}


func (v *mysqlValidator) validateTableConstraintPrimaryKey() error {
	v.flgOn()
	if err := v.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateKeyword("KEY"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v.validateIndexKeysOn(); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateIndexOption(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateTableConstraintUnique() error {
	v.flgOn()
	if err := v.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("INDEX", "KEY") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if !v.matchSymbol("(") {
		if err := v.validateName(); err != nil {
			return err
		}
	}
	v.flgOff()
	if v.matchKeyword("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v.validateIndexKeysOn(); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateIndexOption(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateTableConstraintForeignKey() error {
	v.flgOn()
	if err := v.validateKeyword("FOREIGN"); err != nil {
		return err
	}
	if err := v.validateKeyword("KEY"); err != nil {
		return err
	}
	if !v.matchSymbol("(") {
		v.flgOff()
		if err := v.validateName(); err != nil {
			return err
		}
	}
	v.flgOn()
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	if err := v.validateKeyword("REFERENCES"); err != nil {
		return err
	}
	if err := v.validateTableName(); err != nil {
		return err
	}
	if err := v.validateIndexKeysOn(); err != nil {
		return v.syntaxError()
	}
	v.flgOff()
	if err := v.validateConstraintReferencesAux(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateTableConstraintCheck() error {
	return v.validateConstraintCheck()
}


func (v *mysqlValidator) validateTableConstraintIndex() error {
	v.flgOff()
	if err := v.validateKeyword("INDEX", "KEY"); err != nil {
		return err
	}
	if !v.matchKeyword("USING") && !v.matchSymbol("(") {
		if err := v.validateName(); err != nil {
			return err
		}
	}
	if v.matchKeyword("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v.validateIndexKeysOff(); err != nil {
		return err
	}
	if err := v.validateIndexOption(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexKeysOn() error {
	v.flgOn()
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateIndexKeysOnAux(); err != nil {
		return v.syntaxError()
	}
	v.flgOn()
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}

func (v *mysqlValidator) validateIndexKeysOnAux() error {
	v.flgOff()
	if v.matchSymbol("(") {
		if err := v.validateExpr(); err != nil {
			return err
		}
	} else {
		v.flgOn()
		if err := v.validateName(); err != nil {
			return err
		}
		v.flgOff()
		if err := v.validateTypeDigitN(); err != nil {
			return v.syntaxError()
		}
	}
	if v.matchKeyword("ASC", "DESC") {
		v.flgOff()
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if v.matchSymbol(",") {
		v.flgOn()
		if v.next() != nil {
			return v.syntaxError()
		}
		v.flgOff()
		return v.validateIndexKeysOnAux()
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexKeysOff() error {
	v.flgOff()
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateIndexKeysOffAux(); err != nil {
		return v.syntaxError()
	}
	v.flgOff()
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexKeysOffAux() error {
	v.flgOff()
	if v.matchSymbol("(") {
		if err := v.validateExpr(); err != nil {
			return err
		}
	} else {
		if err := v.validateName(); err != nil {
			return err
		}
		if err := v.validateTypeDigitN(); err != nil {
			return v.syntaxError()
		}
	}
	if v.matchKeyword("ASC", "DESC") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if v.matchSymbol(",") {
		if v.next() != nil {
			return v.syntaxError()
		}
		return v.validateIndexKeysOffAux()
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexType() error {
	v.flgOff()
	if err := v.validateKeyword("USING"); err != nil {
		return err
	}
	if err := v.validateKeyword("BTREE", "HASH"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}

func (v *mysqlValidator) validateIndexOption() error {
	v.flgOff()
	if v.matchKeyword("KEY_BLOCK_SIZE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchSymbol("=") {
			if v.next() != nil {
				return v.syntaxError()
			}
		}
		if err := v.validateLiteralValue(); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchKeyword("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
		return v.validateIndexOption()
		
	} else if v.matchKeyword("WITH") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("PARSER"); err != nil {
			return err
		}
		if err := v.validateName(); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchKeyword("COMMENT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateStringValue(); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchKeyword("VISIBLE", "INVISIBLE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		return v.validateIndexOption()

	} else if v.matchKeyword("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchSymbol("=") {
			if v.next() != nil {
				return v.syntaxError()
			}
		}
		if err := v.validateStringValue(); err != nil {
			return err
		}
		
		return v.validateIndexOption()

	}

	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateCommaSeparatedColumnNames() error {
	if err := v.validateColumnName(); err != nil {
		return err
	}
	if v.matchSymbol(",") {
		if v.next() != nil {
			return v.syntaxError()
		}
		return v.validateCommaSeparatedColumnNames()
	}
	return nil
}


func (v *mysqlValidator) validateCommaSeparatedTableNames() error {
	if err := v.validateTableName(); err != nil {
		return err
	}
	if v.matchSymbol(",") {
		if v.next() != nil {
			return v.syntaxError()
		}
		return v.validateCommaSeparatedTableNames()
	}
	return nil
}


func (v *mysqlValidator) validateTableOptions() error {
	v.flgOff()
	if v.matchKeyword(";") {
		return nil
	}
	if v.matchSymbol(",") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateTableOption(); err != nil {
		return err
	}
	return v.validateTableOptions()
}


func (v *mysqlValidator) validateTableOption() error {
	v.flgOff()
	if v.matchKeyword(
		"AUTOEXTEND_SIZE", "AUTO_INCREMENT", "AVG_ROW_LENGTH", 
		"KEY_BLOCK_SIZE", "MAX_ROWS", "MIN_ROWS", "STATS_SAMPLE_PAGES",
	) {
		return v.validateTableOptionCommonLiteral()
	}
	if v.matchKeyword(
		"COMMENT", "ENGINE_ATTRIBUTE", "PASSWORD", "SECONDARY_ENGINE_ATTRIBUTE", "CONNECTION",
		"COMPRESSION", "ENCRYPTION",
	) {
		return v.validateTableOptionCommonString()
	}
	if v.matchKeyword("COLLATE", "ENGINE", "CHARACTER") {
		return v.validateTableOptionCommonName()
	}
	if v.matchKeyword("CHECKSUM", "DELAY_KEY_WRITE") {
		return v.validateTableOptionCommon01()
	}
	if v.matchKeyword("PACK_KEYS", "STATS_AUTO_RECALC", "STATS_PERSISTENT") {
		return v.validateTableOptionCommon01Default()
	}
	if v.matchKeyword("DATA", "INDEX") {
		return v.validateTableOptionDirectory()
	}
	if v.matchKeyword("TABLESPACE") {
		return v.validateTableOptionTablespace()
	}
	if v.matchKeyword("DEFAULT") {
		return v.validateTableOptionDefault()
	}
	if v.matchKeyword("UNION") {
		return v.validateTableOptionUnion()
	}
	if v.matchKeyword("INSERT_METHOD") {
		return v.validateTableOptionInsertMethod()
	}
	if v.matchKeyword("ROW_FORMAT") {
		return v.validateTableOptionRowFormat()
	}
	
	return v.syntaxError()
}


// option [=] 'value'
func (v *mysqlValidator) validateTableOptionCommonLiteral() error {
	v.flgOff()
	if v.next() != nil {
		return v.syntaxError()
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateLiteralValue(); err != nil {
		return err
	}
	return nil
}


// option [=] 'string'
func (v *mysqlValidator) validateTableOptionCommonString() error {
	v.flgOff()
	if v.next() != nil {
		return v.syntaxError()
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateStringValue(); err != nil {
		return err
	}
	return nil
}


// option [=] name 
func (v *mysqlValidator) validateTableOptionCommonName() error {
	v.flgOff()
	if v.matchKeyword("CHARACTER") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.validateKeyword("SET") != nil {
			return v.syntaxError()
		}
	} else {
		if err := v.validateKeyword("COLLATE", "ENGINE"); err != nil {
			return err
		}
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateName(); err != nil {
		return err
	}
	return nil
}


// option [=] {0 | 1}
func (v *mysqlValidator) validateTableOptionCommon01() error {
	v.flgOff()
	if v.next() != nil {
		return v.syntaxError()
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateSymbol("0", "1"); err != nil {
		return err
	}
	return nil
}


// option [=] {0 | 1 | DEFAULT}
func (v *mysqlValidator) validateTableOptionCommon01Default() error {
	v.flgOff()
	if v.next() != nil {
		return v.syntaxError()
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if (v.matchSymbol("0", "1") || v.matchKeyword("DEFAULT")) {
		if v.next() != nil {
			return v.syntaxError()
		}
	} else {
		return v.syntaxError()
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionDirectory() error {
	v.flgOff()
	if err := v.validateKeyword("DATA", "INDEX"); err != nil {
		return err
	}
	if err := v.validateKeyword("DIRECTORY"); err != nil {
		return err
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateLiteralValue(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionTablespace() error {
	v.flgOff()
	if err := v.validateKeyword("TABLESPACE"); err != nil {
		return err
	}
	if err := v.validateName(); err != nil {
		return err
	}
	if v.matchKeyword("STORAGE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.validateKeyword("DISK", "MEMORY") != nil {
			return v.syntaxError()
		}
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionDefault() error {
	v.flgOff()
	if err := v.validateKeyword("DEFAULT"); err != nil {
		return err
	}
	if v.matchKeyword("CHARACTER") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.validateKeyword("SET") != nil {
			return v.syntaxError()
		}
	} else if v.matchKeyword("COLLATE") {
		if v.next() != nil {
			return v.syntaxError()
		}
	} else {
		return v.syntaxError()
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateName(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionUnion() error {
	v.flgOff()
	if err := v.validateKeyword("UNION"); err != nil {
		return err
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedTableNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionInsertMethod() error {
	v.flgOff()
	if err := v.validateKeyword("INSERT_METHOD"); err != nil {
		return err
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if (v.matchKeyword("NO", "FIRST", "LAST")) {
		if v.next() != nil {
			return v.syntaxError()
		}
	} else {
		return v.syntaxError()
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionRowFormat() error {
	v.flgOff()
	if err := v.validateKeyword("ROW_FORMAT"); err != nil {
		return err
	}
	if v.matchSymbol("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if (v.matchKeyword("DEFAULT", "DYNAMIC", "FIXED", "COMPRESSED", "REDUNDANT", "COMPACT")) {
		if v.next() != nil {
			return v.syntaxError()
		}
	} else {
		return v.syntaxError()
	}
	return nil
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