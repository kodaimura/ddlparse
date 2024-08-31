package validator

import (
	"regexp"
	"strings"

	"github.com/kodaimura/ddlparse/internal/common"
)

type mysqlValidator struct {
	validator
}

func NewMySQLValidator() Validator {
	return &mysqlValidator{validator: validator{}}
}


func (v *mysqlValidator) Validate(tokens []string) ([]string, error) {
	v.init(tokens)
	if err := v.validate(); err != nil {
		return nil, err
	}
	return v.result, nil
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
		!common.Contains(ReservedWords_MySQL, strings.ToUpper(name))
}


func (v *mysqlValidator) isValidQuotedName(name string) bool {
	return true
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
	if v.matchToken(".") {
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


func (v *mysqlValidator) validatePositiveInteger() error {
	if !common.IsPositiveIntegerToken(v.token()) {
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


func (v *mysqlValidator) validateCreateTable() error {
	v.flgOn()
	if err := v.validateToken("CREATE"); err != nil {
		return err
	}
	if err := v.validateToken("TABLE"); err != nil {
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
	if err := v.validateToken(";"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateIfNotExists() error {
	if v.matchToken("IF") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateToken("NOT"); err != nil {
			return err
		}
		if err := v.validateToken("EXISTS"); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validateTableDefinition() error {
	v.flgOn()
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateColumnDefinitions(); err != nil {
		return err
	}
	v.flgOn()
	if err := v.validateToken(")"); err != nil {
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
	if v.matchToken(",") {
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
	if v.matchToken("CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "INDEX", "KEY", "FULLTEXT", "SPATIAL", "CHECK") {
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
	if v.matchToken("VARCHAR", "CHAR", "BINARY", "VARBINARY", "BLOB", "TEXT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitN(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchToken("NUMERIC", "DECIMAL", "FLOAT", "REAL", "DOUBLE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitPS(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchToken("BIT", "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitP(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchToken("TIME", "DATETIME", "TIMESTAMP", "YEAR") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitP(); err != nil {
			return err
		}
		v.flgOff()
		if v.matchToken("WITH", "WITHOUT") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateToken("TIME"); err != nil {
				return err
			}
			if err := v.validateToken("ZONE"); err != nil {
				return err
			}
		}
		v.flgOff()
		return nil
	}

	// TODO if v.matchToken("ENUM") {}
	// TODO if v.matchToken("SET") {}

	v.flgOn()
	if err := v.validateToken(DataType_MySQL...); err != nil {
		return err
	}

	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateTypeDate() error {
	v.flgOn()
	if err := v.validateToken("TIME", "DATETIME", "TIMESTAMP", "YEAR"); err != nil {
		return err
	}
	if err := v.validateTypeDigitP(); err != nil {
		return err
	}
	v.flgOff()
	if v.matchToken("WITH", "WITHOUT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateToken("TIME"); err != nil {
			return err
		}
		if err := v.validateToken("ZONE"); err != nil {
			return err
		}
	}
	if v.matchToken("ON") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateToken("UPDATE"); err != nil {
			return err
		}
		if err := v.validateToken("CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"); err != nil {
			return err
		}
	}
	v.flgOff()
	return nil
}

// (number)
func (v *mysqlValidator) validateTypeDigitN() error {
	if v.matchToken("(") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validatePositiveInteger(); err != nil {
			return err
		}
		if err := v.validateToken(")"); err != nil {
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
	if v.matchToken("(") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validatePositiveInteger(); err != nil {
			return err
		}
		if v.matchToken(",") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validatePositiveInteger(); err != nil {
				return err
			}
		}
		if err := v.validateToken(")"); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validateColumnConstraints() error {
	v.flgOn()
	if v.matchToken("CONSTRAINT") {
		if v.next() != nil {
			return nil
		}
		if !v.matchToken("CHECK") {
			if err := v.validateName(); err != nil {
				return err
			}
		}
	}
	return v.validateColumnConstraintsAux([]string{})
}


func (v *mysqlValidator) isColumnConstraint(token string) bool {
	return v.matchToken(
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
	if v.matchToken("NOT") {
		if common.Contains(ls, "NULL") {
			return v.syntaxError()
		} 
		ls = append(ls, "NULL")
	} else if v.matchToken("PRIMARY", "KEY") {
		if common.Contains(ls, "PRIMARY") {
			return v.syntaxError()
		} 
		ls = append(ls, "PRIMARY")
	} else if v.matchToken("GENERATED", "AS") {
		if common.Contains(ls, "GENERATED") {
			return v.syntaxError()
		} 
		ls = append(ls, "GENERATED")
	} else {
		if common.Contains(ls, strings.ToUpper(v.token())) {
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
	if v.matchToken("PRIMARY", "KEY") {
		return v.validateConstraintPrimaryKey()
	}
	if v.matchToken("NOT") {
		return v.validateConstraintNotNull()
	}
	if v.matchToken("NULL") {
		return v.validateConstraintNull()
	}
	if v.matchToken("UNIQUE") {
		return v.validateConstraintUnique()
	}
	if v.matchToken("CHECK") {
		return v.validateConstraintCheck()
	}
	if v.matchToken("DEFAULT") {
		return v.validateConstraintDefault()
	}
	if v.matchToken("COLLATE") {
		return v.validateConstraintCollate()
	}
	if v.matchToken("REFERENCES") {
		return v.validateConstraintReferences()
	}
	if v.matchToken("GENERATED", "AS") {
		return v.validateConstraintGenerated()
	}
	if v.matchToken("COMMENT") {
		return v.validateConstraintComment()
	}
	if v.matchToken("COLUMN_FORMAT") {
		return v.validateConstraintColumnFormat()
	}
	if v.matchToken("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		return v.validateConstraintEngineAttribute()
	}
	if v.matchToken("STORAGE") {
		return v.validateConstraintStorage()
	}
	if v.matchToken("AUTO_INCREMENT") {
		return v.validateConstraintAutoincrement()
	}
	if v.matchToken("VISIBLE", "INVISIBLE", "VIRTUAL", "STORED") {
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
	if v.matchToken("KEY") {
		v.result = append(v.result, "PRIMARY")
		if v.next() != nil {
			return v.syntaxError()
		}
		v.flgOff()
		return nil
	}
	if err := v.validateToken("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken("KEY"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintNotNull() error {
	v.flgOn()
	if err := v.validateToken("NOT"); err != nil {
		return err
	}
	if err := v.validateToken("NULL"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintNull() error {
	v.flgOff()
	if err := v.validateToken("NULL"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintUnique() error {
	v.flgOn()
	if err := v.validateToken("UNIQUE"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchToken("KEY") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	return nil
}


func (v *mysqlValidator) validateConstraintCheck() error {
	v.flgOn()
	if err := v.validateToken("CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	if v.matchToken("NOT") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if v.matchToken("ENFORCED") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	return nil
}


func (v *mysqlValidator) validateConstraintDefault() error {
	v.flgOn()
	if err := v.validateToken("DEFAULT"); err != nil {
		return err
	}
	if v.matchToken("(") {
		if err := v.validateExpr(); err != nil {
			return err
		}
	} else {
		if err := v.validateLiteralValue(); err != nil {
			return err
		}
		v.flgOff()
		if v.matchToken("ON") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateToken("UPDATE"); err != nil {
				return err
			}
			if err := v.validateToken("CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"); err != nil {
				return err
			}
		}
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintCollate() error {
	v.flgOn()
	if err := v.validateToken("COLLATE"); err != nil {
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
	if err := v.validateToken("REFERENCES"); err != nil {
		return err
	}
	if err := v.validateTableName(); err != nil {
		return err
	}
	if v.matchToken("(") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateColumnName(); err != nil {
			return err
		}
		if err := v.validateToken(")"); err != nil {
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
	if v.matchToken("ON") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateToken("DELETE", "UPDATE"); err != nil {
			return err
		}
		if v.matchToken("SET") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateToken("NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if v.matchToken("CASCADE", "RESTRICT") {
			if v.next() != nil {
				return v.syntaxError()
			}
		} else if v.matchToken("NO") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateToken("ACTION"); err != nil {
				return err
			}
		} else {
			return v.syntaxError()
		}
		return v.validateConstraintReferencesAux()
	}

	if v.matchToken("MATCH") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateToken("SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return v.validateConstraintReferencesAux()
	}

	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintGenerated() error {
	v.flgOff()
	if v.matchToken("GENERATED") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateToken("ALWAYS"); err != nil {
			return err
		}
	}
	if err := v.validateToken("AS"); err != nil {
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
	if err := v.validateToken("COMMENT"); err != nil {
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
	if err := v.validateToken("COLUMN_FORMAT"); err != nil {
		return err
	}
	if err := v.validateToken("FIXED", "DYNAMIC", "DEFAULT"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintEngineAttribute() error {
	v.flgOff()
	if err := v.validateToken("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE"); err != nil {
		return err
	}
	if v.matchToken("=") {
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
	if err := v.validateToken("STORAGE"); err != nil {
		return err
	}
	if err := v.validateToken("DISK", "MEMORY"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateConstraintAutoincrement() error {
	v.flgOn()
	if err := v.validateToken("AUTO_INCREMENT"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateExpr() error {
	return v.validateBrackets()
}


func (v *mysqlValidator) validateLiteralValue() error {
	if common.IsNumericToken(v.token()) {
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
	if err := v.validateToken(ls...); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableConstraint() error {
	v.flgOn()
	if v.matchToken("CONSTRAINT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if !v.matchToken("PRIMARY", "UNIQUE", "FOREIGN", "CHECK") {
			if err := v.validateName(); err != nil {
				return err
			}
		}
	}
	return v.validateTableConstraintAux()
}


func (v *mysqlValidator) validateTableConstraintAux() error {
	if v.matchToken("PRIMARY") {
		return v.validateTableConstraintPrimaryKey()
	}
	if v.matchToken("UNIQUE") {
		return v.validateTableConstraintUnique()
	}
	if v.matchToken("FOREIGN") {
		return v.validateTableConstraintForeignKey()
	}
	if v.matchToken("CHECK") {
		return v.validateTableConstraintCheck()
	}
	if v.matchToken("INDEX", "KEY") {
		return v.validateTableConstraintIndex()
	}
	if v.matchToken("FULLTEXT", "SPATIAL") {
		v.flgOff()
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchToken("INDEX", "KEY") {
			if v.next() != nil {
				return v.syntaxError()
			}
		}
		if !v.matchToken("(") {
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
	if err := v.validateToken("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken("KEY"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchToken("USING") {
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
	if err := v.validateToken("UNIQUE"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchToken("INDEX", "KEY") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if !v.matchToken("(") {
		if err := v.validateName(); err != nil {
			return err
		}
	}
	v.flgOff()
	if v.matchToken("USING") {
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
	if err := v.validateToken("FOREIGN"); err != nil {
		return err
	}
	if err := v.validateToken("KEY"); err != nil {
		return err
	}
	if !v.matchToken("(") {
		v.flgOff()
		if err := v.validateName(); err != nil {
			return err
		}
	}
	v.flgOn()
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateToken(")"); err != nil {
		return err
	}
	if err := v.validateToken("REFERENCES"); err != nil {
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
	if err := v.validateToken("INDEX", "KEY"); err != nil {
		return err
	}
	if !v.matchToken("USING") && !v.matchToken("(") {
		if err := v.validateName(); err != nil {
			return err
		}
	}
	if v.matchToken("USING") {
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
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateIndexKeysOnAux(); err != nil {
		return v.syntaxError()
	}
	v.flgOn()
	if err := v.validateToken(")"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}

func (v *mysqlValidator) validateIndexKeysOnAux() error {
	v.flgOff()
	if v.matchToken("(") {
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
	if v.matchToken("ASC", "DESC") {
		v.flgOff()
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if v.matchToken(",") {
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
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateIndexKeysOffAux(); err != nil {
		return v.syntaxError()
	}
	v.flgOff()
	if err := v.validateToken(")"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *mysqlValidator) validateIndexKeysOffAux() error {
	v.flgOff()
	if v.matchToken("(") {
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
	if v.matchToken("ASC", "DESC") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if v.matchToken(",") {
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
	if err := v.validateToken("USING"); err != nil {
		return err
	}
	if err := v.validateToken("BTREE", "HASH"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}

func (v *mysqlValidator) validateIndexOption() error {
	v.flgOff()
	if v.matchToken("KEY_BLOCK_SIZE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchToken("=") {
			if v.next() != nil {
				return v.syntaxError()
			}
		}
		if err := v.validateLiteralValue(); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchToken("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
		return v.validateIndexOption()
		
	} else if v.matchToken("WITH") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateToken("PARSER"); err != nil {
			return err
		}
		if err := v.validateName(); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchToken("COMMENT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateStringValue(); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchToken("VISIBLE", "INVISIBLE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		return v.validateIndexOption()

	} else if v.matchToken("ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchToken("=") {
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
	if v.matchToken(",") {
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
	if v.matchToken(",") {
		if v.next() != nil {
			return v.syntaxError()
		}
		return v.validateCommaSeparatedTableNames()
	}
	return nil
}


func (v *mysqlValidator) validateTableOptions() error {
	v.flgOff()
	if (v.isOutOfRange()) {
		return nil
	}
	if v.matchToken(";") {
		return nil
	}
	if v.matchToken(",") {
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
	if v.matchToken(
		"AUTOEXTEND_SIZE", "AUTO_INCREMENT", "AVG_ROW_LENGTH", 
		"KEY_BLOCK_SIZE", "MAX_ROWS", "MIN_ROWS", "STATS_SAMPLE_PAGES",
	) {
		return v.validateTableOptionCommonLiteral()
	}
	if v.matchToken(
		"COMMENT", "ENGINE_ATTRIBUTE", "PASSWORD", "SECONDARY_ENGINE_ATTRIBUTE", "CONNECTION",
		"COMPRESSION", "ENCRYPTION",
	) {
		return v.validateTableOptionCommonString()
	}
	if v.matchToken("COLLATE", "ENGINE", "CHARACTER") {
		return v.validateTableOptionCommonName()
	}
	if v.matchToken("CHECKSUM", "DELAY_KEY_WRITE") {
		return v.validateTableOptionCommon01()
	}
	if v.matchToken("PACK_KEYS", "STATS_AUTO_RECALC", "STATS_PERSISTENT") {
		return v.validateTableOptionCommon01Default()
	}
	if v.matchToken("DATA", "INDEX") {
		return v.validateTableOptionDirectory()
	}
	if v.matchToken("TABLESPACE") {
		return v.validateTableOptionTablespace()
	}
	if v.matchToken("DEFAULT") {
		return v.validateTableOptionDefault()
	}
	if v.matchToken("UNION") {
		return v.validateTableOptionUnion()
	}
	if v.matchToken("INSERT_METHOD") {
		return v.validateTableOptionInsertMethod()
	}
	if v.matchToken("ROW_FORMAT") {
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
	if v.matchToken("=") {
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
	if v.matchToken("=") {
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
	if v.matchToken("CHARACTER") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.validateToken("SET") != nil {
			return v.syntaxError()
		}
	} else {
		if err := v.validateToken("COLLATE", "ENGINE"); err != nil {
			return err
		}
	}
	if v.matchToken("=") {
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
	if v.matchToken("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateToken("0", "1"); err != nil {
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
	if v.matchToken("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if (v.matchToken("0", "1") || v.matchToken("DEFAULT")) {
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
	if err := v.validateToken("DATA", "INDEX"); err != nil {
		return err
	}
	if err := v.validateToken("DIRECTORY"); err != nil {
		return err
	}
	if v.matchToken("=") {
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
	if err := v.validateToken("TABLESPACE"); err != nil {
		return err
	}
	if err := v.validateName(); err != nil {
		return err
	}
	if v.matchToken("STORAGE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.validateToken("DISK", "MEMORY") != nil {
			return v.syntaxError()
		}
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionDefault() error {
	v.flgOff()
	if err := v.validateToken("DEFAULT"); err != nil {
		return err
	}
	if v.matchToken("CHARACTER") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.validateToken("SET") != nil {
			return v.syntaxError()
		}
	} else if v.matchToken("COLLATE") {
		if v.next() != nil {
			return v.syntaxError()
		}
	} else {
		return v.syntaxError()
	}
	if v.matchToken("=") {
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
	if err := v.validateToken("UNION"); err != nil {
		return err
	}
	if v.matchToken("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedTableNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateToken(")"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionInsertMethod() error {
	v.flgOff()
	if err := v.validateToken("INSERT_METHOD"); err != nil {
		return err
	}
	if v.matchToken("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if (v.matchToken("NO", "FIRST", "LAST")) {
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
	if err := v.validateToken("ROW_FORMAT"); err != nil {
		return err
	}
	if v.matchToken("=") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if (v.matchToken("DEFAULT", "DYNAMIC", "FIXED", "COMPRESSED", "REDUNDANT", "COMPACT")) {
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