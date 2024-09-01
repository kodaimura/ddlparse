package ddlparse

import (
	"github.com/kodaimura/ddlparse/internal/types"
	"github.com/kodaimura/ddlparse/internal/common"
	"github.com/kodaimura/ddlparse/internal/lexer"
	"github.com/kodaimura/ddlparse/internal/validator"
	"github.com/kodaimura/ddlparse/internal/converter"
)


type (
	Table = types.Table
	Column = types.Column
	DataType = types.DataType
	Constraint = types.Constraint
	Reference = types.Reference
	TableConstraint = types.TableConstraint
	PrimaryKey = types.PrimaryKey
	Unique = types.Unique
	Check = types.Check
	ForeignKey = types.ForeignKey
)

type (
	Rdbms = common.Rdbms
	ValidateError = common.ValidateError
)

const (
	PostgreSQL = common.PostgreSQL
	MySQL = common.MySQL
	SQLite = common.SQLite
)

func Parse(ddl string, rdbms Rdbms) ([]Table, error) {
	l := lexer.NewLexer(rdbms)
	v := validator.NewValidator(rdbms)
	c := converter.NewConverter(rdbms)

	tokens, err := l.Lex(ddl)
	if err != nil {
		return []Table{}, err
	}
	
	validatedTokens, err := v.Validate(tokens)
	if err != nil {
		return []Table{}, err
	}
	
	return c.Convert(validatedTokens), nil
}

func ParseSQLite(ddl string) ([]Table, error) {
	return Parse(ddl, SQLite)
}

func ParsePostgreSQL(ddl string) ([]Table, error) {
	return Parse(ddl, PostgreSQL)
}

func ParseMySQL(ddl string) ([]Table, error) {
	return Parse(ddl, MySQL)
}

func ParseForce(ddl string) ([]Table, error) {
	ls := []Rdbms{SQLite, PostgreSQL, MySQL}
	var err error
	for _, rdbms := range ls {
		tables, err := Parse(ddl, rdbms)
		if err == nil {
			return tables, nil
		}
	}
	return []Table{}, err
}