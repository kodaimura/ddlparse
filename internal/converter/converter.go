package converter

import (
	"errors"
	"strings"
	"strconv"

	"github.com/kodaimura/ddlparse/internal/types"
	"github.com/kodaimura/ddlparse/internal/common"
)


type Converter interface {
	Convert(tokens []string) []types.Table
}

/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  Convert(): 
    Convert the validated token to a Table object.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

type converter struct {
	rdbms common.Rdbms
	tokens []string
	size int
	i int
	result []types.Table
}


func NewConverter(rdbms common.Rdbms) Converter {
	return &converter{rdbms: rdbms}
}


func (c *converter) Convert(tokens []string) []types.Table {
	c.init(tokens)
	c.convert()
	return c.result
}


func (c *converter) init(tokens []string) {
	c.tokens = tokens
	c.size = len(c.tokens)
	c.i = 0
	c.result = []types.Table{}
}


func (c *converter) token() string {
	return c.tokens[c.i]
}


func (c *converter) isOutOfRange() bool {
	return c.i > c.size - 1
}


func (c *converter) next() error {
	pre := c.token()
	c.i += 1
	if (c.isOutOfRange()) {
		return errors.New("out of range")
	}
	if pre == "," && c.token() == "," {
		return c.next()
	}
	return nil
}


func (c *converter) matchKeyword(keywords ...string) bool {
	return common.Contains(
		append(
			common.MapSlice(keywords, strings.ToLower), 
			common.MapSlice(keywords, strings.ToUpper)...,
		), c.token())
}


func (c *converter) matchSymbol(symbols ...string) bool {
	return common.Contains(symbols, c.token())
}


func (c *converter) isIdentifier(token string) bool {
	tc := token[0:1]
	switch (c.rdbms) {
		case common.SQLite:
			return tc == "\"" || tc == "`"
		case common.MySQL:
			return tc == "`"
		case common.PostgreSQL:
			return tc == "\""
	}
	return false
}


func (c *converter) isStringValue(token string) bool {
	tc := token[0:1]
	switch (c.rdbms) {
		case common.SQLite:
			return tc == "'"
		case common.MySQL:
			return tc == "\"" || tc == "'"
		case common.PostgreSQL:
			return tc == "'"
	}
	return false
}


func (c *converter) convert() {
	if c.isOutOfRange() {
		return
	} else {
		table := c.convertTable()
		c.result = append(c.result, table)
	}
	c.convert()
}


func (c *converter) convertToken() string {
	token := c.token()
	c.next()
	return token
}


func (c *converter) convertTable() types.Table {
	var table types.Table
	c.next() // skip "CREATE"
	c.next() // skip "TABLE"

	if c.matchKeyword("IF") {
		c.next() // skip "IF"
		c.next() // skip "NOT"
		c.next() // skip "EXISTS"
		table.IfNotExists = true
	}

	schemaName, tableName := c.convertTableName()
	columns, constraints := c.convertTableDefinition();

	table.Schema = schemaName
	table.Name = tableName
	table.Columns = columns
	table.Constraints = constraints

	if (c.size > c.i) {
		if c.matchSymbol(";") {
			c.next()
		}
	}
	return table
}


func (c *converter) convertTableName() (string, string) {
	schemaName := c.convertName()
	tableName := ""

	if c.matchSymbol(".") {
		c.next()
		tableName = c.convertName()
	} else {
		tableName = schemaName
		schemaName = ""
	}

	return schemaName, tableName 
}


func (c *converter) convertName() string {
	token := c.convertToken()
	if c.isIdentifier(token) {
		return token[1 : len(token)-1]
	} else {
		return token
	}
}


func (c *converter) convertTableDefinition() ([]types.Column, types.TableConstraint) {
	c.next()
	var columns []types.Column
	var constraints types.TableConstraint
	for !c.matchSymbol(")") {
		if (c.matchKeyword("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN")) {
			c.convertTableConstraint(&constraints);
		} else {
			column := c.convertColumnDefinition()
			columns = append(columns, column)
		}
		
		if c.matchKeyword(")") {
			break
		}
		c.next()
	}
	c.next()
	return columns, constraints
}


func (c *converter) convertColumnDefinition() types.Column {
	var column types.Column
	column.Name = c.convertName()
	column.DataType = c.convertDateType()
	column.Constraint = c.convertConstraint()
	
	return column
}


func (c *converter) convertDateType() types.DataType {
	var dataType types.DataType
	dataType.Name = strings.ToUpper(c.convertToken())
	if c.matchKeyword("VARYING") {
		if dataType.Name == "BIT" {
			dataType.Name = "VARBIT"
		} else if dataType.Name == "CHARACTER" {
			dataType.Name = "VARCHAR"
		} else {
			dataType.Name += " " + strings.ToUpper(c.token())
		}
		c.next()
	}
	n, m := c.convertTypeDigit()
	dataType.DigitN = n
	dataType.DigitM = m
	return dataType
}


