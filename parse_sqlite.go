package ddlparse

import (
	"errors"
	"regexp"
	"strings"
)

type sqliteParser struct {
	tokens []string
	size int
	i int
	line int
}

func newSQLiteParser(tokens []string) parser {
	return &sqliteParser{tokens, len(tokens), -1, 1}
}

func (p *sqliteParser) token() string {
	return p.tokens[p.i]
}

func (p *sqliteParser) isOutOfRange() bool {
	return p.i > p.size - 1
}

func (p *sqliteParser) syntaxError() error {
	if p.isOutOfRange() {
		return NewValidateError(p.line, p.tokens[p.size - 1])
	}
	return NewValidateError(p.line, p.tokens[p.i])
}

func (p *sqliteParser) init() {
	p.i = -1
	p.line = 1
	p.next()
}

func (p *sqliteParser) next() error {
	p.i += 1
	if (p.isOutOfRange()) {
		return errors.New("out of range")
	}
	if (p.token() == "\n") {
		p.line += 1
		return p.next()
	} else if (p.token() == "--") {
		p.skipSingleLineComment()
		return p.next()
	} else if (p.token() == "/*") {
		if err := p.skipMultiLineComment(); err != nil {
			return err
		}
		return p.next()
	} else {
		return nil
	}
}

func (p *sqliteParser) skipSingleLineComment() {
	if (p.token() != "--") {
		return
	}
	var skip func()
	skip = func() {
		p.i += 1
		if (p.isOutOfRange()) {
			return
		} else if (p.token() == "\n") {
			p.line += 1
		} else {
			skip()
		}
	}
	skip()
}

func (p *sqliteParser) skipMultiLineComment() error {
	if (p.token() != "/*") {
		return nil
	}
	var skip func() error
	skip = func() error {
		p.i += 1
		if (p.isOutOfRange()) {
			return errors.New("out of range")
		} else if (p.token() == "\n") {
			p.line += 1
			return skip()
		} else if (p.token() == "*/") {
			return nil
		} else {
			return skip()
		}
	}
	return skip()
}

func (p *sqliteParser) isValidName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!contains(ReservedWords_SQLite, strings.ToUpper(name))
}

func (p *sqliteParser) isValidQuotedName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9_]*$`)
	return pattern.MatchString(name)
}

func (p *sqliteParser) Validate() error {
	p.init()
	return p.validate()
}

func (p *sqliteParser) validate() error {
	if (p.isOutOfRange()) {
		return nil
	}
	if err := p.validateCreateTable(); err != nil {
		return err
	}
	return p.validate()
}

func (p *sqliteParser) matchKeyword(keywords ...string) bool {
	return contains(
		append(
			mapSlice(keywords, strings.ToLower), 
			mapSlice(keywords, strings.ToUpper)...,
		), p.token())
}

func (p *sqliteParser) matchSymbol(symbols ...string) bool {
	return contains(symbols, p.token())
}

func (p *sqliteParser) validateKeyword(keywords ...string) error {
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if p.matchKeyword(keywords...) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	return p.syntaxError()
}

func (p *sqliteParser) validateSymbol(symbols ...string) error {
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if p.matchSymbol(symbols...) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	return p.syntaxError()
}

func (p *sqliteParser) validateName() error {
	if p.validateSymbol("\"") == nil {
		if !p.isValidQuotedName(p.token()) {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateSymbol("\""); err != nil {
			return p.syntaxError()
		}
	} else if p.validateSymbol("`") == nil {
		if p.next() != nil {
			return p.syntaxError()
		}
		if !p.isValidQuotedName(p.token()) {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateSymbol("`"); err != nil {
			return p.syntaxError()
		}
	} else {
		if !p.isValidName(p.token()) {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
	}

	return nil
}

