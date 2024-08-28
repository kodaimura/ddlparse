package ddlparse

import (
	"github.com/kodaimura/ddlparse/internal/types"
	"github.com/kodaimura/ddlparse/internal/common"
	"github.com/kodaimura/ddlparse/internal/lexer"
	"github.com/kodaimura/ddlparse/internal/parser"
	"github.com/kodaimura/ddlparse/internal/validator"
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
	Rdbms = common.Rdbms
	ValidateError = common.ValidateError
)

const (
	PostgreSQL = common.PostgreSQL
	MySQL = common.MySQL
	SQLite = common.SQLite
)

func Parse(ddl string, rdbms common.Rdbms) ([]Table, error) {
	tokens, err := tokenize(ddl, rdbms)
	if err != nil {
		return []Table{}, err
	}
	validatedTokens, err := validate(tokens, rdbms)
	if err != nil {
		return []Table{}, err
	}
	tables:= parse(validatedTokens, rdbms)
	return tables, nil
}

func ParseSQLite(ddl string) ([]Table, error) {
	return Parse(ddl, common.SQLite)
}

func ParsePostgreSQL(ddl string) ([]Table, error) {
	return Parse(ddl, common.PostgreSQL)
}

func ParseMySQL(ddl string) ([]Table, error) {
	return Parse(ddl, common.MySQL)
}

func ParseForce(ddl string) ([]Table, error) {
	ls := []common.Rdbms{common.SQLite, common.PostgreSQL, common.MySQL}
	var err error
	for _, rdbms := range ls {
		tables, err := Parse(ddl, rdbms)
		if err == nil {
			return tables, nil
		}
	}
	return []Table{}, err
}

func tokenize (ddl string, rdbms common.Rdbms) ([]string, error) {
	l := lexer.NewLexer(rdbms, ddl)
	return l.Lex()
}

func validate (tokens []string, rdbms common.Rdbms) ([]string, error) {
	v := validator.NewValidator(rdbms, tokens)
	return v.Validate()
}

func parse (tokens []string, rdbms common.Rdbms) []Table {
	p := parser.NewParser(rdbms, tokens)
	return p.Parse()
}
