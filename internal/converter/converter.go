package converter

import (
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
	if c.isOutOfRange() {
		return common.EOF
	}
	return c.tokens[c.i]
}


func (c *converter) isOutOfRange() bool {
	return c.i > c.size - 1
}


func (c *converter) next() string {
	if (c.isOutOfRange()) {
		return common.EOF
	}
	token := c.token()
	c.i += 1
	if token == "," && c.token() == "," {
		return c.next()
	}
	return token
}


func (c *converter) matchToken(keywords ...string) bool {
	return common.Contains(
		append(
			common.MapSlice(keywords, strings.ToLower), 
			common.MapSlice(keywords, strings.ToUpper)...,
		), c.token())
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


func (c *converter) convertTable() types.Table {
	var table types.Table
	c.next() // skip "CREATE"
	c.next() // skip "TABLE"

	if c.matchToken("IF") {
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
		if c.matchToken(";") {
			c.next()
		}
	}
	return table
}


func (c *converter) convertTableName() (string, string) {
	schemaName := c.convertName()
	tableName := ""

	if c.matchToken(".") {
		c.next()
		tableName = c.convertName()
	} else {
		tableName = schemaName
		schemaName = ""
	}

	return schemaName, tableName 
}


func (c *converter) convertName() string {
	token := c.next()
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
	for !c.matchToken(")") {
		if (c.matchToken("CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN")) {
			c.convertTableConstraint(&constraints);
		} else {
			column := c.convertColumnDefinition()
			columns = append(columns, column)
		}
		
		if c.matchToken(")") {
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
	dataType.Name = strings.ToUpper(c.next())
	if c.matchToken("VARYING") {
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
	if c.matchToken("(") {
		c.next()
		n, _ = strconv.Atoi(c.next())
		if c.matchToken(",") {
			c.next()
			m, _ = strconv.Atoi(c.next())
		}
		c.next()   //skip ")"
	}
	return n, m
}


func (c *converter) convertConstraint() types.Constraint {
	var constraint types.Constraint
	if c.matchToken("CONSTRAINT") {
		c.next() // skip "CONSTRAINT"
		if (!c.matchToken("PRIMARY", "UNIQUE", "NOT", "AUTOINCREMENT", "AUTO_INCREMENT", "DEFAULT", "CHECK", "REFERENCES", "COLLATE")) {
			constraint.Name = c.convertName()
		}
	}
	c.convertConstraintAux(&constraint)
	return constraint
}


func (c *converter) convertConstraintAux(constraint *types.Constraint) {
	if c.matchToken(",", ")") {
		return
	}
	if c.matchToken("PRIMARY") {
		c.next() // skip "PRIMARY"
		c.next() // skip "KEY"
		constraint.IsPrimaryKey = true
		c.convertConstraintAux(constraint)
		return 
	}
	if c.matchToken("AUTOINCREMENT", "AUTO_INCREMENT") {
		c.next() // skip "AUTOINCREMENT"
		constraint.IsAutoincrement = true
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchToken("NOT") {
		c.next() // skip "NOT"
		c.next() // skip "NULL"
		constraint.IsNotNull = true
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchToken("UNIQUE") {
		c.next() // skip "UNIQUE"
		constraint.IsUnique = true
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchToken("DEFAULT") {
		c.next() // skip "DEFAULT"
		constraint.Default = c.convertDefaultValue()
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchToken("CHECK") {
		c.next() // skip "CHECK"
		constraint.Check = c.convertExpr()
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchToken("COLLATE") {
		c.next() // skip "COLLATE"
		constraint.Collate = c.convertName()
		c.convertConstraintAux(constraint)
		return
	}
	if c.matchToken("REFERENCES") {
		constraint.References = c.convertReference()
		c.convertConstraintAux(constraint)
		return
	}
}


func (c *converter) convertDefaultValue() interface{} {
	if c.matchToken("(") {
		return c.convertExpr()
	} else {
		return c.convertLiteralValue()
	}
}


func (c *converter) convertExpr() string {
	return c.next() + c.convertExprAux() + c.next()
}


func (c *converter) convertExprAux() string {
	if c.matchToken(")") {
		return ""
	}
	if c.matchToken("(") {
		return c.convertExpr() + c.convertExprAux()
	}
	return c.next() + c.convertExprAux()
}


func (c *converter) convertLiteralValue() interface{} {
	token := c.next()
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
	if c.matchToken("(") {
		reference.ColumnNames = c.convertCommaSeparatedColumnNames()
	}
	return reference
}


func (c *converter) convertTableConstraint(tableConstraint *types.TableConstraint) {
	name := ""
	if c.matchToken("CONSTRAINT") {
		c.next() // skip "CONSTRAINT"
		if !c.matchToken("PRIMARY", "UNIQUE", "CHECK", "FOREIGN") {
			name = c.convertName()
		}
	}

	if c.matchToken("PRIMARY") {
		var primaryKey types.PrimaryKey
		c.next() // skip "PRIMARY"
		c.next() // skip "KEY"
		primaryKey.Name = name
		primaryKey.ColumnNames = c.convertCommaSeparatedColumnNames()
		tableConstraint.PrimaryKey = append(tableConstraint.PrimaryKey, primaryKey)

	} else if c.matchToken("UNIQUE") {
		var unique types.Unique
		c.next() // skip "UNIQUE"
		unique.Name = name
		unique.ColumnNames = c.convertCommaSeparatedColumnNames()
		tableConstraint.Unique = append(tableConstraint.Unique, unique)

	} else if c.matchToken("CHECK") {
		var check types.Check
		c.next() // skip "CHECK"
		check.Name = name
		check.Expr = c.convertExpr()
		tableConstraint.Check = append(tableConstraint.Check, check)

	} else if c.matchToken("FOREIGN") {
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
		if c.matchToken(")") {
			break
		} else if c.matchToken(",") {
			c.next()
			continue
		}
		ls = append(ls, c.token())
		c.next()
	}
	c.next()
	return ls
}