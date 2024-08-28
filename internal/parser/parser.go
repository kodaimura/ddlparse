package parser

import (
	"errors"
	"strings"
	"strconv"

	"github.com/kodaimura/ddlparse/internal/types"
	"github.com/kodaimura/ddlparse/internal/common"
)

/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  PARSE: 
    Convert the validated token to a Table object.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

type Parser interface {
	Parse() []types.Table
}

type parser struct {
	rdbms common.Rdbms
	tokens []string
	size int
	i int
	tables []types.Table
}


func NewParser(rdbms common.Rdbms, tokens []string) Parser {
	return &parser{tokens: tokens, rdbms: rdbms}
}


func (p *parser) Parse() []types.Table {
	p.init()
	p.parse()
	return p.tables
}


func (p *parser) init() {
	p.i = 0
	p.size = len(p.tokens)
	p.tables = []types.Table{}
}


func (p *parser) token() string {
	return p.tokens[p.i]
}


func (p *parser) isOutOfRange() bool {
	return p.i > p.size - 1
}


func (p *parser) next() error {
	pre := p.token()
	p.i += 1
	if (p.isOutOfRange()) {
		return errors.New("out of range")
	}
	if pre == "," && p.token() == "," {
		return p.next()
	}
	return nil
}


func (p *parser) matchKeyword(keywords ...string) bool {
	return common.Contains(
		append(
			common.MapSlice(keywords, strings.ToLower), 
			common.MapSlice(keywords, strings.ToUpper)...,
		), p.token())
}


func (p *parser) matchSymbol(symbols ...string) bool {
	return common.Contains(symbols, p.token())
}


func (p *parser) isIdentifier(token string) bool {
	c := token[0:1]
	switch (p.rdbms) {
		case common.SQLite:
			return c == "\"" || c == "`"
		case common.MySQL:
			return c == "`"
		case common.PostgreSQL:
			return c == "\""
	}
	return false
}


func (p *parser) isStringValue(token string) bool {
	c := token[0:1]
	switch (p.rdbms) {
		case common.SQLite:
			return c == "'"
		case common.MySQL:
			return c == "\"" || c == "'"
		case common.PostgreSQL:
			return c == "'"
	}
	return false
}


func (p *parser) parse() {
	if p.isOutOfRange() {
		return
	} else {
		table := p.parseTable()
		p.tables = append(p.tables, table)
	}
	p.parse()
}


func (p *parser) parseToken() string {
	token := p.token()
	p.next()
	return token
}


func (p *parser) parseTable() types.Table {
	var table types.Table
	p.next() // skip "CREATE"
	p.next() // skip "TABLE"

	if p.matchKeyword("IF") {
		p.next() // skip "IF"
		p.next() // skip "NOT"
		p.next() // skip "EXISTS"
		table.IfNotExists = true
	}

	schemaName, tableName := p.parseTableName()
	columns, constraints := p.parseTableDefinition();

	table.Schema = schemaName
	table.Name = tableName
	table.Columns = columns
	table.Constraints = constraints

	if (p.size > p.i) {
		if p.matchSymbol(";") {
			p.next()
		}
	}
	return table
}


func (p *parser) parseTableName() (string, string) {
	schemaName := p.parseName()
	tableName := ""

	if p.matchSymbol(".") {
		p.next()
		tableName = p.parseName()
	} else {
		tableName = schemaName
		schemaName = ""
	}

	return schemaName, tableName 
}


func (p *parser) parseName() string {
	token := p.parseToken()
	if p.isIdentifier(token) {
		return token[1 : len(token)-1]
	} else {
		return token
	}
}


func (p *parser) parseTableDefinition() ([]types.Column, types.TableConstraint) {
	p.next()
	var columns []types.Column
	var constraints types.TableConstraint
	for !p.matchSymbol(")") {
		if (p.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN")) {
			p.parseTableConstraint(&constraints);
		} else {
			column := p.parseColumnDefinition()
			columns = append(columns, column)
		}
		
		if p.matchKeyword(")") {
			break
		}
		p.next()
	}
	p.next()
	return columns, constraints
}


func (p *parser) parseColumnDefinition() types.Column {
	var column types.Column
	column.Name = p.parseName()
	column.DataType = p.parseDateType()
	column.Constraint = p.parseConstraint()
	
	return column
}


func (p *parser) parseDateType() types.DataType {
	var dataType types.DataType
	dataType.Name = strings.ToUpper(p.parseToken())
	if p.matchKeyword("VARYING") {
		if dataType.Name == "BIT" {
			dataType.Name = "VARBIT"
		} else if dataType.Name == "CHARACTER" {
			dataType.Name = "VARCHAR"
		} else {
			dataType.Name += " " + strings.ToUpper(p.token())
		}
		p.next()
	}
	n, m := p.parseTypeDigit()
	dataType.DigitN = n
	dataType.DigitM = m
	return dataType
}


func (p *parser) parseTypeDigit() (int, int) {
	n := 0
	m := 0
	if p.matchSymbol("(") {
		p.next()
		n, _ = strconv.Atoi(p.parseToken())
		if p.matchSymbol(",") {
			p.next()
			m, _ = strconv.Atoi(p.parseToken())
		}
		p.next()   //skip ")"
	}
	return n, m
}


