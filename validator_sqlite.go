package ddlparse

import (
	"fmt"
	"strconv"
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


func (v *sqliteValidator) token() string {
	return v.tokens[v.i]
}


func (v *sqliteValidator) isOutOfRange() bool {
	return v.i > v.size - 1
}


func (v *sqliteValidator) Validate() ([]string, error) {
	v.initV()
	if err := v.validate(); err != nil {
		return nil, err
	}
	return v.validatedTokens, nil
}


func (v *sqliteValidator) initV() {
	v.validatedTokens = []string{}
	v.i = -1
	v.line = 1
	v.size = len(v.tokens)
	v.flg = false
	v.next()
}


func (v *sqliteValidator) flgOn() {
	v.flg = true
}


func (v *sqliteValidator) flgOff() {
	v.flg = false
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
	pattern := regexv.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
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
	if v.validateSymbol(".") == nil {
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

	v.flgOff()
	if v.validateKeyword("IF") == nil {
		if err := v.validateKeyword("NOT"); err != nil {
			return err
		}
		if err := v.validateKeyword("EXISTS"); err != nil {
			return err
		}
	}

	v.flgOn()
	if err := v.validateTableName(); err != nil {
		return err
	}
	v.flgOn()
	if err := v.validateSymbol("("); err != nil {
		return err
	}
	if err := v.validateColumns(); err != nil {
		return err
	}
	v.flgOn()
	if err := v.validateSymbol(")"); err != nil {
		return err
	}
	if err := v.validateTableOptions(); err != nil {
		return err
	}
	v.flgOn()
	if (v.token() == ";") {
		if v.next() != nil {
			return nil
		}
	}

	return v.validateCreateTable()
}


func (v *sqliteValidator) validateColumns() error {
	v.flgOn()
	if err := v.validateColumn(); err != nil {
		return err
	}
	v.flgOn()
	if v.validateSymbol(",") == nil {
		return v.validateColumns()
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateColumn() error {
	v.flgOn()
	if v.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
		return v.validateTableConstraint()
	}
	if err := v.validateColumnName(); err != nil {
		return err
	}
	if err := v.validateColumnType(); err != nil {
		return err
	}
	if err := v.validateColumnConstraint(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


// Omitting data types is not supported.
func (v *sqliteValidator) validateColumnType() error {
	v.flgOn()
	if err := v.validateKeyword("TEXT", "NUMERIC", "INTEGER", "REAL", "NONE"); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateColumnConstraint() error {
	v.flgOff()
	if v.validateKeyword("CONSTRAINT") == nil {
		if err := v.validateName(); err != nil {
			return err
		}
	}
	v.flgOn()
	return v.validateColumnConstraintAux([]string{})
}


func (v *sqliteValidator) validateColumnConstraintAux(ls []string) error {
	if v.matchKeyword("PRIMARY") {
		if contains(ls, "PRIMARY") {
			return v.syntaxError()
		}
		v.flgOn()
		if err := v.validateConstraintPrimaryKey(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "PRIMARY"))
	}

	if v.matchKeyword("NOT") {
		if contains(ls, "NOT") {
			return v.syntaxError()
		}
		v.flgOn()
		if err := v.validateConstraintNotNull(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "NOT"))
	}

	if v.matchKeyword("UNIQUE") {
		if contains(ls, "UNIQUE") {
			return v.syntaxError()
		}
		v.flgOn()
		if err := v.validateConstraintUnique(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "UNIQUE"))
	}

	if v.matchKeyword("CHECK") {
		if contains(ls, "CHECK") {
			return v.syntaxError()
		}
		v.flgOff()
		if err := v.validateConstraintCheck(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "CHECK"))
	}

	if v.matchKeyword("DEFAULT") {
		if contains(ls, "DEFAULT") {
			return v.syntaxError()
		}
		v.flgOn()
		if err := v.validateConstraintDefault(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "DEFAULT"))
	}

	if v.matchKeyword("COLLATE") {
		if contains(ls, "COLLATE") {
			return v.syntaxError()
		}
		v.flgOff()
		if err := v.validateConstraintCollate(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "COLLATE"))
	}

	if v.matchKeyword("REFERENCES") {
		if contains(ls, "REFERENCES") {
			return v.syntaxError()
		}
		v.flgOff()
		if err := v.validateConstraintForeignKey(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "REFERENCES"))
	}

	if v.matchKeyword("GENERATED", "AS") {
		if contains(ls, "GENERATED") {
			return v.syntaxError()
		}
		v.flgOff()
		if err := v.validateConstraintGenerated(); err != nil {
			return err
		}
		return v.validateColumnConstraintAux(append(ls, "GENERATED"))
	}

	v.flgOff()
	return nil
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
	v.flgOn()
	if v.matchKeyword("AUTOINCREMENT") {
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
	v.flgOff()
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
	v.flgOff()
	if err := v.validateKeyword("COLLATE"); err != nil {
		return err
	}
	if err := v.validateKeyword("BINARY","NOCASE", "RTRIM"); err != nil {
		return err
	}
	return nil
}