func (p *sqliteParser) validateCreateTable() error {
	if err := p.validateKeyword("CREATE"); err != nil {
		return err
	}
	if err := p.validateKeyword("TABLE"); err != nil {
		return err
	}
	if p.validateKeyword("IF") == nil {
		if err := p.validateKeyword("NOT"); err != nil {
			return err
		}
		if err := p.validateKeyword("EXISTS"); err != nil {
			return err
		}
	}

	if err := p.validateTableName(); err != nil {
		return err
	}
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateColumns(); err != nil {
		return err
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	if (p.token() == ";") {
		if p.next() != nil {
			return nil
		}
	}

	return p.validateCreateTable()
}

func (p *sqliteParser) validateTableName() error {
	if err := p.validateName(); err != nil {
		return err
	}
	if p.validateSymbol(".") == nil {
		if err := p.validateName(); err != nil {
			return err
		}
	}

	return nil
}

func (p *sqliteParser) validateColumns() error {
	if err := p.validateColumn(); err != nil {
		return err
	}
	if p.validateSymbol(",") == nil {
		return p.validateColumns()
	}

	return nil
}

func (p *sqliteParser) validateColumn() error {
	if contains(TableConstraint_SQLite, strings.ToUpper(p.token())) {
		return p.validateTableConstraint()
	}

	if err := p.validateColumnName(); err != nil {
		return err
	}
	if err := p.validateColumnType(); err != nil {
		return err
	}
	if err := p.validateColumnConstraint(); err != nil {
		return err
	}
	
	return nil
}

func (p *sqliteParser) validateColumnName() error {
	return p.validateName()
}

// Omitting data types is not supported.
func (p *sqliteParser) validateColumnType() error {
	return p.validateKeyword(DataType_SQLite...)
}

func (p *sqliteParser) validateColumnConstraint() error {
	if p.validateKeyword("CONSTRAINT") == nil {
		if err := p.validateName(); err != nil {
			return err
		}
	}

	return p.validateColumnConstraintAux([]string{})
}

func (p *sqliteParser) validateColumnConstraintAux(ls []string) error {
	if p.matchKeyword("PRIMARY") {
		if contains(ls, "PRIMARY") {
			return p.syntaxError()
		}
		if err := p.validateConstraintPrimaryKey(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "PRIMARY"))
	}

	if p.matchKeyword("NOT") {
		if contains(ls, "NOT") {
			return p.syntaxError()
		}
		if err := p.validateConstraintNotNull(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "NOT"))
	}

	if p.matchKeyword("UNIQUE") {
		if contains(ls, "UNIQUE") {
			return p.syntaxError()
		}
		if err := p.validateConstraintUnique(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "UNIQUE"))
	}

	if p.matchKeyword("CHECK") {
		if contains(ls, "CHECK") {
			return p.syntaxError()
		}
		if err := p.validateConstraintCheck(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "CHECK"))
	}

	if p.matchKeyword("DEFAULT") {
		if contains(ls, "DEFAULT") {
			return p.syntaxError()
		}
		if err := p.validateConstraintDefault(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "DEFAULT"))
	}

	if p.matchKeyword("COLLATE") {
		if contains(ls, "COLLATE") {
			return p.syntaxError()
		}
		if err := p.validateConstraintCollate(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "COLLATE"))
	}

	if p.matchKeyword("REFERENCES") {
		if contains(ls, "REFERENCES") {
			return p.syntaxError()
		}
		if err := p.validateConstraintForeignKey(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "REFERENCES"))
	}

	if p.matchKeyword("GENERATED", "AS") {
		if contains(ls, "GENERATED") {
			return p.syntaxError()
		}
		if err := p.validateConstraintGenerated(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(append(ls, "GENERATED"))
	}

	return nil
}

func (p *sqliteParser) validateConstraintPrimaryKey() error {
	if err := p.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	if p.matchKeyword("ASC", "DESC") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	if p.matchKeyword("AUTOINCREMENT") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	return nil
}

