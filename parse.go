package ddlparse

import (
	"fmt"
	"errors"
	"strings"
)

type Rdbms string

const (
	SQLite Rdbms = "SQLite"
	PostgreSQL Rdbms = "PostgreSQL"
	MySQL Rdbms = "MySQL"
)

type Table struct {
	Name string
	Columns []Column
}

type Column struct {
	Name string
	DataType string
	IsPK bool
	IsNotNull bool
	IsUnique bool
	IsAutoIncrement bool
	Default interface{}
	Check func(interface{}) bool
}

func Parse (ddl string, rdbms Rdbms) ([]Table, error) {
	switch rdbms {
	case SQLite:
		return ParseSQLite(ddl)
	//case PostgreSQL:
	//	return ParsePostgreSQL(ddl)
	//case MySQL:
	//	return ParseMySQL(ddl)
	default:
		return []Table{}, errors.New("Not yet supported.")
	}
}

func ParseSQLite (ddl string) ([]Table, error) {
	tokens := tokenize(ddl)

	for _, t := range tokens {
		fmt.Sprintf(t)
	}

	return []Table{}, nil
}

func tokenize (ddl string) []string {
	ddl = strings.Replace(ddl, "(", " ( ", -1)
	ddl = strings.Replace(ddl, ")", " ) ", -1)
	ddl = strings.Replace(ddl, ";", " ; ", -1)
	ddl = strings.Replace(ddl, "\"", " \" ", -1)
	ddl = strings.Replace(ddl, "'", " ' ", -1)
	ddl = strings.Replace(ddl, "`", " ` ", -1)
	ddl = strings.Replace(ddl, ",", " , ", -1)
	ddl = strings.Replace(ddl, "\t", " ", -1)
	ddl = strings.Replace(ddl, "\n", " ", -1)

	return filter(
		strings.Split(ddl, " "), 
		func(s string) bool {return s != " " && s != ""},
	)
}

func filter (array []string, f func(string) bool) []string {
	var ret []string
	for _, s := range array {
		if f(s) {
			ret = append(ret, s)
		}
	}
	return ret
}