func (v *sqliteValidator) validateConstraintForeignKey() error {
	v.flgOff()
	if err := v.validateKeyword("REFERENCES"); err != nil {
		return err
	}
	if err := v.validateTableName(); err != nil {
		return err
	}
	if v.validateSymbol("(") == nil {
		if err := v.validateCommaSeparatedColumnNames(); err != nil {
			return err
		}
		if err := v.validateSymbol(")"); err != nil {
			return err
		}
	}
	if err := v.validateConstraintForeignKeyAux(); err != nil {
		return v.syntaxError()
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintForeignKeyAux() error {
	v.flgOff()
	if v.validateKeyword("ON") == nil {
		if err := v.validateKeyword("DELETE", "UPDATE"); err != nil {
			return err
		}
		if v.validateKeyword("SET") == nil {
			if err := v.validateKeyword("NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if v.validateKeyword("CASCADE", "RESTRICT") == nil {

		} else if v.validateKeyword("NO") == nil {
			if err := v.validateKeyword("ACTION"); err != nil {
				return err
			}
		} else {
			return v.syntaxError()
		}
		return v.validateConstraintForeignKeyAux()
	}

	if v.validateKeyword("MATCH") == nil {
		if err := v.validateKeyword("SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return v.validateConstraintForeignKeyAux()
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
		if v.validateKeyword("INITIALLY") == nil {
			if err := v.validateKeyword("DEFERRED", "IMMEDIATE"); err != nil {
				return err
			}
		}
		return v.validateConstraintForeignKeyAux()
	}

	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateConstraintGenerated() error {
	v.flgOff()
	if v.validateKeyword("GENERATED") == nil {
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
	if v.validateKeyword("ON") == nil {
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
	v.flgOff()
	if v.validateKeyword("CONSTRAINT") == nil{
		if err := v.validateName(); err != nil {
			return err
		}
	}
	return v.validateTableConstraintAux()
}


func (v *sqliteValidator) validateTableConstraintAux() error {
	if v.matchKeyword("PRIMARY") {
		return v.validateTablePrimaryKey()
	}

	if v.matchKeyword("UNIQUE") {
		return v.validateTableUnique()
	}

	if v.matchKeyword("CHECK") {
		return v.validateTableCheck()
	}

	if v.matchKeyword("FOREIGN") {
		return v.validateTableForeignKey()
	}

	return v.syntaxError()
}


func (v *sqliteValidator) validateTablePrimaryKey() error {
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


func (v *sqliteValidator) validateTableUnique() error {
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


func (v *sqliteValidator) validateTableCheck() error {
	v.flgOff()
	if err := v.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := v.validateExpr(); err != nil {
		return err
	}
	v.flgOff()
	return nil
}


func (v *sqliteValidator) validateTableForeignKey() error {
	v.flgOff()
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
	if err := v.validateConstraintForeignKey(); err != nil {
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