func (p *sqliteParser) validateConstraintNotNull() error {
	if err := p.validateKeyword("NOT"); err != nil {
		return err
	}
	if err := p.validateKeyword("NULL"); err != nil {
		return err
	}
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateConstraintUnique() error {
	if err := p.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateConstraintCheck() error {
	if err := p.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateConstraintDefault() error {
	if err := p.validateKeyword("DEFAULT"); err != nil {
		return err
	}
	if p.matchSymbol("(") {
		if err := p.validateExpr(); err != nil {
			return err
		}
	} else {
		if err := p.validateLiteralValue(); err != nil {
			return err
		}
	}
	return nil
}

func (p *sqliteParser) validateConstraintCollate() error {
	if err := p.validateKeyword("COLLATE"); err != nil {
		return err
	}
	if err := p.validateKeyword(CollatingFunction_SQLite...); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateConstraintForeignKey() error {
	if err := p.validateKeyword("REFERENCES"); err != nil {
		return err
	}
	if err := p.validateTableName(); err != nil {
		return err
	}
	if p.validateSymbol("(") == nil {
		if err := p.validateCommaSeparatedColumnNames(); err != nil {
			return err
		}
		if err := p.validateSymbol(")"); err != nil {
			return err
		}
	}
	if err := p.validateConstraintForeignKeyAux(); err != nil {
		return p.syntaxError()
	}
	return nil
}

func (p *sqliteParser) validateConstraintForeignKeyAux() error {
	if p.validateKeyword("ON") == nil {
		if err := p.validateKeyword("DELETE", "UPDATE"); err != nil {
			return err
		}
		if p.validateKeyword("SET") == nil {
			if err := p.validateKeyword("NULL", "DEFAULT"); err != nil {
				return err
			}
		} else if p.validateKeyword("CASCADE", "RESTRICT") == nil {

		} else if p.validateKeyword("NO") == nil {
			if err := p.validateKeyword("ACTION"); err != nil {
				return err
			}
		} else {
			return p.syntaxError()
		}
		return p.validateConstraintForeignKeyAux()
	}

	if p.validateKeyword("MATCH") == nil {
		if err := p.validateKeyword("SIMPLE", "PARTIAL", "FULL"); err != nil {
			return err
		}
		return p.validateConstraintForeignKeyAux()
	}

	if p.matchKeyword("NOT", "DEFERRABLE") {
		if p.matchKeyword("NOT") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateKeyword("DEFERRABLE"); err != nil {
			return err
		}
		if p.validateKeyword("INITIALLY") == nil {
			if err := p.validateKeyword("DEFERRED", "IMMEDIATE"); err != nil {
				return err
			}
		}
		return p.validateConstraintForeignKeyAux()
	}

	return nil
}

func (p *sqliteParser) validateConstraintGenerated() error {
	if p.validateKeyword("GENERATED") == nil {
		if err := p.validateKeyword("ALWAYS"); err != nil {
			return err
		}
	}
	if err := p.validateKeyword("AS"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	if p.matchKeyword("STORED", "VIRTUAL") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	return nil
}

func (p *sqliteParser) validateConflictClause() error {
	if p.validateKeyword("ON") == nil {
		if err := p.validateKeyword("CONFLICT"); err != nil {
			return err
		}
		if err := p.validateKeyword(ConflictAction_SQLite...); err != nil {
			return err
		}
	}
	return nil
}

func (p *sqliteParser) validateExpr() error {
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateExprAux(); err != nil {
		return err
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateExprAux() error {
	if p.matchSymbol(")") {
		return nil
	}
	if p.matchSymbol("(") {
		if err := p.validateExpr(); err != nil {
			return err
		}
		return p.validateExprAux()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return p.validateExprAux()
}

func (p *sqliteParser) validateLiteralValue() error {
	if isNumeric(p.token()) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	if p.matchSymbol("'") {
		return p.validateStringLiteral()
	}
	return p.validateKeyword(LiteralValue_SQLite...)
}

func (p *sqliteParser) validateStringLiteral() error {
	if err := p.validateSymbol("'"); err != nil {
		return err
	}
	if err := p.validateStringLiteralAux(); err != nil {
		return err
	}
	if err := p.validateSymbol("'"); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateStringLiteralAux() error {
	if p.matchSymbol("'") {
		return nil
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return p.validateStringLiteralAux()
}

func (p *sqliteParser) validateTableConstraint() error {
	if p.validateKeyword("CONSTRAINT") == nil{
		if err := p.validateName(); err != nil {
			return err
		}
	}
	return p.validateTableConstraintAux()
}

func (p *sqliteParser) validateTableConstraintAux() error {
	if p.matchKeyword("PRIMARY") {
		return p.validateTablePrimaryKey()
	}

	if p.matchKeyword("UNIQUE") {
		return p.validateTableUnique()
	}

	if p.matchKeyword("CHECK") {
		return p.validateTableCheck()
	}

	return p.syntaxError()
}

func (p *sqliteParser) validateTablePrimaryKey() error {
	if err := p.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateCommaSeparatedColumnNames(); err != nil {
		return p.syntaxError()
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateTableUnique() error {
	if err := p.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	if err := p.validateSymbol("("); err != nil {
		return err
	}
	if err := p.validateCommaSeparatedColumnNames(); err != nil {
		return p.syntaxError()
	}
	if err := p.validateSymbol(")"); err != nil {
		return err
	}
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateTableCheck() error {
	if err := p.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	return nil
}

func (p *sqliteParser) validateCommaSeparatedColumnNames() error {
	if err := p.validateColumnName(); err != nil {
		return err
	}
	if p.matchSymbol(",") {
		if p.next() != nil {
			return p.syntaxError()
		}
		return p.validateCommaSeparatedColumnNames()
	}
	return nil
}

func (p *sqliteParser) Parse() ([]Table, error) {
	p.init()
	var tables []Table
	return tables, nil
}

var DataType_SQLite = []string{
	"TEXT",
	"NUMERIC",
	"INTEGER",
	"REAL",
	"NONE",
}

var ConflictAction_SQLite = []string{
	"ROLLBACK",
	"ABORT",
	"FAIL",
	"IGNORE",
	"REPLACE",
}

var CollatingFunction_SQLite = []string{
	"BINARY",
	"NOCASE",
	"RTRIM",
}

var LiteralValue_SQLite = []string{
	"NULL",
	"TRUE",
	"FALSE",
	"CURRENT_TIME",
	"CURRENT_DATE",
	"CURRENT_TIMESTAMP",
}

var TableConstraint_SQLite = []string{
	"CONSTRAINT",
	"PRIMARY",
	"UNIQUE",
	"CHECK",
	"FOREIGN",
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