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
	if v.isIdentifier(name) {
		return true
	} else {
		pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
		return pattern.MatchString(name) && 
			!common.Contains(ReservedWords_MySQL, strings.ToUpper(name))
	}
}


func (v *mysqlValidator) validateName(set bool) error {
	if v.isValidName(v.token()) {
		if set {
			v.set(v.next())
		} else {
			v.next()
		}
		return nil
	}
	return v.syntaxError()
}


func (v *mysqlValidator) validateTableName(set bool) error {
	if err := v.validateName(set); err != nil {
		return err
	}
	if v.matchTokenNext(set, ".") {
		if err := v.validateName(set); err != nil {
			return err
		}
	}

	return nil
}


func (v *mysqlValidator) validateColumnName(set bool) error {
	return v.validateName(set)
}


func (v *mysqlValidator) validateStringValue(set bool) error {
	if !v.isStringValue(v.token()) {
		return v.syntaxError()
	}
	if set {
		v.set(v.next())
	} else {
		v.next()
	}
	return nil
}


// (number)
func (v *mysqlValidator) validateTypeDigitN(set bool) error {
	if v.matchTokenNext(set, "(") {
		if err := v.validatePositiveInteger(set); err != nil {
			return err
		}
		if err := v.validateToken(set, ")"); err != nil {
			return err
		}
	} 
	return nil
}


// (presision)
func (v *mysqlValidator) validateTypeDigitP(set bool) error {
	return v.validateTypeDigitN(set)
}


