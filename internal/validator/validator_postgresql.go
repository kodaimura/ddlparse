package validator

import (
	"regexp"
	"strings"

	"github.com/kodaimura/ddlparse/internal/common"
)

type postgresqlValidator struct {
	validator
}

func NewPostgreSQLValidator() Validator {
	return &postgresqlValidator{validator: validator{}}
}


func (v *postgresqlValidator) Validate(tokens []string) ([]string, error) {
	v.init(tokens)
	if err := v.validate(); err != nil {
		return nil, err
	}
	return v.result, nil
}


func (v *postgresqlValidator) validate() error {
	if (v.isOutOfRange()) {
		return nil
	}
	if err := v.validateCreateTable(); err != nil {
		return err
	}
	return v.validate()
}


func (v *postgresqlValidator) isStringValue(token string) bool {
	return token[0:1] == "'"
}


func (v *postgresqlValidator) isIdentifier(token string) bool {
	return token[0:1] == "\""
}


func (v *postgresqlValidator) isValidName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!common.Contains(ReservedWords_PostgreSQL, strings.ToUpper(name))
}


func (v *postgresqlValidator) isValidQuotedName(name string) bool {
	return true
}


func (v *postgresqlValidator) validateName() error {
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


func (v *postgresqlValidator) validateTableName() error {
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


func (v *postgresqlValidator) validateColumnName() error {
	return v.validateName()
}


func (v *postgresqlValidator) validatePositiveInteger() error {
	if !common.IsPositiveIntegerToken(v.token()) {
		return v.syntaxError()
	}
	if v.next() != nil {
		return v.syntaxError()
	}
	return nil
}


func (v *postgresqlValidator) validateBrackets() error {
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


func (v *postgresqlValidator) validateBracketsAux() error {
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


func (v *postgresqlValidator) validateCreateTable() error {
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
	if err := v.validateSymbol(";"); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateIfNotExists() error {
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


func (v *postgresqlValidator) validateTableDefinition() error {
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


func (v *postgresqlValidator) validateColumnDefinitions() error {
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


func (v *postgresqlValidator) validateColumnDefinition() error {
	v.flgOn()
	if v.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN", "EXCLUDE") {
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


func (v *postgresqlValidator) validateColumnType() error {
	v.flgOn()
	if v.matchKeyword("BIT", "CHARACTER") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchKeyword("VARYING") {
			if v.next() != nil {
				return v.syntaxError()
			}
		}
		if err := v.validateTypeDigitN(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchKeyword("VARBIT", "VARCHAR", "CHAR") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitN(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchKeyword("NUMERIC", "DECIMAL") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateTypeDigitPS(); err != nil {
			return err
		}
		v.flgOff()
		return nil
	}

	if v.matchKeyword("DOUBLE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		v.flgOff()
		if err := v.validateKeyword("PRECISION"); err != nil {
			return err
		}
		return nil
	}

	// TODO
	//if v.matchKeyword("INTERVAL") {
	//}

	if v.matchKeyword("TIME", "TIMESTAMP") {
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
		return nil
	}

	v.flgOn()
	if err := v.validateKeyword(DataType_PostgreSQL...); err != nil {
		return err
	}

	v.flgOff()
	return nil
}

// (number)
func (v *postgresqlValidator) validateTypeDigitN() error {
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
func (v *postgresqlValidator) validateTypeDigitP() error {
	return v.validateTypeDigitN()
}

// (presision. scale)
func (v *postgresqlValidator) validateTypeDigitPS() error {
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


func (v *postgresqlValidator) validateColumnConstraints() error {
	if v.matchKeyword("CONSTRAINT") {
		v.flgOn()
		if v.next() != nil {
			return nil
		}
		if err := v.validateName(); err != nil {
			return err
		}
	}
	return v.validateColumnConstraintsAux([]string{})
}


func (v *postgresqlValidator) isColumnConstraint(token string) bool {
	return v.matchKeyword(
		"PRIMARY", "NOT", "NULL", "UNIQUE", "CHECK", 
		"DEFAULT", "REFERENCES", "GENERATED", "AS",
	)
}


func (v *postgresqlValidator) validateColumnConstraintsAux(ls []string) error {
	if !v.isColumnConstraint(v.token()) {
		v.flgOff()
		return nil
	} 
	if v.matchKeyword("NOT") {
		if common.Contains(ls, "NULL") {
			return v.syntaxError()
		} 
		ls = append(ls, strings.ToUpper("NULL"))
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


func (v *postgresqlValidator) validateColumnConstraint() error {
	if v.matchKeyword("PRIMARY") {
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
	if v.matchKeyword("REFERENCES") {
		return v.validateConstraintReferences()
	}
	if v.matchKeyword("GENERATED", "AS") {
		return v.validateConstraintGenerated()
	}
	return v.syntaxError()
}


func (v *postgresqlValidator) validateConstraintPrimaryKey() error {
	v.flgOn()
	if err := v.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateKeyword("KEY"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateConstraintNotNull() error {
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


func (v *postgresqlValidator) validateConstraintNull() error {
	v.flgOff()
	if err := v.validateKeyword("NULL"); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintUnique() error {
	v.flgOn()
	if err := v.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateConstraintCheck() error {
	v.flgOn()
	if err := v.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("NO") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("INHERIT"); err != nil {
			return err
		}
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateConstraintDefault() error {
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
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateConstraintReferences() error {
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
		return err
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateConstraintReferencesAux() error {
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


func (v *postgresqlValidator) validateConstraintGenerated() error {
	v.flgOff()
	if err := v.validateKeyword("GENERATED"); err != nil {
		return err
	}

	if v.matchKeyword("ALWAYS") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("AS"); err != nil {
			return err
		}
		if v.matchKeyword("IDENTITY") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if v.matchSymbol("(") {
				if err := v.validateBrackets(); err != nil {
					return err
				}
			}
			v.flgOff()
			return nil

		} else if v.matchSymbol("(") {
			if err := v.validateBrackets(); err != nil {
				return err
			}
			if err := v.validateKeyword("STORED"); err != nil {
				return err
			}
			v.flgOff()
			return nil

		} else {
			return v.syntaxError()
		}
	} else if v.matchKeyword("BY") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("DEFAULT"); err != nil {
			return err
		}
		if err := v.validateKeyword("AS"); err != nil {
			return err
		}
		if err := v.validateKeyword("IDENTITY"); err != nil {
			return err
		}
		if v.matchSymbol("(") {
			if err := v.validateBrackets(); err != nil {
				return err
			}
		}
		v.flgOff()
		return nil

	} else if v.matchKeyword("AS") {
		if err := v.validateKeyword("AS"); err != nil {
			return err
		}
		if err := v.validateKeyword("IDENTITY"); err != nil {
			return err
		}
		if v.matchSymbol("(") {
			if err := v.validateBrackets(); err != nil {
				return err
			}
		}
		v.flgOff()
		return nil
	}

	return v.syntaxError()
}


func (v *postgresqlValidator) validateExpr() error {
	return v.validateBrackets()
}


func (v *postgresqlValidator) validateIndexParameters() error {
	v.flgOff()
	if v.matchKeyword("INCLUDE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateSymbol("("); err != nil {
			return err
		}
		if err := v.validateCommaSeparatedColumnNames(); err != nil {
			return v.syntaxError()
		}
		if err := v.validateSymbol(")"); err != nil {
			return err
		}
	}
	if v.matchKeyword("WITH") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateBrackets(); err != nil {
			return err
		}
	}
	if v.matchKeyword("USING") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("INDEX"); err != nil {
			return err
		}
		if err := v.validateKeyword("TABLESPACE"); err != nil {
			return err
		}
		if err := v.validateName(); err != nil {
			return err
		}
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateLiteralValue() error {
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
	if err := v.validateKeyword(ls...); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableConstraint() error {
	v.flgOn()
	if v.matchKeyword("CONSTRAINT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateName(); err != nil {
			return err
		}
	}
	if v.matchKeyword("PRIMARY") {
		return v.validateTableConstraintPrimaryKey()
	}
	if v.matchKeyword("UNIQUE") {
		return v.validateTableConstraintUnique()
	}
	if v.matchKeyword("CHECK") {
		return v.validateTableConstraintCheck()
	}
	if v.matchKeyword("FOREIGN") {
		return v.validateTableConstraintForeignKey()
	}
	if v.matchKeyword("EXCLUDE") {
		return v.validateTableConstraintExclude()
	}

	return v.syntaxError()
}


func (v *postgresqlValidator) validateTableConstraintPrimaryKey() error {
	v.flgOn()
	if err := v.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateKeyword("KEY"); err != nil {
		return err
	}
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateTableConstraintUnique() error {
	v.flgOn()
	if err := v.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateTableConstraintCheck() error {
	v.flgOn()
	if err := v.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("NO") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("INHERIT"); err != nil {
			return err
		}
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateTableConstraintForeignKey() error {
	v.flgOn()
	if err := v.validateKeyword("FOREIGN"); err != nil {
		return err
	}
	if err := v.validateKeyword("KEY"); err != nil {
		return err
	}
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
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
		if err := v.validateCommaSeparatedColumnNames(); err != nil {
			return v.syntaxError()
		}
		if err := v.validateSymbol(")"); err != nil {
			return err
		}
	}
	v.flgOff()
	if err := v.validateConstraintReferencesAux(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateTableConstraintExclude() error {
	v.flgOff()
	if err := v.validateKeyword("EXCLUDE"); err != nil {
		return err
	}
	if v.matchKeyword("USING") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateName(); err != nil {
			return err
		}
	}
	if err := v.validateBrackets(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	if v.matchKeyword("WHERE") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateBrackets(); err != nil {
			return err
		}
	}
	v.flgOff()
	return nil
}


func (v *postgresqlValidator) validateCommaSeparatedColumnNames() error {
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


func (v *postgresqlValidator) validateTableOptions() error {
	v.flgOff()
	if (v.isOutOfRange()) {
		return nil
	}
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


func (v *postgresqlValidator) validateTableOption() error {
	v.flgOff()
	if v.matchKeyword("WITH") {
		return v.validateTableOptionWith()
	}
	if v.matchKeyword("WITHOUT") {
		return v.validateTableOptionWithout()
	}
	if v.matchKeyword("TABLESPACE") {
		return v.validateTableOptionTablespace()
	}
	if v.matchKeyword("INHERITS") {
		return v.validateTableOptionInherits()
	}
	if v.matchKeyword("PARTITION") {
		return v.validateTableOptionPartition()
	}
	if v.matchKeyword("USING") {
		return v.validateTableOptionUsing()
	}
	return v.syntaxError()
}


func (v *postgresqlValidator) validateTableOptionWith() error {
	v.flgOff()
	if err := v.validateKeyword("WITH"); err != nil {
		return err
	}
	if err := v.validateBrackets(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionWithout() error {
	v.flgOff()
	if err := v.validateKeyword("WITHOUT"); err != nil {
		return err
	}
	if err := v.validateKeyword("OIDS"); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionTablespace() error {
	v.flgOff()
	if err := v.validateKeyword("TABLESPACE"); err != nil {
		return err
	}
	if err := v.validateName(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionInherits() error {
	v.flgOff()
	if err := v.validateKeyword("INHERITS"); err != nil {
		return err
	}
	if err := v.validateBrackets(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionPartition() error {
	v.flgOff()
	if err := v.validateKeyword("PARTITION"); err != nil {
		return err
	}
	if err := v.validateKeyword("BY"); err != nil {
		return err
	}
	if err := v.validateKeyword("RANGE", "LIST", "HASH"); err != nil {
		return err
	}
	if err := v.validateBrackets(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionUsing() error {
	v.flgOff()
	if err := v.validateKeyword("USING"); err != nil {
		return err
	}
	if err := v.validateName(); err != nil {
		return err
	}
	return nil
}


var DataType_PostgreSQL = []string{
	"BIGINT",
	"INT8",
	"BIGSERIAL",
	"SERIAL8",
	"BIT",
	"VARBIT",
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
	"PG_SNAPSHOT",
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