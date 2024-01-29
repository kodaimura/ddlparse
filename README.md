# ddlparse
SQLのCREATE TABLE文を下記のTableオブジェクトの形に変換する。  
SQLite、MySQL、PostgreSQLに対応。対応の構文は下記[Learn more](#Learn-more)参照

## Tableオブジェクト
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
パース前に下記ルールに沿って構文チェックを行う。構文チェックに失敗した場合はValidateErrorを返し、成功した場合にのみパースを行い、Tableオブジェクトに変換する。
構文エラー以外の不正（カラム名の重複、テーブル制約で存在しないカラムを指定、など）は検出せず、構文が合っていればパースを行う。

### SQLite
```
CREATE TABLE [IF NOT EXISTS] [schema_name.]table_name (
    column_name type_name [column-constraint ...],
    [table-constraint, ...]
)[table-options][;]
```
* column-constraint
```
[CONSTRAINT name] RIMARY KEY [DESC | ASC] [conflict-clause] [AUTOINCREMENT]
[CONSTRAINT name] UNIQUE [conflict-clause]
[CONSTRAINT name] NOT NULL [conflict-clause]
[CONSTRAINT name] CHECK (expr)
[CONSTRAINT name] DEFAULT {literal-value | (expr)}
[CONSTRAINT name] COLLATE {BINARY | NOCASE | RTRIM}
[CONSTRAINT name] GENERATED ALWAYS AS (expr) [STORED | VIRTUAL]
[CONSTRAINT name] AS (expr) [STORED | VIRTUAL]
[CONSTRAINT name] REFERENCES table_name [(column_name)]
                  [ON {DELETE | UPDATE} {SET NULL | SET DEFAULT | CASCADE| RESTRICT | NO ACTION}]
                  [MATCH name]
                  [[NOT] DEFERRABLE [INITIALLY DEFERRED | INITIALLY IMMEDIATE]]
```
* table-constraint
```
[CONSTRAINT name] RIMARY KEY (column_name, ...) [conflict-clause]
[CONSTRAINT name] UNIQUE (column_name, ...) [conflict-clause]
[CONSTRAINT name] CHECK (expr)
[CONSTRAINT name] FOREIGN KEY (column_name, ...) REFERENCES table_name [(column_name, ...)]
                  [ON {DELETE | UPDATE} {SET NULL | SET DEFAULT | CASCADE|RESTRICT | NO ACTION}]
                  [MATCH name]
                  [[NOT] DEFERRABLE [INITIALLY DEFERRED | INITIALLY IMMEDIATE]]
```
* conflict-clause
```
ON CONFLICT {ROLLBACK | ABORT | FAIL | IGNORE|REPLACE}
```
* table-options
```
[WITHOUT ROWID][STRICT]
```
### PostgreSQL
```
CREATE TABLE [IF NOT EXISTS] [schema_name.]table_name (
    column_name type_name [column-constraint ...],
    [table-constraint, ...]
)[table-options];
```
* column-constraint
```
[CONSTRAINT name] RIMARY KEY [index-parameters]
[CONSTRAINT name] UNIQUE [index-parameters]
[CONSTRAINT name] NOT NULL
[CONSTRAINT name] NULL
[CONSTRAINT name] CHECK (expr) [NO INHERIT]
[CONSTRAINT name] DEFAULT {literal-value | (expr)}
[CONSTRAINT name] GENERATED ALWAYS AS (expr) STORED
[CONSTRAINT name] GENERATED {ALWAYS | BY DEFAULT} AS IDENTITY [(...)] 
[CONSTRAINT name] AS (expr) [STORED | VIRTUAL]
[CONSTRAINT name] REFERENCES table_name [(column_name)]
                  [MATCH {FULL | PARTIAL | SIMPLE}]
                  [ON {DELETE | UPDATE} {SET NULL | SET DEFAULT | CASCADE | RESTRICT | NO ACTION}]
```
* table-constraint
```
[CONSTRAINT name] RIMARY KEY (column_name, ...) [index-parameters]
[CONSTRAINT name] UNIQUE (column_name, ...) [index-parameters]
[CONSTRAINT name] CHECK (expr) [NO INHERIT]
[CONSTRAINT name] EXCLUDE [USING index-method] (...) [index-parameters] [WHERE (...)]
[CONSTRAINT name] FOREIGN KEY (column_name, ...) REFERENCES table_name [(column_name, ...)]
                  [MATCH {FULL | PARTIAL | SIMPLE}]
                  [ON {DELETE | UPDATE} {SET NULL | SET DEFAULT | CASCADE | RESTRICT | NO ACTION}]
```
* index-parameters
```
[INCLUDE (column_name , ... )]
[WITH (...) ]
[USING INDEX TABLESPACE tablespace_name]
```
* table-options
```
WITH (...)
WITHOUT OIDS
TABLESPACE tablespace_name
INHERITS (...)
PARTITION BY {RANGE | LIST | HASH} (...)
USING method
```

### MySQL
```
CREATE TABLE [IF NOT EXISTS] [schema_name.]table_name (
    column_name type_name [column-constraint ...],
    [table-constraint, ...]
)[table-options];
```
* column-constraint
```
[RIMARY] KEY
UNIQUE [KEY]
AUTO_INCREMENT
NOT NULL | NULL
DEFAULT {literal-value | (expr)}
VISIBLE | INVISIBL
COMMENT 'string'
COLLATE collation_name
COLUMN_FORMAT {FIXED | DYNAMIC | DEFAULT}
ENGINE_ATTRIBUTE [=] 'string'
SECONDARY_ENGINE_ATTRIBUTE [=] 'string'
STORAGE {DISK | MEMORY}
[GENERATED ALWAYS] AS (expr)
VIRTUAL | STORED
[CONSTRAINT [symbol]] CHECK (expr) [[NOT] ENFORCED]
REFERENCES table_name (column_name)
    [MATCH {FULL | PARTIAL | SIMPLE}]
    [ON {DELETE | UPDATE} {SET NULL | SET DEFAULT | CASCADE | RESTRICT| NO ACTION}]
```
* table-constraint
```
{INDEX | KEY} [index_name] [USING {BTREE | HASH}] (key-part, ...) [index-option] ...
{FULLTEXT | SPATIAL} [INDEX | KEY] [index_name] (key-part, ...) [index-option] ...
[CONSTRAINT [symbol]] PRIMARY KEY [USING {BTREE | HASH}] (key-part, ...) [index-option] ...
[CONSTRAINT [symbol]] UNIQUE [INDEX | KEY] [index_name] [USING {BTREE | HASH}] (key-part, ...)[index-option] ...
[CONSTRAINT [symbol]] CHECK (expr) [NO INHERIT]
[CONSTRAINT [symbol]] FOREIGN KEY [index_name] (column_name,...) REFERENCES table_name (key-part, ...)
                      [MATCH {FULL | PARTIAL | SIMPLE}]
                      [ON {DELETE | UPDATE} {SET NULL | SET DEFAULT | CASCADE | RESTRICT | NO ACTION}]
```
* key-part
```
{column_name [(length)] | (expr)} [ASC | DESC]
```
* table-options
```
AUTOEXTEND_SIZE [=] value
AUTO_INCREMENT [=] value
AVG_ROW_LENGTH [=] value
[DEFAULT] CHARACTER SET [=] charset_name
CHECKSUM [=] {0 | 1}
[DEFAULT] COLLATE [=] collation_name
COMMENT [=] 'string'
COMPRESSION [=] {'ZLIB' | 'LZ4' | 'NONE'}
CONNECTION [=] 'connect_string'
{DATA | INDEX} DIRECTORY [=] 'absolute path to directory'
DELAY_KEY_WRITE [=] {0 | 1}
ENCRYPTION [=] {'Y' | 'N'}
ENGINE [=] engine_name
ENGINE_ATTRIBUTE [=] 'string'
INSERT_METHOD [=] { NO | FIRST | LAST }
KEY_BLOCK_SIZE [=] value
MAX_ROWS [=] value
MIN_ROWS [=] value
PACK_KEYS [=] {0 | 1 | DEFAULT}
PASSWORD [=] 'string'
ROW_FORMAT [=] {DEFAULT | DYNAMIC | FIXED | COMPRESSED | REDUNDANT | COMPACT}
SECONDARY_ENGINE_ATTRIBUTE [=] 'string'
STATS_AUTO_RECALC [=] {DEFAULT | 0 | 1}
STATS_PERSISTENT [=] {DEFAULT | 0 | 1}
STATS_SAMPLE_PAGES [=] value
TABLESPACE tablespace_name [STORAGE {DISK | MEMORY}]
UNION [=] (tbl_name[,tbl_name]...)
```
