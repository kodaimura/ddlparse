# ddlparse
SQLのCREATE TABLE文を下記のTableオブジェクトの形に変換する。  
SQLite、MySQL、PostgreSQLに対応。対応の構文は下記[Learn more](#Learn-more)参照

### Table
```go
type Table struct {
	Schema string `json:"schema"`
	Name string `json:"name"`
	IfNotExists bool `json:"if_not_exists"`
	Columns []Column `json:"columns"`
	Constraints TableConstraint `json:"constraints"`
}

type Column struct {
	Name string `json:"name"`
	DataType DataType `json:"data_type"`
	Constraint Constraint `json:"constraint"`
}

type DataType struct {
	Name string `json:"name"`
	DigitN int `json:"digit_n"`
	DigitM int `json:"digit_m"`
}

type Constraint struct {
	Name string `json:"name"`
	IsPrimaryKey bool `json:"is_primary_key"`
	IsUnique bool `json:"is_unique"`
	IsNotNull bool `json:"is_not_null"`
	IsAutoincrement bool `json:"is_autoincrement"`
	Default interface{} `json:"default"`
	Check string `json:"check"`
	Collate string `json:"collate"`
	References Reference `json:"references"`
}

type Reference struct {
	TableName string `json:"table_name"`
	ColumnNames []string `json:"column_names"`
}

type TableConstraint struct {
	PrimaryKey []PrimaryKey `json:"primary_key"`
	Unique []Unique `json:"unique"`
	Check []Check `json:"check"`
	ForeignKey []ForeignKey `json:"foreign_key"`
}

type PrimaryKey struct {
	Name string `json:"name"`
	ColumnNames []string `json:"column_names"`
}

type Unique struct {
	Name string `json:"name"`
	ColumnNames []string `json:"column_names"`
}

type Check struct {
	Name string `json:"name"`
	Expr string `json:"expr"`
}

type ForeignKey struct {
	Name string `json:"name"`
	ColumnNames []string `json:"column_names"`
	References Reference `json:"references"`
}
```
## Install
```
$ go get github.com/kodaimura/ddlparse
```
## Usage
利用可能な関数は https://github.com/kodaimura/ddlparse/blob/main/ddlparse.go 参照
```go
package main

import (
    "fmt"
    "github.com/kodaimura/ddlparse"
)

func main() {
    ddl := 
    `CREATE TABLE users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        email VARCHAR(50) NOT NULL UNIQUE,
        username VARCHAR(50),
        password TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );`

    tables, err := ddlparse.Parse(ddl, ddlparse.MySQL)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(tables)
    }
}
```
## Learn more
### DDL構文サポート状況
パース前に下記ルールに沿って構文チェックを行う。構文チェックに失敗した場合はValidateErrorを返し、成功した場合にのみパースが行われ、Tableオブジェクトに変換される。  
※構文のチェックしか行わないため、カラム名の重複、テーブル制約で存在しないカラムを指定した場合のようなものはエラーとして検出せず、Tableオブジェクトに変換される。

### SQLite
```
CREATE TABLE [IF NOT EXISTS] [schema-name.]table-name (
    column-name type-name [column-constraint ...],
    [table-constraint, ...]
)[table-options][;]
```
* column-constraint
```
[CONSTRAINT name] RIMARY KEY [DESC|ASC] [conflict-clause] [AUTOINCREMENT]
[CONSTRAINT name] UNIQUE [conflict-clause]
[CONSTRAINT name] NOT NULL [conflict-clause]
[CONSTRAINT name] CHECK (expr)
[CONSTRAINT name] DEFAULT {literal-value|(expr)}
[CONSTRAINT name] COLLATE {BINARY|NOCASE|RTRIM}
[CONSTRAINT name] GENERATED ALWAYS AS (expr) [STORED|VIRTUAL]
[CONSTRAINT name] AS (expr) [STORED|VIRTUAL]
[CONSTRAINT name] REFERENCES table-name [(column-name)]
                  [ON {DELETE|UPDATE} {SET NULL|SET DEFAULT|CASCADE|RESTRICT|NO ACTION}]
                  [MATCH name]
                  [[NOT] DEFERRABLE [INITIALLY DEFERRED | INITIALLY IMMEDIATE]]
```
* table-constraint
```
[CONSTRAINT name] RIMARY KEY (column-name, ...) [conflict-clause]
[CONSTRAINT name] UNIQUE (column-name, ...) [conflict-clause]
[CONSTRAINT name] CHECK (expr)
[CONSTRAINT name] FOREIGN KEY (column-name, ...) REFERENCES table-name [(column-name, ...)]
                  [ON {DELETE|UPDATE} {SET NULL|SET DEFAULT|CASCADE|RESTRICT|NO ACTION}]
                  [MATCH name]
                  [[NOT] DEFERRABLE [INITIALLY DEFERRED | INITIALLY IMMEDIATE]]
```
* conflict-clause
```
ON CONFLICT {ROLLBACK|ABORT|FAIL|IGNORE|REPLACE}
```
* table-options
```
[WITHOUT ROWID][STRICT]
```
### PostgreSQL
```
CREATE TABLE [IF NOT EXISTS] [schema-name.]table-name (
    column-name type-name [column-constraint ...],
    [table-constraint, ...]
)[table-options][;]
```
* column-constraint
```
[CONSTRAINT name] RIMARY KEY [index-parameters]
[CONSTRAINT name] UNIQUE [index-parameters]
[CONSTRAINT name] NOT NULL
[CONSTRAINT name] NULL
[CONSTRAINT name] CHECK (expr) [NO INHERIT]
[CONSTRAINT name] DEFAULT {literal-value|(expr)}
[CONSTRAINT name] GENERATED ALWAYS AS (expr) STORED
[CONSTRAINT name] GENERATED {ALWAYS|BY DEFAULT} AS IDENTITY [(sequence_options)] 
[CONSTRAINT name] AS (expr) [STORED|VIRTUAL]
[CONSTRAINT name] REFERENCES table-name [(column-name)]
                  [MATCH FULL|MATCH PARTIAL|MATCH SIMPLE]
                  [ON {DELETE|UPDATE} {SET NULL|SET DEFAULT|CASCADE|RESTRICT|NO ACTION}]
```
* table-constraint
```
[CONSTRAINT name] RIMARY KEY (column-name, ...) [index-parameters]
[CONSTRAINT name] UNIQUE (column-name, ...) [index-parameters]
[CONSTRAINT name] CHECK (expr) [NO INHERIT]
[CONSTRAINT name] EXCLUDE [USING index_method] (...) [index-parameters] [WHERE (...)]
[CONSTRAINT name] FOREIGN KEY (column-name, ...) REFERENCES table-name [(column-name, ...)]
                  [MATCH FULL|MATCH PARTIAL|MATCH SIMPLE]
                  [ON {DELETE|UPDATE} {SET NULL|SET DEFAULT|CASCADE|RESTRICT|NO ACTION}]
```
* index-parameters
```
[INCLUDE (column_name , ... )]
[WITH (...) ]
[USING INDEX TABLESPACE tablespace_name]
```
* table-options
```
[WITH](...)
[WITHOUT OIDS]
[TABLESPACE tablespace_name]
```