func (c *converter) convertTypeDigit() (int, int) {
	n := 0
	m := 0
	if c.matchSymbol("(") {
		c.next()
		n, _ = strconv.Atoi(c.convertToken())
		if c.matchSymbol(",") {
			c.next()
			m, _ = strconv.Atoi(c.convertToken())
		}
		c.next()   //skip ")"
	}
	return n, m
}


func (c *converter) convertConstraint() types.Constraint {
	var constraint types.Constraint
	if c.matchKeyword("CONSTRAINT") {
		c.next() // skip "CONSTRAINT"
		if (!c.matchKeyword("PRIMARY", "UNIQUE", "NOT", "AUTOINCREMENT", "AUTO_INCREMENT", "DEFAULT", "CHECK", "REFERENCES", "COLLATE")) {
			constraint.Name = c.convertName()
		}
	}
	c.convertConstraintAux(&constraint)
	return constraint
}


func (c *converter) convertConstraintAux(constraint *types.Constraint) {
	if c.matchSymbol(",", ")") {
		return
	}
	if c.matchKeyword("PRIMARY") {
		c.next() // skip "PRIMARY"
		c.next() // skip "KEY"
		constraint.IsPrimaryKey = true
		c.convertConstraintAux(constraint)
		return 
	}
	if c.matchKeyword("AUTOINCREMENT", "AUTO_INCREMENT") {
		c.next() // skip "AUTOINCREMENT"
		constraint.IsAutoincrement = true
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchKeyword("NOT") {
		c.next() // skip "NOT"
		c.next() // skip "NULL"
		constraint.IsNotNull = true
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchKeyword("UNIQUE") {
		c.next() // skip "UNIQUE"
		constraint.IsUnique = true
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchKeyword("DEFAULT") {
		c.next() // skip "DEFAULT"
		constraint.Default = c.convertDefaultValue()
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchKeyword("CHECK") {
		c.next() // skip "CHECK"
		constraint.Check = c.convertExpr()
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchKeyword("COLLATE") {
		c.next() // skip "COLLATE"
		constraint.Collate = c.convertName()
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchKeyword("REFERENCES") {
		constraint.References = c.convertReference()
		c.convertConstraintAux(constraint)
		return
	}
}


func (c *converter) convertDefaultValue() interface{} {
	if c.matchSymbol("(") {
		return c.convertExpr()
	} else {
		return c.convertLiteralValue()
	}
}


func (c *converter) convertExpr() string {
	return c.convertToken() + c.convertExprAux() + c.convertToken()
}


func (c *converter) convertExprAux() string {
	if c.matchSymbol(")") {
		return ""
	}
	if c.matchSymbol("(") {
		return c.convertExpr() + c.convertExprAux()
	}
	return c.convertToken() + c.convertExprAux()
}


func (c *converter) convertLiteralValue() interface{} {
	token := c.token()
	c.next()
	if common.IsNumericToken(token) {
		n, _ := strconv.ParseFloat(token, 64)
		return n
	}
	if c.isStringValue(token) {
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


func (c *converter) convertReference() types.Reference {
	var reference types.Reference
	c.next() // skip "REFERENCES"
	reference.TableName = c.convertName()
	if c.matchSymbol("(") {
		reference.ColumnNames = c.convertCommaSeparatedColumnNames()
	}
	return reference
}


func (c *converter) convertTableConstraint(tableConstraint *types.TableConstraint) {
	name := ""
	if c.matchKeyword("CONSTRAINT") {
		c.next() // skip "CONSTRAINT"
		if !c.matchKeyword("PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
			name = c.convertName()
		}
	}

	if c.matchKeyword("PRIMARY") {
		var primaryKey types.PrimaryKey
		c.next() // skip "PRIMARY"
		c.next() // skip "KEY"
		primaryKey.Name = name
		primaryKey.ColumnNames = c.convertCommaSeparatedColumnNames()
		tableConstraint.PrimaryKey = append(tableConstraint.PrimaryKey, primaryKey)

	} else if c.matchKeyword("UNIQUE") {
		var unique types.Unique
		c.next() // skip "UNIQUE"
		unique.Name = name
		unique.ColumnNames = c.convertCommaSeparatedColumnNames()
		tableConstraint.Unique = append(tableConstraint.Unique, unique)

	} else if c.matchKeyword("CHECK") {
		var check types.Check
		c.next() // skip "CHECK"
		check.Name = name
		check.Expr = c.convertExpr()
		tableConstraint.Check = append(tableConstraint.Check, check)

	} else if c.matchKeyword("FOREIGN") {
		var foreignKey types.ForeignKey
		c.next() // skip "FOREIGN"
		c.next() // skip "KEY"
		foreignKey.Name = name
		foreignKey.ColumnNames = c.convertCommaSeparatedColumnNames()
		foreignKey.References = c.convertReference()
		tableConstraint.ForeignKey = append(tableConstraint.ForeignKey, foreignKey)
	}
	return
}


func (c *converter) convertCommaSeparatedColumnNames() []string {
	c.next() // skip "(""
	ls := []string{}
	for {
		if c.matchSymbol(")") {
			break
		} else if c.matchSymbol(",") {
			c.next()
			continue
		}
		ls = append(ls, c.token())
		c.next()
	}
	c.next()
	return ls
}