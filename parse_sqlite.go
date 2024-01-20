package ddlparse

import (
	"fmt"
	"strconv"
	"errors"
	"regexp"
	"strings"
)

type sqliteParser struct {
	ddl string
	tokens []string
	validatedTokens []string
	size int
	i int
	line int
	flg bool
	tables []Table
}

func newSQLiteParser(ddl string) parser {
	return &sqliteParser{ddl: ddl}
}


func (p *sqliteParser) token() string {
	return p.tokens[p.i]
}


func (p *sqliteParser) appendToken(token string) {
	if (token != "") {
		p.tokens = append(p.tokens, token)
	}
}


func (p *sqliteParser) isOutOfRange() bool {
	return p.i > p.size - 1
}


/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  VALIDATE: 
    Check the syntax of DDL (tokens). 
    And eliminate unnecessary tokens during parsing.
	Return an ValidateError if validation fails.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

Example:

***** tokens *****
[CREATE TABLE IF NOT users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

***** validatedTokens *****
[CREATE TABLE users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

func (p *sqliteParser) Validate() error {
	tokens, err := Tokenize(p.ddl, SQLite)
	if err != nil {
		return err
	}
	p.tokens = tokens
	return p.validate()
}


func (p *sqliteParser) validate() error {
	p.initV()
	return p.validateProc()
}


func (p *sqliteParser) initV() {
	p.validatedTokens = []string{}
	p.i = -1
	p.line = 1
	p.size = len(p.tokens)
	p.flg = false
	p.next()
}


func (p *sqliteParser) flgOn() {
	p.flg = true
}


func (p *sqliteParser) flgOff() {
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
	} else {
		return nil
	}
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


func (p *sqliteParser) validateProc() error {
	if (p.isOutOfRange()) {
		return nil
	}
	if err := p.validateCreateTable(); err != nil {
		return err
	}
	return p.validateProc()
}


func (p *sqliteParser) syntaxError() error {
	if p.isOutOfRange() {
		return NewValidateError(p.line, p.tokens[p.size - 1])
	}
	return NewValidateError(p.line, p.tokens[p.i])
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


func (p *sqliteParser) isValidName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return pattern.MatchString(name) && 
		!contains(ReservedWords_SQLite, strings.ToUpper(name))
}


func (p *sqliteParser) isValidQuotedName(name string) bool {
	return true
}


func (p *sqliteParser) validateName() error {
	if isQuotedToken(p.token()) {
		if !p.isValidQuotedName(p.token()) {
			return p.syntaxError()
		}
		if p.next() != nil {
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
	if isNumericToken(p.token()) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	if isQuotedToken(p.token()) {
		if p.next() != nil {
			return p.syntaxError()
		}
		return nil
	}
	ls := []string{"NULL", "TRUE", "FALSE", "CURRENT_TIME", "CURRENT_DATE", "CURRENT_TIMESTAMP"}
	if err := p.validateKeyword(ls...); err != nil {
		return err
	}
	return nil
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


/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  PARSE: 
    Tokenize, validate and convert to a Table object and return it.
	Return an ParseError if there are errors other than syntax errors in the DDL.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

Example:

***** ddl *****
"CREATE TABLE IF NOT users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	password TEXT NOT NULL, --hashing
	created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
	updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
	UNIQUE(name)
);"

***** Table *****
{ 
	Name: users 
	Columns: [
		{ Name: id, DataType: INTEGER, IsPK: true, IsNotNull: false, 
			IsUnique: false, IsAutoIncrement: true, Default: nil },
		{ Name: name, DataType: TEXT, IsPK: false, IsNotNull: true, 
			IsUnique: true, IsAutoIncrement: false, Default: nil },
		{ Name: password, DataType: TEXT, IsPK: false, IsNotNull: true, 
			IsUnique: true, IsAutoIncrement: false, Default: nil },
		{ Name: created_at, DataType: TEXT, IsPK: false, IsNotNull: true, 
			IsUnique: true, IsAutoIncrement: false, Default: (DATETIME('now','localtime') },
		{ Name: updated_at, DataType: TEXT, IsPK: false, IsNotNull: true, 
			IsUnique: true, IsAutoIncrement: false, Default: (DATETIME('now','localtime') },
	]
}

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

func (p *sqliteParser) Parse() ([]Table, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p.tables, nil
}


func (p *sqliteParser) parse() error {
	p.initP()
	return p.parseProc()
}


func (p *sqliteParser) initP() {
	p.tables = []Table{}
	p.i = 0
	p.line = 0
	p.size = len(p.validatedTokens)
	p.tokens = p.validatedTokens
}


func (p *sqliteParser) parseProc() error {
	if p.isOutOfRange() {
		return nil
	} else {
		table, err := p.parseTable()
		if err != nil {
			return err
		}
		p.tables = append(p.tables, table)
	}

	return p.parseProc()
}


func (p *sqliteParser) parseTable() (Table, error) {
	var table Table
	p.i += 2

	schemaName, tableName := p.parseTableName()
	table.Schema = schemaName
	table.Name = tableName

	columns, err := p.parseColumns()
	if err != nil {
		return Table{}, err
	}
	table.Columns = columns

	if (p.size > p.i) {
		if p.matchSymbol(";") {
			p.i += 1
		}
	}
	return table, nil
}


func (p *sqliteParser) parseTableName() (string, string) {
	schemaName := p.parseName()
	tableName := ""

	if p.token() == "." {
		p.i += 1
		tableName = p.parseName()
	} else {
		tableName = schemaName
		schemaName = ""
	}

	return schemaName, tableName 
}


func (p *sqliteParser) parseName() string {
	token := p.token()
	p.i += 1
	if isQuotedToken(token) {
		return token[1 : len(token)-1]
	} else {
		return token
	}
}


func (p *sqliteParser) parseColumns() ([]Column, error) {
	p.i += 1
	var columns []Column
	for !p.matchSymbol(")") {
		if p.matchSymbol(",") {
			p.i += 1
			continue
		}
		var err error
		if (p.matchKeyword("PRIMARY", "UNIQUE")) {
			err = p.parseTableConstraint(&columns)
		} else {
			err = p.parseColumn(&columns)
		}

		if err != nil {
			return nil, err
		}
	}
	p.i += 1
	return columns, nil
}


func (p *sqliteParser) parseColumn(columns *[]Column) error {
	name := p.parseName()

	for _, column := range *columns {
		if column.Name == name {
			return NewParseError(fmt.Sprintf("Duplicate column name: '%s'.", name))
		}
	}
	
	var column Column
	column.Name = name
	column.DataType = p.token()
	p.i += 1
	p.parseConstraint(&column)

	*columns = append(*columns, column)
	return nil
}


func (p *sqliteParser) parseConstraint(column *Column) error {
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
		return p.parseConstraint(column)
	}

	if p.matchKeyword("NOT") {
		p.i += 2
		column.IsNotNull = true
		return p.parseConstraint(column)
	}

	if p.matchKeyword("UNIQUE") {
		p.i += 1
		column.IsUnique = true
		return p.parseConstraint(column)
	}

	if p.matchKeyword("DEFAULT") {
		p.i += 1
		column.Default = p.parseDefaultValue()
		return p.parseConstraint(column)
	}

	return errors.New("program error")
}


func (p *sqliteParser) parseDefaultValue() interface{} {
	if p.matchSymbol("(") {
		expr := ""
		p.parseExpr(&expr)
		return expr
	} else {
		return p.parseLiteralValue()
	}
}

func (p *sqliteParser) parseExpr(expr *string) {
	*expr +=  p.token()
	p.i += 1
	p.parseExprAux(expr)
	*expr +=  p.token()
	p.i += 1
}


func (p *sqliteParser) parseExprAux(expr *string) {
	if p.matchSymbol(")") {
		return
	}
	if p.matchSymbol("(") {
		p.parseExpr(expr)
		return
	}
	*expr += p.token()
	p.i += 1
	p.parseExprAux(expr)
}


func (p *sqliteParser) parseLiteralValue() interface{} {
	token := p.token()
	p.i += 1
	if isNumericToken(token) {
		n, _ := strconv.ParseFloat(token, 64)
		return n
	}
	if isQuotedToken(token) {
		return token[1 : len(token)-1]
	}
	return token
}


func (p *sqliteParser) parseStringValue() interface{} {
	token := p.token()
	if isNumericToken(token) {
		p.i += 1
		n, _ := strconv.ParseFloat(token, 64)
		return n
	}
	if isQuotedToken(token) {
		p.i += 1
		return token[1 : len(token)-1]
	}
	p.i += 1
	return token
}


func (p *sqliteParser) parseTableConstraint(columns *[]Column) error {
	c := strings.ToUpper(p.token())
	if p.matchKeyword("PRIMARY") {
		p.i += 2
	}
	if p.matchKeyword("UNIQUE") {
		p.i += 1
	}

	columnNames, err := p.parseCommaSeparatedColumnNames()
	if err != nil {
		return err
	}

	for _, name := range columnNames {
		exists := false
		for i, column := range *columns {
			if column.Name != name {
				continue
			}
			exists = true
			if c == "PRIMARY" {
				if column.IsPK {
					return NewParseError(fmt.Sprintf("Multiple primary key defined: '%s'.", name))
				}
				(*columns)[i].IsPK = true
				break
			}
			if c == "UNIQUE" {
				if column.IsUnique {
					return NewParseError(fmt.Sprintf("Multiple unique constraint defined: '%s'.", name))
				}
				(*columns)[i].IsUnique = true
				break
			}
		}
		if !exists {
			return NewParseError(fmt.Sprintf("Unknown column: '%s'.", name))
		}
	}
	return nil
}


func (p *sqliteParser) parseCommaSeparatedColumnNames() ([]string, error) {
	p.i += 1
	ls := []string{}
	for {
		ls = append(ls, p.token())
		p.i += 1
		if p.matchSymbol(")") {
			break
		} else if p.matchSymbol(",") {
			p.i += 1
			continue
		} else {
			return nil, errors.New("program error")
		}
	}
	p.i += 1
	return ls, nil
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