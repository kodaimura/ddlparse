package ddlparse

import (
	"fmt"

	"github.com/kodaimura/ddlparse"
	"github.com/kodaimura/ddlparse/types"
	"github.com/kodaimura/ddlparse/common"
)


func Parse(ddl string, rdbms common.Rdbms) ([]types.Table, error) {
	tokens, err := tokenize(ddl, rdbms)
	if err != nil {
		return []types.Table{}, err
	}
	validatedTokens, err := validate(tokens, rdbms)
	if err != nil {
		return []types.Table{}, err
	}
	tables:= parse(validatedTokens, rdbms)
	return tables, nil
}

func ParseSQLite(ddl string) ([]types.Table, error) {
	return Parse(ddl, common.SQLite)
}

func ParsePostgreSQL(ddl string) ([]types.Table, error) {
	return Parse(ddl, common.PostgreSQL)
}

func ParseMySQL(ddl string) ([]types.Table, error) {
	return Parse(ddl, common.MySQL)
}

func ParseForce(ddl string) ([]types.Table, error) {
	ls := []Rdbms{common.SQLite, common.PostgreSQL, common.MySQL}
	var err error
	for _, rdbms := range ls {
		tables, err := Parse(ddl, rdbms)
		if err == nil {
			return tables, nil
		}
	}
	return []Table{}, err
}

func tokenize (ddl string, rdbms Rdbms) ([]string, error) {
	l := NewLexer(rdbms, ddl)
	return l.Lex()
}

func validate (tokens []string, rdbms Rdbms) ([]string, error) {
	v := validator.NewValidator(rdbms, tokens)
	return v.Validate()
}

func parse (tokens []string, rdbms Rdbms) []types.Table {
	p := parser.NewParser(rdbms, tokens)
	return p.Parse()
}
