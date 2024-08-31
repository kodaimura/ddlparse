package validator

import (
	"regexp"
	"strings"

	"github.com/kodaimura/ddlparse/internal/common"
)

type sqliteValidator struct {
	validator
}

func NewSQLiteValidator() Validator {
	return &sqliteValidator{validator: validator{}}
}


func (v *sqliteValidator) Validate(tokens []string) ([]string, error) {
	v.init(tokens)
	if err := v.validate(); err != nil {
		return nil, err
	}
	return v.result, nil
}


func (v *sqliteValidator) validate() error {
	if (v.isOutOfRange()) {
		return nil
	}
	if err := v.validateCreateTable(); err != nil {
		return err
	}
	return v.validate()
}


func (v *sqliteValidator) isStringValue(token string) bool {
	return token[0:1] == "'"
}


func (v *sqliteValidator) isIdentifier(token string) bool {
	tmp := token[0:1]
	return tmp == "\"" || tmp == "`"
}


func (v *sqliteValidator) isValidName(name string) bool {
	if v.isIdentifier(name) {
		return true
	} else {
		pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
		return pattern.MatchString(name) && 
			!common.Contains(ReservedWords_SQLite, strings.ToUpper(name))
	}
}


func (v *sqliteValidator) validateName() error {
	if !v.isValidName(v.token()) {
		return v.syntaxError()
	}
	v.next()

	return nil
}


func (v *sqliteValidator) validateTableName() error {
	if err := v.validateName(); err != nil {
		return err
	}
	if v.matchToken(".") {
		v.next()
		if err := v.validateName(); err != nil {
			return err
		}
	}

	return nil
}


func (v *sqliteValidator) validateColumnName() error {
	return v.validateName()
}


