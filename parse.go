package ddlparse

import (
	"fmt"
	"errors"
	"strings"
	"strconv"
)

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

func parse (tokens []string, rdbms Rdbms) ([]Table, error) {
	p := newParser(rdbms, tokens)
	return p.Parse()
}

type parserI interface {
	Parse() ([]Table, error)
}

type parser struct {
	rdbms Rdbms
	tokens []string
	size int
	i int
	tables []Table
}

func newParser(rdbms Rdbms, tokens []string) parserI {
	return &parser{tokens: tokens, rdbms: rdbms}
}

func (p *parser) Parse() ([]Table, error) {
	p.initP()
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p.tables, nil
}

func (p *parser) initP() {
	p.i = 0
	p.size = len(p.tokens)
	p.tables = []Table{}
}

func (p *parser) token() string {
	return p.tokens[p.i]
}

func (p *parser) isOutOfRange() bool {
	return p.i > p.size - 1
}

func (p *parser) matchKeyword(keywords ...string) bool {
	return contains(
		append(
			mapSlice(keywords, strings.ToLower), 
			mapSlice(keywords, strings.ToUpper)...,
		), p.token())
}


func (p *parser) matchSymbol(symbols ...string) bool {
	return contains(symbols, p.token())
}

func (p *parser) isIdentifier(token string) bool {
	c := token[0:1]
	switch (p.rdbms) {
		case SQLite:
			return c == "\"" || c == "`"
		case MySQL:
			return c == "`"
		case PostgreSQL:
			return c == "\""
	}
	return false
}

func (p *parser) isStringValue(token string) bool {
	c := token[0:1]
	switch (p.rdbms) {
		case SQLite:
			return c == "'"
		case MySQL:
			return c == "\"" || c == "'"
		case PostgreSQL:
			return c == "'"
	}
	return false
}

func (p *parser) parse() error {
	if p.isOutOfRange() {
		return nil
	} else {
		table, err := p.parseTable()
		if err != nil {
			return err
		}
		p.tables = append(p.tables, table)
	}

	return p.parse()
}


func (p *parser) parseTable() (Table, error) {
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


func (p *parser) parseTableName() (string, string) {
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


func (p *parser) parseName() string {
	token := p.token()
	p.i += 1
	if p.isIdentifier(token) {
		return token[1 : len(token)-1]
	} else {
		return token
	}
}


func (p *parser) parseColumns() ([]Column, error) {
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


func (p *parser) parseColumn(columns *[]Column) error {
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


func (p *parser) parseConstraint(column *Column) error {
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


func (p *parser) parseDefaultValue() interface{} {
	if p.matchSymbol("(") {
		expr := ""
		p.parseExpr(&expr)
		return expr
	} else {
		return p.parseLiteralValue()
	}
}

func (p *parser) parseExpr(expr *string) {
	*expr +=  p.token()
	p.i += 1
	p.parseExprAux(expr)
	*expr +=  p.token()
	p.i += 1
}


func (p *parser) parseExprAux(expr *string) {
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


func (p *parser) parseLiteralValue() interface{} {
	token := p.token()
	p.i += 1
	if isNumericToken(token) {
		n, _ := strconv.ParseFloat(token, 64)
		return n
	}
	if p.isStringValue(token) {
		return token[1 : len(token)-1]
	}
	return token
}


func (p *parser) parseTableConstraint(columns *[]Column) error {
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


func (p *parser) parseCommaSeparatedColumnNames() ([]string, error) {
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