func (p *parser) parseConstraint() types.Constraint {
	var constraint types.Constraint
	if p.matchKeyword("CONSTRAINT") {
		p.next() // skip "CONSTRAINT"
		if (!p.matchKeyword("PRIMARY", "UNIQUE", "NOT", "AUTOINCREMENT", "AUTO_INCREMENT", "DEFAULT", "CHECK", "REFERENCES", "COLLATE")) {
			constraint.Name = p.parseName()
		}
	}
	p.parseConstraintAux(&constraint)
	return constraint
}


func (p *parser) parseConstraintAux(constraint *types.Constraint) {
	if p.matchSymbol(",", ")") {
		return
	}
	if p.matchKeyword("PRIMARY") {
		p.next() // skip "PRIMARY"
		p.next() // skip "KEY"
		constraint.IsPrimaryKey = true
		p.parseConstraintAux(constraint)
		return 
	}
	if p.matchKeyword("AUTOINCREMENT", "AUTO_INCREMENT") {
		p.next() // skip "AUTOINCREMENT"
		constraint.IsAutoincrement = true
		p.parseConstraintAux(constraint)
		return
	}
	if p.matchKeyword("NOT") {
		p.next() // skip "NOT"
		p.next() // skip "NULL"
		constraint.IsNotNull = true
		p.parseConstraintAux(constraint)
		return
	}
	if p.matchKeyword("UNIQUE") {
		p.next() // skip "UNIQUE"
		constraint.IsUnique = true
		p.parseConstraintAux(constraint)
		return
	}
	if p.matchKeyword("DEFAULT") {
		p.next() // skip "DEFAULT"
		constraint.Default = p.parseDefaultValue()
		p.parseConstraintAux(constraint)
		return
	}
	if p.matchKeyword("CHECK") {
		p.next() // skip "CHECK"
		constraint.Check = p.parseExpr()
		p.parseConstraintAux(constraint)
		return
	}
	if p.matchKeyword("COLLATE") {
		p.next() // skip "COLLATE"
		constraint.Collate = p.parseName()
		p.parseConstraintAux(constraint)
		return
	}
	if p.matchKeyword("REFERENCES") {
		constraint.References = p.parseReference()
		p.parseConstraintAux(constraint)
		return
	}
}


func (p *parser) parseDefaultValue() interface{} {
	if p.matchSymbol("(") {
		return p.parseExpr()
	} else {
		return p.parseLiteralValue()
	}
}


func (p *parser) parseExpr() string {
	return p.parseToken() + p.parseExprAux() + p.parseToken()
}


func (p *parser) parseExprAux() string {
	if p.matchSymbol(")") {
		return ""
	}
	if p.matchSymbol("(") {
		return p.parseExpr() + p.parseExprAux()
	}
	return p.parseToken() + p.parseExprAux()
}


func (p *parser) parseLiteralValue() interface{} {
	token := p.token()
	p.next()
	if common.IsNumericToken(token) {
		n, _ := strconv.ParseFloat(token, 64)
		return n
	}
	if p.isStringValue(token) {
		return token[1 : len(token)-1]
	}
	if strings.ToUpper(token) == "NULL" {
		return nil
	}
	if strings.ToUpper(token) == "TRUE" {
		return true
	}
	if strings.ToUpper(token) == "FALSE" {
		return false
	}
	return token	
}


func (p *parser) parseReference() types.Reference {
	var reference types.Reference
	p.next() // skip "REFERENCES"
	reference.TableName = p.parseName()
	if p.matchSymbol("(") {
		reference.ColumnNames = p.parseCommaSeparatedColumnNames()
	}
	return reference
}


func (p *parser) parseTableConstraint(tableConstraint *types.TableConstraint) {
	name := ""
	if p.matchKeyword("CONSTRAINT") {
		p.next() // skip "CONSTRAINT"
		if !p.matchKeyword("PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
			name = p.parseName()
		}
	}

	if p.matchKeyword("PRIMARY") {
		var primaryKey types.PrimaryKey
		p.next() // skip "PRIMARY"
		p.next() // skip "KEY"
		primaryKey.Name = name
		primaryKey.ColumnNames = p.parseCommaSeparatedColumnNames()
		tableConstraint.PrimaryKey = append(tableConstraint.PrimaryKey, primaryKey)

	} else if p.matchKeyword("UNIQUE") {
		var unique types.Unique
		p.next() // skip "UNIQUE"
		unique.Name = name
		unique.ColumnNames = p.parseCommaSeparatedColumnNames()
		tableConstraint.Unique = append(tableConstraint.Unique, unique)

	} else if p.matchKeyword("CHECK") {
		var check types.Check
		p.next() // skip "CHECK"
		check.Name = name
		check.Expr = p.parseExpr()
		tableConstraint.Check = append(tableConstraint.Check, check)

	} else if p.matchKeyword("FOREIGN") {
		var foreignKey types.ForeignKey
		p.next() // skip "FOREIGN"
		p.next() // skip "KEY"
		foreignKey.Name = name
		foreignKey.ColumnNames = p.parseCommaSeparatedColumnNames()
		foreignKey.References = p.parseReference()
		tableConstraint.ForeignKey = append(tableConstraint.ForeignKey, foreignKey)
	}
	return
}


func (p *parser) parseCommaSeparatedColumnNames() []string {
	p.next() // skip "(""
	ls := []string{}
	for {
		if p.matchSymbol(")") {
			break
		} else if p.matchSymbol(",") {
			p.next()
			continue
		}
		ls = append(ls, p.token())
		p.next()
	}
	p.next()
	return ls
}