func (v *sqliteValidator) validateCreateTable() error {
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


func (v *sqliteValidator) validateIfNotExists() error {
	if v.matchToken("IF") {
		v.next()
		if err := v.validateToken("NOT"); err != nil {
			return err
		}
		if err := v.validateToken("EXISTS"); err != nil {
			return err
		}
	}
	return nil
}


func (v *sqliteValidator) validateTableDefinition() error {
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


func (v *sqliteValidator) validateColumnDefinitions() error {
	if err := v.validateColumnDefinition(); err != nil {
		return err
	}
	if v.matchToken(",") {
		v.flgOn()
		v.next()
		return v.validateColumnDefinitions()
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateColumnDefinition() error {
	if v.matchToken("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
		return v.validateTableConstraint()
	}
	v.flgOn()
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


func (v *sqliteValidator) validateColumnType() error {
	v.flgOn()
	if err := v.validateToken("TEXT", "NUMERIC", "INTEGER", "REAL", "NONE"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateColumnConstraints() error {
	v.flgOn()
	if v.matchToken("CONSTRAINT") {
		if v.next() == EOF {
			return nil
		}
		if err := v.validateName(); err != nil {
			return err
		}
	}
	return v.validateColumnConstraintsAux([]string{})
}


func (v *sqliteValidator) isColumnConstraint(token string) bool {
	return v.matchToken(
		"PRIMARY", "NOT", "UNIQUE", "CHECK", "DEFAULT", 
		"COLLATE", "REFERENCES", "GENERATED", "AS",
	)
}


func (v *sqliteValidator) validateColumnConstraintsAux(ls []string) error {
	if !v.isColumnConstraint(v.token()) {
		return nil
	} 
	if common.Contains(ls, strings.ToUpper(v.token())) {
		return v.syntaxError()
	} 
	ls = append(ls, strings.ToUpper(v.token()))
	if err := v.validateColumnConstraint(); err != nil {
		return err
	}
	return v.validateColumnConstraintsAux(ls)
}


func (v *sqliteValidator) validateColumnConstraint() error {
	if v.matchToken("PRIMARY") {
		return v.validateConstraintPrimaryKey()
	}
	if v.matchToken("NOT") {
		return v.validateConstraintNotNull()
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
	return v.syntaxError()
}


func (v *sqliteValidator) validateConstraintPrimaryKey() error {
	v.flgOn()
	if err := v.validateToken("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken("KEY"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchToken("ASC", "DESC") {
		v.next()
	}
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	if v.matchToken("AUTOINCREMENT") {
		v.flgOn()
		v.next()
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintNotNull() error {
	v.flgOn()
	if err := v.validateToken("NOT"); err != nil {
		return err
	}
	if err := v.validateToken("NULL"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintUnique() error {
	v.flgOn()
	if err := v.validateToken("UNIQUE"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintCheck() error {
	v.flgOn()
	if err := v.validateToken("CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintDefault() error {
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
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintCollate() error {
	v.flgOn()
	if err := v.validateToken("COLLATE"); err != nil {
		return err
	}
	if err := v.validateToken("BINARY","NOCASE", "RTRIM"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintReferences() error {
	v.flgOn()
	if err := v.validateToken("REFERENCES"); err != nil {
		return err
	}
	if err := v.validateTableName(); err != nil {
		return err
	}
	if v.matchToken("(") {
		v.next()
		if err := v.validateColumnName(); err != nil {
			return err
		}
		if err := v.validateToken(")"); err != nil {
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


func (v *sqliteValidator) validateConstraintReferencesAux() error {
	v.flgOff()
	if v.matchToken("ON") {
		v.next()
		if err := v.validateToken("DELETE", "UPDATE"); err != nil {
			return err
		}
		if v.matchToken("SET") {
			v.next()
			if err := v.validateToken("NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if v.matchToken("CASCADE", "RESTRICT") {
			v.next()
		} else if v.matchToken("NO") {
			v.next()
			if err := v.validateToken("ACTION"); err != nil {
				return err
			}
		} else {
			return v.syntaxError()
		}
		return v.validateConstraintReferencesAux()
	}

	if v.matchToken("MATCH") {
		v.next()
		if err := v.validateToken("SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return v.validateConstraintReferencesAux()
	}

	if v.matchToken("NOT", "DEFERRABLE") {
		if v.matchToken("NOT") {
			v.next()
		}
		if err := v.validateToken("DEFERRABLE"); err != nil {
			return err
		}
		if v.matchToken("INITIALLY") {
			v.next()
			if err := v.validateToken("DEFERRED", "IMMEDIATE"); err != nil {
				return err
			}
		}
		return v.validateConstraintReferencesAux()
	}

	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintGenerated() error {
	v.flgOff()
	if v.matchToken("GENERATED") {
		v.next()
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
	if v.matchToken("STORED", "VIRTUAL") {
		v.next()
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConflictClause() error {
	v.flgOff()
	if v.matchToken("ON") {
		v.next()
		if err := v.validateToken("CONFLICT"); err != nil {
			return err
		}
		if err := v.validateToken("ROLLBACK", "ABORT", "FAIL", "IGNORE","REPLACE"); err != nil {
			return err
		}
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateExpr() error {
	return v.validateBrackets()
}


func (v *sqliteValidator) validateLiteralValue() error {
	if common.IsNumericToken(v.token()) {
		v.next()
		return nil
	}
	if v.isStringValue(v.token()) {
		v.next()
		return nil
	}
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := v.validateToken(ls...); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateTableConstraint() error {
	v.flgOn()
	if v.matchToken("CONSTRAINT") {
		v.next()
		if err := v.validateName(); err != nil {
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

	return v.syntaxError()
}


func (v *sqliteValidator) validateTableConstraintPrimaryKey() error {
	v.flgOn()
	if err := v.validateToken("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken("KEY"); err != nil {
		return err
	}
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateToken(")"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateTableConstraintUnique() error {
	v.flgOn()
	if err := v.validateToken("UNIQUE"); err != nil {
		return err
	}
	if err := v.validateToken("("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(); err != nil {
		return v.syntaxError()
	}
	if err := v.validateToken(")"); err != nil {
		return err
	}
	v.flgOff()
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateTableConstraintCheck() error {
	v.flgOn()
	if err := v.validateToken("CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateTableConstraintForeignKey() error {
	v.flgOn()
	if err := v.validateToken("FOREIGN"); err != nil {
		return err
	}
	if err := v.validateToken("KEY"); err != nil {
		return err
	}
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
	if v.matchToken("(") {
		v.next()
		if err := v.validateCommaSeparatedColumnNames(); err != nil {
			return v.syntaxError()
		}
		if err := v.validateToken(")"); err != nil {
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


func (v *sqliteValidator) validateCommaSeparatedColumnNames() error {
	if err := v.validateColumnName(); err != nil {
		return err
	}
	if v.matchToken(",") {
		v.next()
		return v.validateCommaSeparatedColumnNames()
	}
	return nil
}


func (v *sqliteValidator) validateTableOptions() error {
	v.flgOff()
	if (v.isOutOfRange()) {
		return nil
	}
	if v.matchToken("WITHOUT") {
		v.next()
		if err := v.validateToken("ROWID"); err != nil {
			return err
		}
		if v.matchToken(",") {
			v.next()
			if err := v.validateToken("STRICT"); err != nil {
				return err
			}
		}
	} else if v.matchToken("STRICT") {
		v.next()
		if v.matchToken(",") {
			v.next()
			if err := v.validateToken("WITHOUT"); err != nil {
				return err
			}
			if err := v.validateToken("ROWID"); err != nil {
				return err
			}
		}
	}
	return nil
}


var ReservedWords_SQLite = []string{
	"ABORT",
	"ACTION",
	"ADD",
	"AFTER",
	"ALL",
	"ALTER",
	"ANALYZE",
	"AND",
	"AS",
	"ASC",
	"ATTACH",
	"AUTOINCREMENT",
	"BEFORE",
	"BEGIN",
	"BETWEEN",
	"BY",
	"CASCADE",
	"CASE",
	"CAST",
	"CHECK",
	"COLLATE",
	"COLUMN",
	"COMMIT",
	"CONFLICT",
	"CONSTRAINT",
	"CREATE",
	"CROSS",
	"CURRENT",
	"CURRENT_DATE",
	"CURRENT_TIME",
	"CURRENT_TIMESTAMP",
	"DATABASE",
	"DEFAULT",
	"DEFERRABLE",
	"DEFERRED",
	"DELETE",
	"DESC",
	"DETACH",
	"DISTINCT",
	"DO",
	"DROP",
	"EACH",
	"ELSE",
	"END",
	"ESCAPE",
	"EXCEPT",
	"EXCLUSIVE",
	"EXISTS",
	"EXPLAIN",
	"FAIL",
	"FILTER",
	"FOLLOWING",
	"FOR",
	"FOREIGN",
	"FROM",
	"FULL",
	"GLOB",
	"GROUP",
	"HAVING",
	"IF",
	"IGNORE",
	"IMMEDIATE",
	"IN",
	"INDEX",
	"INDEXED",
	"INITIALLY",
	"INNER",
	"INSERT",
	"INSTEAD",
	"INTERSECT",
	"INTO",
	"IS",
	"ISNULL",
	"JOIN",
	"KEY",
	"LEFT",
	"LIKE",
	"LIMIT",
	"MATCH",
	"NATURAL",
	"NO",
	"NOT",
	"NOTHING",
	"NOTNULL",
	"NULL",
	"OF",
	"OFFSET",
	"ON",
	"OR",
	"ORDER",
	"OUTER",
	"OVER",
	"PARTITION",
	"PLAN",
	"PRAGMA",
	"PRECEDING",
	"PRIMARY",
	"QUERY",
	"RAISE",
	"RANGE",
	"RECURSIVE",
	"REFERENCES",
	"REGEXP",
	"REINDEX",
	"RELEASE",
	"RENAME",
	"REPLACE",
	"RESTRICT",
	"RIGHT",
	"ROLLBACK",
	"ROW",
	"ROWS",
	"SAVEPOINT",
	"SELECT",
	"SET",
	"TABLE",
	"TEMP",
	"TEMPORARY",
	"THEN",
	"TO",
	"TRANSACTION",
	"TRIGGER",
	"UNBOUNDED",
	"UNION",
	"UNIQUE",
	"UPDATE",
	"USING",
	"VACUUM",
	"VALUES",
	"VIEW",
	"VIRTUAL",
	"WHEN",
	"WHERE",
	"WINDOW",
	"WITH",
	"WITHOUT",
}