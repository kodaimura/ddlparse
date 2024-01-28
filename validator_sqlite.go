package ddlparse

import (
	"errors"
	"regexp"
	"strings"
)

type sqliteValidator struct {
	tokens []string
	validatedTokens []string
	size int
	i int
	line int
	flg bool
}

func newSQLiteValidator(tokens []string) validator {
	return &sqliteValidator{tokens: tokens}
}


func (v *sqliteValidator) Validate() ([]string, error) {
	v.init()
	if err := v.validate(); err != nil {
		return nil, err
	}
	return v.validatedTokens, nil
}


func (v *sqliteValidator) init() {
	v.validatedTokens = []string{}
	v.i = -1
	v.line = 1
	v.size = len(v.tokens)
	v.flg = false
	v.next()
}


func (v *sqliteValidator) token() string {
	return v.tokens[v.i]
}


func (v *sqliteValidator) flgOn() {
	v.flg = true
}


func (v *sqliteValidator) flgOff() {
	v.flg = false
}


func (v *sqliteValidator) isOutOfRange() bool {
	return v.i > v.size - 1
}


func (v *sqliteValidator) next() error {
	if v.flg {
		v.validatedTokens = append(v.validatedTokens, v.token())
	}
	return v.nextAux()
}


func (v *sqliteValidator) nextAux() error {
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


func (v *sqliteValidator) syntaxError() error {
	if v.isOutOfRange() {
		return NewValidateError(v.line, v.tokens[v.size - 1])
	}
	return NewValidateError(v.line, v.tokens[v.i])
}


func (v *sqliteValidator) matchKeyword(keywords ...string) bool {
	return contains(
		append(
			mapSlice(keywords, strings.ToLower), 
			mapSlice(keywords, strings.ToUpper)...,
		), v.token())
}


func (v *sqliteValidator) matchSymbol(symbols ...string) bool {
	return contains(symbols, v.token())
}


func (v *sqliteValidator) isStringValue(token string) bool {
	return token[0:1] == "'"
}


func (v *sqliteValidator) isIdentifier(token string) bool {
	tmp := token[0:1]
	return tmp == "\"" || tmp == "`"
}


func (v *sqliteValidator) isValidName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!contains(ReservedWords_SQLite, strings.ToUpper(name))
}


func (v *sqliteValidator) isValidQuotedName(name string) bool {
	return true
}


