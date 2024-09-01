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
	if err := v.validateDdl(); err != nil {
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


func (v *sqliteValidator) validateName(set bool) error {
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


func (v *sqliteValidator) validateTableName(set bool) error {
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


func (v *sqliteValidator) validateColumnName(set bool) error {
	return v.validateName(set)
}


func (v *sqliteValidator) validateCommaSeparatedColumnNames(set bool) error {
	if err := v.validateColumnName(set); err != nil {
		return err
	}
	if v.matchTokenNext(set, ",") {
		return v.validateCommaSeparatedColumnNames(set)
	}
	return nil
}


func (v *sqliteValidator) validateExpr(set bool) error {
	return v.validateBrackets(set)
}


func (v *sqliteValidator) validateDdl() error {
	if err := v.validateToken(false, "CREATE"); err != nil {
		return err
	}
	v.matchTokenNext(false, "TEMP", "TEMPORARY")
	if v.matchToken("TABLE") {
		if err := v.validateCreateTable(); err != nil {
			return err
		}
	} else {
		if err := v.validateCreateOther(); err != nil {
			return err
		}
	}
	return nil
}


func (v *sqliteValidator) validateCreateTable() error {
	v.set("CREATE")
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


func (v *sqliteValidator) validateCreateOther() error {
	if err := v.validateToken(false, "VIRTUAL", "VIEW", "TRIGGER", "INDEX", "UNIQUE"); err != nil {
		return err
	}
	begin := false
	for true {
		if v.isOutOfRange() {
			return v.syntaxError()
		}
		if v.matchToken("BEGIN") {
			begin = true
		}
		if begin {
			if v.matchToken("END") {
				begin = false
			}
		} else {
			if v.matchToken(";") {
				break
			}
		}
		v.next()
	}
	if err := v.validateToken(false, ";"); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateIfNotExists() error {
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


func (v *sqliteValidator) validateTableDefinition() error {
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


func (v *sqliteValidator) validateColumnDefinitions() error {
	if err := v.validateColumnDefinition(); err != nil {
		return err
	}
	if v.matchTokenNext(true, ",") {
		return v.validateColumnDefinitions()
	}
	return nil
}


func (v *sqliteValidator) validateColumnDefinition() error {
	if v.matchToken("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
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


func (v *sqliteValidator) validateColumnType() error {
	return v.validateToken(true,
		"TEXT", "NUMERIC", "INTEGER", "REAL", "NONE",
	)
}


func (v *sqliteValidator) validateColumnConstraints() error {
	if v.matchTokenNext(true, "CONSTRAINT") {
		if err := v.validateName(true); err != nil {
			return err
		}
	}
	return v.validateColumnConstraintsAux([]string{})
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


func (v *sqliteValidator) isColumnConstraint(token string) bool {
	return v.matchToken(
		"PRIMARY", "NOT", "UNIQUE", "CHECK", "DEFAULT", 
		"COLLATE", "REFERENCES", "GENERATED", "AS",
	)
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
	if err := v.validateToken(true, "PRIMARY"); err != nil {
		return err
	}
	if err := v.validateToken(true, "KEY"); err != nil {
		return err
	}
	v.matchTokenNext(false, "ASC", "DESC")
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	v.matchTokenNext(true, "AUTOINCREMENT")
	return nil
}


func (v *sqliteValidator) validateConstraintNotNull() error {
	if err := v.validateToken(true, "NOT"); err != nil {
		return err
	}
	if err := v.validateToken(true, "NULL"); err != nil {
		return err
	}
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateConstraintUnique() error {
	if err := v.validateToken(true, "UNIQUE"); err != nil {
		return err
	}
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateConstraintCheck() error {
	if err := v.validateToken(true, "CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(true); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateConstraintDefault() error {
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


func (v *sqliteValidator) validateLiteralValue() error {
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


func (v *sqliteValidator) validateConstraintCollate() error {
	if err := v.validateToken(true, "COLLATE"); err != nil {
		return err
	}
	if err := v.validateToken(true, "BINARY","NOCASE", "RTRIM"); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateConstraintReferences() error {
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


func (v *sqliteValidator) validateConstraintReferencesAux() error {
	if v.matchTokenNext(false, "ON") {
		if !v.matchTokenNext(false, "DELETE", "UPDATE") {
			return v.syntaxError()
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

	if v.matchToken("NOT", "DEFERRABLE") {
		v.matchTokenNext(false, "NOT")
		if err := v.validateToken(false, "DEFERRABLE"); err != nil {
			return err
		}
		if v.matchTokenNext(false, "INITIALLY") {
			if err := v.validateToken(false, "DEFERRED", "IMMEDIATE"); err != nil {
				return err
			}
		}
		return v.validateConstraintReferencesAux()
	}

	return nil
}


func (v *sqliteValidator) validateConstraintGenerated() error {
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
	v.matchTokenNext(false, "STORED", "VIRTUAL")
	return nil
}


func (v *sqliteValidator) validateConflictClause() error {
	if v.matchTokenNext(false, "ON") {
		if err := v.validateToken(false, "CONFLICT"); err != nil {
			return err
		}
		if err := v.validateToken(false, "ROLLBACK", "ABORT", "FAIL", "IGNORE","REPLACE"); err != nil {
			return err
		}
	}
	return nil
}


func (v *sqliteValidator) validateTableConstraint() error {
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

	return v.syntaxError()
}


func (v *sqliteValidator) validateTableConstraintPrimaryKey() error {
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
		return v.syntaxError()
	}
	if err := v.validateToken(true, ")"); err != nil {
		return err
	}
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateTableConstraintUnique() error {
	if err := v.validateToken(true, "UNIQUE"); err != nil {
		return err
	}
	if err := v.validateToken(true, "("); err != nil {
		return err
	}
	if err := v.validateCommaSeparatedColumnNames(true, ); err != nil {
		return err
	}
	if err := v.validateToken(true, ")"); err != nil {
		return err
	}
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateTableConstraintCheck() error {
	if err := v.validateToken(true, "CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(true); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateTableConstraintForeignKey() error {
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


func (v *sqliteValidator) validateTableOptions() error {
	if (v.isOutOfRange()) {
		return nil
	}
	if v.matchTokenNext(false, "WITHOUT") {
		if err := v.validateToken(false, "ROWID"); err != nil {
			return err
		}
		if v.matchTokenNext(false, ",") {
			if err := v.validateToken(false, "STRICT"); err != nil {
				return err
			}
		}
	} else if v.matchTokenNext(false, "STRICT") {
		if v.matchTokenNext(false, ",") {
			if err := v.validateToken(false, "WITHOUT"); err != nil {
				return err
			}
			if err := v.validateToken(false, "ROWID"); err != nil {
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