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
	if v.isIdentifier(name) {
		return true
	} else {
		pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
		return pattern.MatchString(name) && 
			!common.Contains(ReservedWords_PostgreSQL, strings.ToUpper(name))
	}
}


func (v *postgresqlValidator) validateName(set bool) error {
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


func (v *postgresqlValidator) validateTableName(set bool) error {
	if err := v.validateName(set); err != nil {
		return err
	}
	if v.matchToken(".") {
		if set {
			v.set(v.next())
		} else {
			v.next()
		}
		if err := v.validateName(set); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validateColumnName(set bool) error {
	return v.validateName(set)
}


func (v *postgresqlValidator) validateExpr(set bool) error {
	return v.validateBrackets(set)
}


func (v *postgresqlValidator) validateCommaSeparatedColumnNames(set bool) error {
	if err := v.validateColumnName(set); err != nil {
		return err
	}
	if v.matchTokenNext(set, ",") {
		return v.validateCommaSeparatedColumnNames(set)
	}
	return nil
}


func (v *postgresqlValidator) validateCreateTable() error {
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


func (v *postgresqlValidator) validateIfNotExists() error {
	if v.matchTokenNext(false, "IF") {
		if err := v.validateToken(false, "NOT"); err != nil {
			return err
		}
		if err := v.validateToken(false, "EXISTS"); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validateTableDefinition() error {
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


func (v *postgresqlValidator) validateColumnDefinitions() error {
	if err := v.validateColumnDefinition(); err != nil {
		return err
	}
	if v.matchTokenNext(true, ",") {
		return v.validateColumnDefinitions()
	}
	return nil
}


func (v *postgresqlValidator) validateColumnDefinition() error {
	if v.matchToken("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN", "EXCLUDE") {
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


func (v *postgresqlValidator) validateColumnType() error {
	if v.matchTokenNext(true, "BIT", "CHARACTER") {
		v.matchTokenNext(true, "VARYING")
		if err := v.validateTypeDigitN(); err != nil {
			return err
		}
		return nil
	}

	if v.matchTokenNext(true, "VARBIT", "VARCHAR", "CHAR") {
		if err := v.validateTypeDigitN(); err != nil {
			return err
		}
		return nil
	}

	if v.matchTokenNext(true, "NUMERIC", "DECIMAL") {
		if err := v.validateTypeDigitPS(); err != nil {
			return err
		}
		return nil
	}

	if v.matchTokenNext(true, "DOUBLE") {
		if err := v.validateToken(false, "PRECISION"); err != nil {
			return err
		}
		return nil
	}

	// TODO
	//if v.matchToken("INTERVAL") {
	//}

	if v.matchTokenNext(true, "TIME", "TIMESTAMP") {
		if err := v.validateTypeDigitP(); err != nil {
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

	if err := v.validateToken(true, DataType_PostgreSQL...); err != nil {
		return err
	}

	return nil
}

// (number)
func (v *postgresqlValidator) validateTypeDigitN() error {
	if v.matchTokenNext(true, "(") {
		if err := v.validatePositiveInteger(); err != nil {
			return err
		}
		if err := v.validateToken(true, ")"); err != nil {
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
	if v.matchTokenNext(true, "(") {
		if err := v.validatePositiveInteger(); err != nil {
			return err
		}
		if v.matchTokenNext(true, ",") {
			if err := v.validatePositiveInteger(); err != nil {
				return err
			}
		}
		if err := v.validateToken(true, ")"); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validatePositiveInteger() error {
	if !common.IsPositiveIntegerToken(v.token()) {
		return v.syntaxError()
	}
	v.set(v.next())
	return nil
}


func (v *postgresqlValidator) validateColumnConstraints() error {
	if v.matchTokenNext(true, "CONSTRAINT") {
		if err := v.validateName(true); err != nil {
			return err
		}
	}
	return v.validateColumnConstraintsAux([]string{})
}


func (v *postgresqlValidator) isColumnConstraint(token string) bool {
	return v.matchToken(
		"PRIMARY", "NOT", "NULL", "UNIQUE", "CHECK", 
		"DEFAULT", "REFERENCES", "GENERATED", "AS",
	)
}


func (v *postgresqlValidator) validateColumnConstraintsAux(ls []string) error {
	if !v.isColumnConstraint(v.token()) {
		return nil
	} 
	if v.matchToken("NOT") {
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
	if v.matchToken("PRIMARY") {
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
	if v.matchToken("REFERENCES") {
		return v.validateConstraintReferences()
	}
	if v.matchToken("GENERATED", "AS") {
		return v.validateConstraintGenerated()
	}
	return v.syntaxError()
}


func (v *postgresqlValidator) validateConstraintPrimaryKey() error {
	if err := v.validateToken(true, "PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken(true, "KEY"); err != nil {
		return err
	}
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintNotNull() error {
	if err := v.validateToken(true, "NOT"); err != nil {
		return err
	}
	if err := v.validateToken(true, "NULL"); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintNull() error {
	if err := v.validateToken(false, "NULL"); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintUnique() error {
	if err := v.validateToken(true, "UNIQUE"); err != nil {
		return err
	}
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintCheck() error {
	if err := v.validateToken(true, "CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(true); err != nil {
		return err
	}
	if v.matchTokenNext(false, "NO") {
		if err := v.validateToken(false, "INHERIT"); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintDefault() error {
	if err := v.validateToken(true, "DEFAULT"); err != nil {
		return err
	}
	if v.matchToken("(") {
		if err := v.validateExpr(true); err != nil {
			return err
		}
	} else {
		if err := v.validateLiteralValue(); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validateLiteralValue() error {
	if common.IsNumericToken(v.token()) {
		v.set(v.next())
		return nil
	}
	if v.isStringValue(v.token()) {
		v.set(v.next())
		return nil
	}
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := v.validateToken(true, ls...); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintReferences() error {
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
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateConstraintReferencesAux() error {
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


func (v *postgresqlValidator) validateConstraintGenerated() error {
	if err := v.validateToken(false, "GENERATED"); err != nil {
		return err
	}

	if v.matchTokenNext(false, "ALWAYS") {
		if err := v.validateToken(false, "AS"); err != nil {
			return err
		}
		if v.matchTokenNext(false, "IDENTITY") {
			if v.matchToken("(") {
				if err := v.validateBrackets(false); err != nil {
					return err
				}
			}
			return nil

		} else if v.matchToken("(") {
			if err := v.validateBrackets(false); err != nil {
				return err
			}
			if err := v.validateToken(false, "STORED"); err != nil {
				return err
			}
			return nil

		} else {
			return v.syntaxError()
		}
	} else if v.matchTokenNext(false, "BY") {
		if err := v.validateToken(false, "DEFAULT"); err != nil {
			return err
		}
		if err := v.validateToken(false, "AS"); err != nil {
			return err
		}
		if err := v.validateToken(false, "IDENTITY"); err != nil {
			return err
		}
		if v.matchToken("(") {
			if err := v.validateBrackets(false); err != nil {
				return err
			}
		}
		return nil

	} else if v.matchTokenNext(false, "AS") {
		if err := v.validateToken(false, "IDENTITY"); err != nil {
			return err
		}
		if v.matchToken("(") {
			if err := v.validateBrackets(false); err != nil {
				return err
			}
		}
		return nil
	}

	return v.syntaxError()
}


func (v *postgresqlValidator) validateIndexParameters() error {
	if v.matchTokenNext(false, "INCLUDE") {
		if err := v.validateToken(false, "("); err != nil {
			return err
		}
		if err := v.validateCommaSeparatedColumnNames(false); err != nil {
			return err
		}
		if err := v.validateToken(false, ")"); err != nil {
			return err
		}
	}
	if v.matchTokenNext(false, "WITH") {
		if err := v.validateBrackets(false); err != nil {
			return err
		}
	}
	if v.matchTokenNext(false, "USING") {
		if err := v.validateToken(false, "INDEX"); err != nil {
			return err
		}
		if err := v.validateToken(false, "TABLESPACE"); err != nil {
			return err
		}
		if err := v.validateName(false); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validateTableConstraint() error {
	if v.matchTokenNext(true, "CONSTRAINT") {
		if err := v.validateName(true); err != nil {
			return err
		}
	}
	if v.matchToken("PRIMARY") {
		return v.validateTableConstraintPrimaryKey()
	}
	if v.matchToken("UNIQUE") {
		return v.validateTableConstraintUnique()
	}
	if v.matchToken("CHECK") {
		return v.validateTableConstraintCheck()
	}
	if v.matchToken("FOREIGN") {
		return v.validateTableConstraintForeignKey()
	}
	if v.matchToken("EXCLUDE") {
		return v.validateTableConstraintExclude()
	}

	return v.syntaxError()
}


func (v *postgresqlValidator) validateTableConstraintPrimaryKey() error {
	if err := v.validateToken(true, "PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken(true, "KEY"); err != nil {
		return err
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
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableConstraintUnique() error {
	if err := v.validateToken(true, "UNIQUE"); err != nil {
		return err
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
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableConstraintCheck() error {
	if err := v.validateToken(true, "CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(true); err != nil {
		return err
	}
	if v.matchTokenNext(false, "NO") {
		if err := v.validateToken(false, "INHERIT"); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validateTableConstraintForeignKey() error {
	if err := v.validateToken(true, "FOREIGN"); err != nil {
		return err
	}
	if err := v.validateToken(true, "KEY"); err != nil {
		return err
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
	if v.matchTokenNext(true, "(") {
		if err := v.validateCommaSeparatedColumnNames(true); err != nil {
			return err
		}
		if err := v.validateToken(true, ")"); err != nil {
			return err
		}
	}
	if err := v.validateConstraintReferencesAux(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableConstraintExclude() error {
	if err := v.validateToken(false, "EXCLUDE"); err != nil {
		return err
	}
	if v.matchTokenNext(false, "USING") {
		if err := v.validateName(false); err != nil {
			return err
		}
	}
	if err := v.validateBrackets(false); err != nil {
		return err
	}
	if err := v.validateIndexParameters(); err != nil {
		return err
	}
	if v.matchTokenNext(false, "WHERE") {
		if err := v.validateBrackets(false); err != nil {
			return err
		}
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptions() error {
	if (v.isOutOfRange()) {
		return nil
	}
	if v.matchToken(";") {
		return nil
	}
	if v.matchToken(",") {
		v.next()
	}
	if err := v.validateTableOption(); err != nil {
		return err
	}
	return v.validateTableOptions()
}


func (v *postgresqlValidator) validateTableOption() error {
	if v.matchToken("WITH") {
		return v.validateTableOptionWith()
	}
	if v.matchToken("WITHOUT") {
		return v.validateTableOptionWithout()
	}
	if v.matchToken("TABLESPACE") {
		return v.validateTableOptionTablespace()
	}
	if v.matchToken("INHERITS") {
		return v.validateTableOptionInherits()
	}
	if v.matchToken("PARTITION") {
		return v.validateTableOptionPartition()
	}
	if v.matchToken("USING") {
		return v.validateTableOptionUsing()
	}
	return v.syntaxError()
}


func (v *postgresqlValidator) validateTableOptionWith() error {
	if err := v.validateToken(false, "WITH"); err != nil {
		return err
	}
	if err := v.validateBrackets(false); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionWithout() error {
	if err := v.validateToken(false, "WITHOUT"); err != nil {
		return err
	}
	if err := v.validateToken(false, "OIDS"); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionTablespace() error {
	v.flgOff()
	if err := v.validateToken("TABLESPACE"); err != nil {
		return err
	}
	if err := v.validateName(); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionInherits() error {
	if err := v.validateToken(false, "INHERITS"); err != nil {
		return err
	}
	if err := v.validateBrackets(false); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionPartition() error {
	if err := v.validateToken(false, "PARTITION"); err != nil {
		return err
	}
	if err := v.validateToken(false, "BY"); err != nil {
		return err
	}
	if err := v.validateToken(false, "RANGE", "LIST", "HASH"); err != nil {
		return err
	}
	if err := v.validateBrackets(false); err != nil {
		return err
	}
	return nil
}


func (v *postgresqlValidator) validateTableOptionUsing() error {
	if err := v.validateToken(false, "USING"); err != nil {
		return err
	}
	if err := v.validateName(false); err != nil {
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