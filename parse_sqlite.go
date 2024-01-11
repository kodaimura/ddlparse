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

func (p *sqliteParser) validateName() error {
	if (p.token() == "\"") {
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
		if !p.isValidQuotedName(p.token()) {
			return p.syntaxError()
		}
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
		if (p.token() != "\"") {
			return p.syntaxError()
		}
	} else if (p.token() == "`") {
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
		if !p.isValidQuotedName(p.token()) {
			return p.syntaxError()
		}
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
		if (p.token() != "`") {
			return p.syntaxError()
		}
	} else {
		if !p.isValidName(p.token()) {
			return p.syntaxError()
		}
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return nil
}

func (p *sqliteParser) validateCreateTable() error {
	if (p.token() != "create" && p.token() != "CREATE") {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	if (p.token() != "table" && p.token() != "TABLE") {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	if (p.token() == "if" || p.token() == "IF") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if (p.token() != "not" && p.token() != "NOT") {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
		if (p.token() != "exists" && p.token() != "EXISTS") {
			return p.syntaxError()
		}
	}
	if p.next() != nil {
		return p.syntaxError()
	}

	if err := p.validateName(); err != nil {
		return err
	}
	if (p.token() != "(") {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}

	if err := p.validateColumns(); err != nil {
		return err
	}
	if (p.token() != ")") {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	p.next()

	return nil
}

func (p *sqliteParser) validateTableName() error {
	return p.validateName()
}

func (p *sqliteParser) validateColumns() error {
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if err := p.validateColumn(); err != nil {
		return err
	}
	if (p.token() != ",") {
		return p.validateColumns()
	}
	return nil
}

func (p *sqliteParser) validateColumn() error {
	if err := p.validateColumnName(); err != nil {
		return err
	}
	if err := p.validateColumnType(); err != nil {
		return err
	}
	if err := p.validateColumnConstraint(); err != nil {
		return err
	}
	return p.validateColumns()
}

func (p *sqliteParser) validateColumnName() error {
	return p.validateName()
}

//Omitting data types is not supported.
func (p *sqliteParser) validateColumnType() error {
	if !contains(DataType_SQLite, strings.ToUpper(p.token())) {
		return p.syntaxError()
	}
	if p.next() != nil {
		return p.syntaxError()
	}
	return nil
}

func (p *sqliteParser) validateColumnConstraint() error {
	if p.token() == "CONSTRAINT" || p.token() == "constraint" {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateName(); err != nil {
			return err
		}
	}

	return p.validateColumnConstraintAux([]string{"pk", "nn", "uq", "ck", "de", "co", "fk", "ga"})
}

func (p *sqliteParser) validateColumnConstraintAux(ls []string) error {
	if p.token() == "PRIMARY" || p.token() == "primary" {
		if !contains(ls, "pk") {
			return p.syntaxError()
		}
		if err := p.validateConstraintPrimaryKey(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(remove(ls, "pk"))
	}

	if p.token() == "NOT" || p.token() == "not" {
		if !contains(ls, "nn") {
			return p.syntaxError()
		}
		if err := p.validateConstraintNotNull(); err != nil {
			return err
		}
		return p.validateColumnConstraintAux(remove(ls, "nn"))
	}
	return nil
}

func (p *sqliteParser) validateConstraintPrimaryKey() error {
	if p.token() == "PRIMARY" || p.token() == "primary" {
		if p.next() != nil {
			return p.syntaxError()
		}
		if !(p.token() == "KEY" || p.token() == "key") {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
		if (p.token() == "ASC" || p.token() == "DESC" || p.token() == "asc" || p.token() == "desc") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
		if err := p.validateConflictClause(); err != nil {
			return err
		}
		if (p.token() == "AUTOINCREMENT" || p.token() == "autoincrement") {
			if p.next() != nil {
				return p.syntaxError()
			}
		}
	}
	return nil
}

func (p *sqliteParser) validateConstraintNotNull() error {
	if p.token() == "NOT" || p.token() == "not" {
		if p.next() != nil {
			return p.syntaxError()
		}
		if !(p.token() == "NULL" || p.token() == "null") {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateConflictClause(); err != nil {
			return err
		}
	}
	return nil
}

func (p *sqliteParser) validateConflictClause() error {
	if p.token() == "ON" || p.token() == "on" {
		if p.next() != nil {
			return p.syntaxError()
		}
		if !(p.token() == "CONFLICT" || p.token() == "conflict") {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
		if !contains(ConflictAction_SQLite, strings.ToUpper(p.token())) {
			return p.syntaxError()
		}
		if p.next() != nil {
			return p.syntaxError()
		}
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