// (presision. scale)
func (v *mysqlValidator) validateTypeDigitPS(set bool) error {
	if v.matchTokenNext(set, "(") {
		if err := v.validatePositiveInteger(set); err != nil {
			return err
		}
		if v.matchTokenNext(set, ",") {
			if err := v.validatePositiveInteger(set); err != nil {
				return err
			}
		}
		if err := v.validateToken(set, ")"); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validatePositiveInteger(set bool) error {
	if !common.IsPositiveIntegerToken(v.token()) {
		return v.syntaxError()
	}
	if set {
		v.set(v.next())
	} else {
		v.next()
	}
	
	return nil
}


func (v *mysqlValidator) validateExpr(set bool) error {
	return v.validateBrackets(set)
}


func (v *mysqlValidator) validateLiteralValue(set bool) error {
	if common.IsNumericToken(v.token()) {
		if set {
			v.set(v.next())
		} else {
			v.next()
		}
		return nil
	}
	if v.isStringValue(v.token()) {
		if set {
			v.set(v.next())
		} else {
			v.next()
		}
		return nil
	}
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := v.validateToken(set, ls...); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateIndexKeys(set bool) error {
	if err := v.validateToken(set, "("); err != nil {
		return err
	}
	if err := v.validateIndexKeysAux(set); err != nil {
		return err
	}
	if err := v.validateToken(true, ")"); err != nil {
		return err
	}
	return nil
}

func (v *mysqlValidator) validateIndexKeysAux(set bool) error {
	if v.matchToken("(") {
		if err := v.validateExpr(false); err != nil {
			return err
		}
	} else {
		if err := v.validateName(set); err != nil {
			return err
		}
		if err := v.validateTypeDigitN(false); err != nil {
			return err
		}
	}
	v.matchTokenNext(false, "ASC", "DESC")
	if v.matchTokenNext(set, ",") {
		return v.validateIndexKeysAux(set)
	}
	return nil
}


func (v *mysqlValidator) validateCreateTable() error {
	if err := v.validateToken(true, "CREATE"); err != nil {
		return err
	}
	if err := v.validateToken(true, "TABLE"); err != nil {
		return err
	}
	if err := v.validateIfNotExists(); err != nil {
		return err
	}
	if err := v.validateTableName(true); err != nil {
		return err
	}
	if err := v.validateTableDefinition(); err != nil {
		return err
	}
	if err := v.validateToken(true, ";"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateIfNotExists() error {
	if v.matchTokenNext(true, "IF") {
		if err := v.validateToken(true, "NOT"); err != nil {
			return err
		}
		if err := v.validateToken(true, "EXISTS"); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validateTableDefinition() error {
	if err := v.validateToken(true, "("); err != nil {
		return err
	}
	if err := v.validateColumnDefinitions(); err != nil {
		return err
	}
	if err := v.validateToken(true, ")"); err != nil {
		return err
	}
	if err := v.validateTableOptions(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateColumnDefinitions() error {
	if err := v.validateColumnDefinition(); err != nil {
		return err
	}
	if v.matchTokenNext(true, ",") {
		return v.validateColumnDefinitions()
	}
	return nil
}


func (v *mysqlValidator) validateColumnDefinition() error {
	if v.matchToken("CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "INDEX", "KEY", "FULLTEXT", "SPATIAL", "CHECK") {
		return v.validateTableConstraint()
	}
	if err := v.validateColumnName(true); err != nil {
		return err
	}
	if err := v.validateColumnType(); err != nil {
		return err
	}
	if err := v.validateColumnConstraints(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateColumnType() error {
	if v.matchTokenNext(true, "VARCHAR", "CHAR", "BINARY", "VARBINARY", "BLOB", "TEXT") {
		if err := v.validateTypeDigitN(true); err != nil {
			return err
		}
		return nil
	}

	if v.matchTokenNext(true, "NUMERIC", "DECIMAL", "FLOAT", "REAL", "DOUBLE") {
		if err := v.validateTypeDigitPS(true); err != nil {
			return err
		}
		return nil
	}

	if v.matchTokenNext(true, "BIT", "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT") {
		if err := v.validateTypeDigitP(true); err != nil {
			return err
		}
		return nil
	}

	if v.matchTokenNext(true, "TIME", "DATETIME", "TIMESTAMP", "YEAR") {
		if err := v.validateTypeDigitP(true); err != nil {
			return err
		}
		if v.matchTokenNext(false, "WITH", "WITHOUT") {
			if err := v.validateToken(false, "TIME"); err != nil {
				return err
			}
			if err := v.validateToken(false, "ZONE"); err != nil {
				return err
			}
		}
		return nil
	}

	// TODO if v.matchToken("ENUM") {}
	// TODO if v.matchToken("SET") {}

	if err := v.validateToken(true, DataType_MySQL...); err != nil {
		return err
	}

	return nil
}


func (v *mysqlValidator) validateColumnConstraints() error {
	if v.matchTokenNext(true, "CONSTRAINT") {
		if !v.matchToken("CHECK") {
			if err := v.validateName(true); err != nil {
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
	if v.matchTokenNext(false, "VISIBLE", "INVISIBLE", "VIRTUAL", "STORED") {
		return nil
	}
	
	return v.syntaxError()
}


func (v *mysqlValidator) validateConstraintPrimaryKey() error {
	if v.matchToken("KEY") {
		v.set("PRIMARY")
		v.set("KEY")
		v.next()
		return nil
	}
	if err := v.validateToken(true, "PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken(true, "KEY"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintNotNull() error {
	if err := v.validateToken(true, "NOT"); err != nil {
		return err
	}
	if err := v.validateToken(true, "NULL"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintNull() error {
	if err := v.validateToken(false, "NULL"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintUnique() error {
	if err := v.validateToken(true, "UNIQUE"); err != nil {
		return err
	}
	v.matchTokenNext(false, "KEY")
	return nil
}


func (v *mysqlValidator) validateConstraintCheck() error {
	if err := v.validateToken(true, "CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(true); err != nil {
		return err
	}
	v.matchTokenNext(false, "NOT")
	v.matchTokenNext(false, "ENFORCED")
	return nil
}


func (v *mysqlValidator) validateConstraintDefault() error {
	if err := v.validateToken(true, "DEFAULT"); err != nil {
		return err
	}
	if v.matchToken("(") {
		if err := v.validateExpr(true); err != nil {
			return err
		}
	} else {
		if err := v.validateLiteralValue(true); err != nil {
			return err
		}

		if v.matchTokenNext(false, "ON") {
			if err := v.validateToken(false, "UPDATE"); err != nil {
				return err
			}
			if err := v.validateToken(false, "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"); err != nil {
				return err
			}
		}
	}
	return nil
}


func (v *mysqlValidator) validateConstraintCollate() error {
	if err := v.validateToken(true, "COLLATE"); err != nil {
		return err
	}
	if err := v.validateName(true); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintReferences() error {
	if err := v.validateToken(true, "REFERENCES"); err != nil {
		return err
	}
	if err := v.validateTableName(true); err != nil {
		return err
	}
	if v.matchTokenNext(true, "(") {
		if err := v.validateColumnName(true); err != nil {
			return err
		}
		if err := v.validateToken(true, ")"); err != nil {
			return err
		}
	}
	if err := v.validateConstraintReferencesAux(); err != nil {
		return v.syntaxError()
	}
	return nil
}


func (v *mysqlValidator) validateConstraintReferencesAux() error {
	if v.matchTokenNext(false, "ON") {
		if err := v.validateToken(false, "DELETE", "UPDATE"); err != nil {
			return err
		}
		if v.matchTokenNext(false, "SET") {
			if err := v.validateToken(false, "NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if v.matchTokenNext(false, "CASCADE", "RESTRICT") {

		} else if v.matchTokenNext(false, "NO") {
			if err := v.validateToken(false, "ACTION"); err != nil {
				return err
			}
		} else {
			return v.syntaxError()
		}
		return v.validateConstraintReferencesAux()
	}

	if v.matchTokenNext(false, "MATCH") {
		if err := v.validateToken(false, "SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return v.validateConstraintReferencesAux()
	}

	return nil
}


func (v *mysqlValidator) validateConstraintGenerated() error {
	if v.matchTokenNext(false, "GENERATED") {
		if err := v.validateToken(false, "ALWAYS"); err != nil {
			return err
		}
	}
	if err := v.validateToken(false, "AS"); err != nil {
		return err
	}
	if err := v.validateExpr(false); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintComment() error {
	if err := v.validateToken(false, "COMMENT"); err != nil {
		return err
	}
	if err := v.validateStringValue(false); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintColumnFormat() error {
	if err := v.validateToken(false, "COLUMN_FORMAT"); err != nil {
		return err
	}
	if err := v.validateToken(false, "FIXED", "DYNAMIC", "DEFAULT"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintEngineAttribute() error {
	if err := v.validateToken(false, "ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE"); err != nil {
		return err
	}
	v.matchTokenNext(false, "=")
	if err := v.validateStringValue(false); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintStorage() error {
	if err := v.validateToken(false, "STORAGE"); err != nil {
		return err
	}
	if err := v.validateToken(false, "DISK", "MEMORY"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateConstraintAutoincrement() error {
	if err := v.validateToken(true, "AUTO_INCREMENT"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableConstraint() error {
	if v.matchTokenNext(true, "CONSTRAINT") {
		if !v.matchToken("PRIMARY", "UNIQUE", "FOREIGN", "CHECK") {
			if err := v.validateName(true); err != nil {
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
	if v.matchTokenNext(false, "FULLTEXT", "SPATIAL") {
		v.matchTokenNext(false, "INDEX", "KEY")
		if !v.matchToken("(") {
			if err := v.validateName(true); err != nil {
				return err
			}
		}
		if err := v.validateIndexKeys(false); err != nil {
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
	if err := v.validateToken(true, "PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken(true, "KEY"); err != nil {
		return err
	}
	if v.matchToken("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v.validateIndexKeys(true); err != nil {
		return err
	}
	if err := v.validateIndexOption(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableConstraintUnique() error {
	if err := v.validateToken(true, "UNIQUE"); err != nil {
		return err
	}
	v.matchTokenNext(false, "INDEX", "KEY")
	if !v.matchToken("(") {
		if err := v.validateName(false); err != nil {
			return err
		}
	}
	if v.matchToken("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v.validateIndexKeys(true); err != nil {
		return err
	}
	if err := v.validateIndexOption(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableConstraintForeignKey() error {
	if err := v.validateToken(true, "FOREIGN"); err != nil {
		return err
	}
	if err := v.validateToken(true, "KEY"); err != nil {
		return err
	}
	if !v.matchToken("(") {
		if err := v.validateName(false); err != nil {
			return err
		}
	}
	if err := v.validateToken(true, "("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(true); err != nil {
		return err
	}
	if err := v.validateToken(true, ")"); err != nil {
		return err
	}
	if err := v.validateToken(true, "REFERENCES"); err != nil {
		return err
	}
	if err := v.validateTableName(true); err != nil {
		return err
	}
	if err := v.validateIndexKeys(true); err != nil {
		return v.syntaxError()
	}
	if err := v.validateConstraintReferencesAux(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableConstraintCheck() error {
	return v.validateConstraintCheck()
}


func (v *mysqlValidator) validateTableConstraintIndex() error {
	if err := v.validateToken(false, "INDEX", "KEY"); err != nil {
		return err
	}
	if !v.matchToken("USING") && !v.matchToken("(") {
		if err := v.validateName(false); err != nil {
			return err
		}
	}
	if v.matchToken("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
	} 
	if err := v.validateIndexKeys(false); err != nil {
		return err
	}
	if err := v.validateIndexOption(); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateIndexType() error {
	if err := v.validateToken(false, "USING"); err != nil {
		return err
	}
	if err := v.validateToken(false, "BTREE", "HASH"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateIndexOption() error {
	if v.matchTokenNext(false, "KEY_BLOCK_SIZE") {
		v.matchTokenNext(false, "=")
		if err := v.validateLiteralValue(false); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchToken("USING") {
		if err := v.validateIndexType(); err != nil {
			return err
		}
		return v.validateIndexOption()
		
	} else if v.matchTokenNext(false, "WITH") {
		if err := v.validateToken(false, "PARSER"); err != nil {
			return err
		}
		if err := v.validateName(false); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchTokenNext(false, "COMMENT") {
		if err := v.validateStringValue(false); err != nil {
			return err
		}
		return v.validateIndexOption()

	} else if v.matchTokenNext(false, "VISIBLE", "INVISIBLE") {
		return v.validateIndexOption()

	} else if v.matchTokenNext(false, "ENGINE_ATTRIBUTE", "SECONDARY_ENGINE_ATTRIBUTE") {
		v.matchTokenNext(false, "=")
		if err := v.validateStringValue(false); err != nil {
			return err
		}
		return v.validateIndexOption()
	}

	return nil
}


func (v *mysqlValidator) validateCommaSeparatedColumnNames(set bool) error {
	if err := v.validateColumnName(set); err != nil {
		return err
	}
	if v.matchTokenNext(set, ",") {
		return v.validateCommaSeparatedColumnNames(set)
	}
	return nil
}


func (v *mysqlValidator) validateCommaSeparatedTableNames(set bool) error {
	if err := v.validateTableName(set); err != nil {
		return err
	}
	if v.matchTokenNext(set, ",") {
		return v.validateCommaSeparatedTableNames(set)
	}
	return nil
}


func (v *mysqlValidator) validateTableOptions() error {
	if (v.isOutOfRange()) {
		return nil
	}
	if v.matchToken(";") {
		return nil
	}
	v.matchTokenNext(false, ",")
	if err := v.validateTableOption(); err != nil {
		return err
	}
	return v.validateTableOptions()
}


func (v *mysqlValidator) validateTableOption() error {
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
	v.next()
	v.matchTokenNext(false, "=")
	if err := v.validateLiteralValue(false); err != nil {
		return err
	}
	return nil
}


// option [=] 'string'
func (v *mysqlValidator) validateTableOptionCommonString() error {
	v.next()
	v.matchTokenNext(false, "=")
	if err := v.validateStringValue(false); err != nil {
		return err
	}
	return nil
}


// option [=] name 
func (v *mysqlValidator) validateTableOptionCommonName() error {
	if v.matchTokenNext(false, "CHARACTER") {
		if err := v.validateToken(false, "SET"); err != nil {
			return err
		}
	} else {
		if err := v.validateToken(false, "COLLATE", "ENGINE"); err != nil {
			return err
		}
	}
	v.matchTokenNext(false, "=")
	if err := v.validateName(false); err != nil {
		return err
	}
	return nil
}


// option [=] {0 | 1}
func (v *mysqlValidator) validateTableOptionCommon01() error {
	v.next()
	v.matchTokenNext(false, "=")
	if err := v.validateToken(false, "0", "1"); err != nil {
		return err
	}
	return nil
}


// option [=] {0 | 1 | DEFAULT}
func (v *mysqlValidator) validateTableOptionCommon01Default() error {
	v.next()
	v.matchTokenNext(false, "=")
	if err := v.validateToken(false, "0", "1", "DEFAULT"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionDirectory() error {
	if err := v.validateToken(false, "DATA", "INDEX"); err != nil {
		return err
	}
	if err := v.validateToken(false, "DIRECTORY"); err != nil {
		return err
	}
	v.matchTokenNext(false, "=")
	if err := v.validateLiteralValue(false); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionTablespace() error {
	if err := v.validateToken(false, "TABLESPACE"); err != nil {
		return err
	}
	if err := v.validateName(false); err != nil {
		return err
	}
	if v.matchTokenNext(false, "STORAGE") {
		if err := v.validateToken(false, "DISK", "MEMORY"); err != nil {
			return err
		}
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionDefault() error {
	if err := v.validateToken(false, "DEFAULT"); err != nil {
		return err
	}
	if v.matchTokenNext(false, "CHARACTER") {
		if err := v.validateToken(false, "SET"); err != nil {
			return err
		}
	} else if v.matchTokenNext(false, "COLLATE") {

	} else {
		return v.syntaxError()
	}
	v.matchTokenNext(false, "=")
	if err := v.validateName(false); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionUnion() error {
	if err := v.validateToken(false, "UNION"); err != nil {
		return err
	}
	v.matchTokenNext(false, "=")
	if err := v.validateToken(false, "("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedTableNames(false); err != nil {
		return err
	}
	if err := v.validateToken(false, ")"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionInsertMethod() error {
	if err := v.validateToken(false, "INSERT_METHOD"); err != nil {
		return err
	}
	v.matchTokenNext(false, "=")
	if err := v.validateToken(false, "NO", "FIRST", "LAST"); err != nil {
		return err
	}
	return nil
}


func (v *mysqlValidator) validateTableOptionRowFormat() error {
	if err := v.validateToken(false, "ROW_FORMAT"); err != nil {
		return err
	}
	v.matchTokenNext(false, "=")
	if err := v.validateToken(false,"DEFAULT", "DYNAMIC", "FIXED", "COMPRESSED", "REDUNDANT", "COMPACT"); err != nil {
		return err
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