func (v *sqliteValidator) validateKeyword(keywords ...string) error {
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


func (v *sqliteValidator) validateSymbol(symbols ...string) error {
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


func (v *sqliteValidator) validateName() error {
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


func (v *sqliteValidator) validateTableName() error {
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


func (v *sqliteValidator) validateColumnName() error {
	return v.validateName()
}


func (v *sqliteValidator) validateBrackets() error {
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


func (v *sqliteValidator) validateBracketsAux() error {
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


func (v *sqliteValidator) validate() error {
	if (v.isOutOfRange()) {
		return nil
	}
	if err := v.validateCreateTable(); err != nil {
		return err
	}
	return v.validate()
}


func (v *sqliteValidator) validateCreateTable() error {
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

	v.flgOn()
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


func (v *sqliteValidator) validateIfNotExists() error {
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


func (v *sqliteValidator) validateTableDefinition() error {
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


func (v *sqliteValidator) validateColumnDefinitions() error {
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


func (v *sqliteValidator) validateColumnDefinition() error {
	if v.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
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
	if err := v.validateKeyword("TEXT", "NUMERIC", "INTEGER", "REAL", "NONE"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateColumnConstraints() error {
	v.flgOn()
	if v.matchKeyword("CONSTRAINT") {
		if v.next() != nil {
			return nil
		}
		if err := v.validateName(); err != nil {
			return err
		}
	}
	return v.validateColumnConstraintsAux([]string{})
}


func (v *sqliteValidator) isColumnConstraint(token string) bool {
	return v.matchKeyword(
		"PRIMARY", "NOT", "UNIQUE", "CHECK", "DEFAULT", 
		"COLLATE", "REFERENCES", "GENERATED", "AS",
	)
}


func (v *sqliteValidator) validateColumnConstraintsAux(ls []string) error {
	if !v.isColumnConstraint(v.token()) {
		return nil
	} 
	if contains(ls, strings.ToUpper(v.token())) {
		return v.syntaxError()
	} 
	ls = append(ls, strings.ToUpper(v.token()))
	if err := v.validateColumnConstraint(); err != nil {
		return err
	}
	return v.validateColumnConstraintsAux(ls)
}


func (v *sqliteValidator) validateColumnConstraint() error {
	if v.matchKeyword("PRIMARY") {
		return v.validateConstraintPrimaryKey()
	}
	if v.matchKeyword("NOT") {
		return v.validateConstraintNotNull()
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
	return v.syntaxError()
}


func (v *sqliteValidator) validateConstraintPrimaryKey() error {
	v.flgOn()
	if err := v.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := v.validateKeyword("KEY"); err != nil {
		return err
	}
	v.flgOff()
	if v.matchKeyword("ASC", "DESC") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	if v.matchKeyword("AUTOINCREMENT") {
		v.flgOn()
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintNotNull() error {
	v.flgOn()
	if err := v.validateKeyword("NOT"); err != nil {
		return err
	}
	if err := v.validateKeyword("NULL"); err != nil {
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
	if err := v.validateKeyword("UNIQUE"); err != nil {
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
	if err := v.validateKeyword("CHECK"); err != nil {
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


func (v *sqliteValidator) validateConstraintCollate() error {
	v.flgOn()
	if err := v.validateKeyword("COLLATE"); err != nil {
		return err
	}
	if err := v.validateKeyword("BINARY","NOCASE", "RTRIM"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintReferences() error {
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
	v.flgOff()
	if err := v.validateConstraintReferencesAux(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintReferencesAux() error {
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

	if v.matchKeyword("NOT", "DEFERRABLE") {
		if v.matchKeyword("NOT") {
			if v.next() != nil {
				return v.syntaxError()
			}
		}
		if err := v.validateKeyword("DEFERRABLE"); err != nil {
			return err
		}
		if v.matchKeyword("INITIALLY") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateKeyword("DEFERRED", "IMMEDIATE"); err != nil {
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
	if v.matchKeyword("STORED", "VIRTUAL") {
		if v.next() != nil {
			return v.syntaxError()
		}
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConflictClause() error {
	v.flgOff()
	if v.matchKeyword("ON") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("CONFLICT"); err != nil {
			return err
		}
		if err := v.validateKeyword("ROLLBACK", "ABORT", "FAIL", "IGNORE","REPLACE"); err != nil {
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


func (v *sqliteValidator) validateTableConstraint() error {
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

	return v.syntaxError()
}


func (v *sqliteValidator) validateTableConstraintPrimaryKey() error {
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
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateTableConstraintUnique() error {
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
	if err := v.validateConflictClause(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateTableConstraintCheck() error {
	v.flgOn()
	if err := v.validateKeyword("CHECK"); err != nil {
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


func (v *sqliteValidator) validateCommaSeparatedColumnNames() error {
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


func (v *sqliteValidator) validateTableOptions() error {
	v.flgOff()
	if v.matchKeyword("WITHOUT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if err := v.validateKeyword("ROWID"); err != nil {
			return err
		}
		if v.matchSymbol(",") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateKeyword("STRICT"); err != nil {
				return err
			}
		}
	} else if v.matchKeyword("STRICT") {
		if v.next() != nil {
			return v.syntaxError()
		}
		if v.matchSymbol(",") {
			if v.next() != nil {
				return v.syntaxError()
			}
			if err := v.validateKeyword("WITHOUT"); err != nil {
				return err
			}
			if err := v.validateKeyword("ROWID"); err != nil {
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