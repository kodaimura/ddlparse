package ddlparse

import (
	"strconv"
	"errors"
	"regexp"
	"strings"
)

type sqliteParser struct {
	tokens []string
	size int
	i int
	line int
	flg bool
	validatedTokens []string
	tables []Table
}

func newSQLiteParser(tokens []string) parser {
	return &sqliteParser{tokens, 0, 0, 0, false, []string{}}
}

func (p *sqliteParser) Parse() ([]Table, error) {
	p.initV()
	if err := p.validate(); err != nil {
		return nil, err
	}
	p.initP()
	return p.parse()
}

func (p *sqliteParser) Validate() error {
	p.initV()
	return p.validate()
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

func (p *sqliteParser) flgOn() {
	p.flg = true
}

func (p *sqliteParser) flgOff() {
	p.flg = false
}

func (p *sqliteParser) initV() {
	p.i = -1
	p.line = 1
	p.size = len(tokens)
	p.validatedTokens = []string{}
	p.flg = false
	p.next()
}

func (p *sqliteParser) initP() {
	p.i = 0
	p.line = 0
	p.size = len(p.validatedTokens)
	p.validatedTokens = []string{}
	p.flg = false
}

func (p *sqliteParser) next() error {
	if p.flg {
		p.validatedTokens = append(p.validatedTokens, p.token())
	}
	return p.nextAux()
}

func (p *sqliteParser) nextAux() error {
	p.i += 1
	if (p.isOutOfRange()) {
		return errors.New("out of range")
	}
	if (p.token() == "\n") {
		p.line += 1
		return p.nextAux()
	} else if (p.token() == "--") {
		p.skipSingleLineComment()
		return p.nextAux()
	} else if (p.token() == "/*") {
		if err := p.skipMultiLineComment(); err != nil {
			return err
		}
		return p.nextAux()
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
	p.flgOn()
	if err := p.validateKeyword("CREATE"); err != nil {
		return err
	}
	if err := p.validateKeyword("TABLE"); err != nil {
		return err
	}

	p.flgOff()
	if p.validateKeyword("IF") == nil {
		if err := p.validateKeyword("NOT"); err != nil {
			return err
		}
		if err := p.validateKeyword("EXISTS"); err != nil {
			return err
		}
	}

	p.flgOn()
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
	if err := p.validateTableOptions(); err != nil {
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
	if p.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
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
	return p.validateKeyword("TEXT", "NUMERIC", "INTEGER", "REAL", "NONE")
}

func (p *sqliteParser) validateColumnConstraint() error {
	p.flgOff()
	if p.validateKeyword("CONSTRAINT") == nil {
		if err := p.validateName(); err != nil {
			return err
		}
	}
	p.flgOn()
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
	p.flgOn()
	if err := p.validateKeyword("PRIMARY"); err != nil {
		return err
	}
	if err := p.validateKeyword("KEY"); err != nil {
		return err
	}
	p.flgOff()
	if p.matchKeyword("ASC", "DESC") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	p.flgOn()
	if p.matchKeyword("AUTOINCREMENT") {
		if p.next() != nil {
			return p.syntaxError()
		}
	}
	return nil
}

func (p *sqliteParser) validateConstraintNotNull() error {
	p.flgOn()
	if err := p.validateKeyword("NOT"); err != nil {
		return err
	}
	if err := p.validateKeyword("NULL"); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateConstraintUnique() error {
	p.flgOn()
	if err := p.validateKeyword("UNIQUE"); err != nil {
		return err
	}
	p.flgOff()
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateConstraintCheck() error {
	p.flgOff()
	if err := p.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateConstraintDefault() error {
	p.flgOn()
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
	p.flgOff()
	if err := p.validateKeyword("COLLATE"); err != nil {
		return err
	}
	if err := p.validateKeyword("BINARY","NOCASE", "RTRIM"); err != nil {
		return err
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateConstraintForeignKey() error {
	p.flgOff()
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
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateConstraintForeignKeyAux() error {
	p.flgOff()
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

	p.flgOn()
	return nil
}

func (p *sqliteParser) validateConstraintGenerated() error {
	p.flgOff()
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
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateConflictClause() error {
	p.flgOff()
	if p.validateKeyword("ON") == nil {
		if err := p.validateKeyword("CONFLICT"); err != nil {
			return err
		}
		if err := p.validateKeyword("ROLLBACK", "ABORT", "FAIL", "IGNORE","REPLACE"); err != nil {
			return err
		}
	}
	p.flgOn()
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
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := p.validateKeyword(ls...); err != nil {
		return err
	}
	return nil
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
	p.flgOff()
	if p.validateKeyword("CONSTRAINT") == nil{
		if err := p.validateName(); err != nil {
			return err
		}
	}
	p.flgOn()
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

	if p.matchKeyword("FOREIGN") {
		return p.validateTableForeignKey()
	}

	return p.syntaxError()
}

func (p *sqliteParser) validateTablePrimaryKey() error {
	p.flgOn()
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
	p.flgOff()
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateTableUnique() error {
	p.flgOn()
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
	p.flgOff()
	if err := p.validateConflictClause(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateTableCheck() error {
	p.flgOff()
	if err := p.validateKeyword("CHECK"); err != nil {
		return err
	}
	if err := p.validateExpr(); err != nil {
		return err
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) validateTableForeignKey() error {
	p.flgOff()
	if err := p.validateKeyword("FOREIGN"); err != nil {
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
	if err := p.validateConstraintForeignKey(); err != nil {
		return err
	}
	p.flgOn()
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

func (p *sqliteParser) validateTableOptions() error {
	p.flgOff()
	if p.matchKeyword("WITHOUT") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if err := p.validateKeyword("ROWID"); err != nil {
			return err
		}
		if p.matchSymbol(",") {
			if p.next() != nil {
				return p.syntaxError()
			}
			if err := p.validateKeyword("STRICT"); err != nil {
				return err
			}
		}
	} else if p.matchKeyword("STRICT") {
		if p.next() != nil {
			return p.syntaxError()
		}
		if p.matchSymbol(",") {
			if p.next() != nil {
				return p.syntaxError()
			}
			if err := p.validateKeyword("WITHOUT"); err != nil {
				return err
			}
			if err := p.validateKeyword("ROWID"); err != nil {
				return err
			}
		}
	}
	p.flgOn()
	return nil
}

func (p *sqliteParser) parse() error {
	if (p.size <= p.i) {
		return tables
	} else {
		table, err := p.parseTable()
		if err != nil {
			return err
		}
		p.tables = append(p.tables, table)
	}

	return p.parse()
}

func (p *sqliteParser) parseTable() (Table, err) {
	var table Table
	p.i += 2

	schemaName, tableName = p.parseTableName()

	table.Schema = schemaName
	table.Name = tableName

	columns, err := p.parseColumns()
	if err != nil {
		return nil, err
	}
	table.Columns = columns

	return table
}

func (p *sqliteParser) parseTableName() (string, string) {
	schemaName := ""
	tableName := ""

	tmp := ""
	if p.matchSymbol("\"", "`") {
		p.i += 1
		tmp = p.token()
		p.i += 1
	} else {
		tmp = p.token()
	}

	if p.token() == "." {
		p.i += 1
		schemaName = tmp
		if p.matchSymbol("\"", "`") {
			p.i += 1
			tableName = p.token()
			p.i += 1
		} else {
			tableName = p.token()
		}
	} else {
		tableName = tmp
	}

	return schemaName, tableName 
}

func (p *sqliteParser) parseColumns() ([]Column, error) {
	p.i += 1
	var columns []Column
	for p.matchSymbol(")") {
		if p.matchSymbol(",") {
			p.i += 1
			continue
		}
		err := nil
		if (p.matchKeyword("PRIMARY", "UNIQUE", "NOT", "DEFAULT")) {
			err = p.parseTableConstraint(columns)
		} else {
			err = p.parseColumn(columns)
		}

		if err != nil {
			return err
		}
	}
	p.i += 1
	return columns
}

func (p *sqliteParser) parseColumn(columns []Column) error {
	name := ""
	if p.matchSymbol("\"", "`") {
		p.i += 1
		name = p.token()
		p.i += 1
	} else {
		name = p.token()
	}

	for _, column := range columns {
		if column.Name == name {
			return errors.New("")
		}
	}
	
	var column Column
	column.Name = name
	p.i += 1
	column.DataType = p.token()
	p.i += 1
	p.parseConstraint(&column)

	columns = append(columns, column)
}

func (p *sqliteParser) parseConstrainte(column *Column) error {
	if p.matchSymbol(",") {
		p.i += 1
		return nil
	}
	if p.matchSymbol(")") {
		return nil
	}

	if p.matchKeyword("PRIMARY") {
		p.i += 2
		column.IsPK = true
		if p.matchKeyword("AUTOINCREMENT") {
			p.i += 1
			column.IsAutoIncrement = true
		}
		return p.parseConstraint(&column)
	}

	if p.matchKeyword("NOT") {
		p.i += 2
		column.IsNotNull = true
		return p.parseConstraint(&column)
	}

	if p.matchKeyword("UNIQUE") {
		p.i += 1
		column.IsUnique = true
		return p.parseConstraint(&column)
	}

	if p.matchKeyword("DEFAULT") {
		p.i += 1
		column.Default = p.parseDefaultValue()
		return p.parseConstraint(&column)
	}

	return errors.New("")
}

func (p *sqliteParser) parseDefaultValue() interface{} {
	if p.matchSymbol("(") {
		p.skipExpr()
		return func(){}
	} else {
		return p.parseLiteralValue()
	}
}

func (p *sqliteParser) skipExpr() {
	p.i += 1
	p.parseExprAux()
	p.i += 1
}

func (p *sqliteParser) skipExprAux() {
	if p.matchSymbol(")") {
		return
	}
	if p.matchSymbol("(") {
		p.skipExpr()
	}
	p.i += 1
	p.skipExprAux()
}

func (p *sqliteParser) parseLiteralValue() interface{} {
	token := p.token()
	if isNumeric(token) {
		p.i += 1
		n, _ := strconv.ParseFloat(token, 64)
		return n
	}
	if p.matchSymbol("'") {
		return p.parseStringLiteral()
	}
	p.i += 1
	return token
}

func (p *sqliteParser) parseStringLiteral() string {
	p.i += 1
	ret := ""
	for p.token() != "'" {
		ret += " " + p.token()
	}
	p.i += 1

	return ret
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