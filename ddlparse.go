package sqlparse

import (
	"errors"
	"fmt"
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

func ParseDdl (ddl string, rdbms Rdbms) ([]Table, error) {
	switch rdbms {
	case SQLite:
		return ParseDdlSQLite(ddl)
	//case PostgreSQL:
	//	return ParseDdlPostgreSQL(ddl)
	//case MySQL:
	//	return ParseDdlMySQL(ddl)
	default:
		return []Table{} errors.New("Not yet supported.")
	}
}

func ParseDdlSQLite (ddl string) ([]Table, error) {

}

