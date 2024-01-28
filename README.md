# ddlparse
SQLのCREATE TABLE文を下記のTableオブジェクトの形に変換する。  
SQLite、MySQL、PostgreSQLに対応。

### Table
```go
type Table struct {
    Schema string        //スキーマ名
    Name string          //テーブル名
    Columns []Column     //カラム定義
}

type Column struct {
    Name string          //カラム名
    DataType string      //データ型
    DigitN int           //データ桁 例) NUMERIC(N, _) / CHAR(N)
    DigitM int           //データ桁 例) NUMERIC(_, M)
    IsPK bool            //PrimaryKey制約
    IsNotNull bool       //NotNull制約
    IsUnique bool        //Unique制約
    IsAutoIncrement bool //Autoincrement制約
    Default interface{}  //デフォルト値 例) 10 / "aaa" / true / false / nil / "(DATETIME('now', 'localtime'))"
}
```

## Install
```
$ go get github.com/kodaimura/ddlparse
```

## Usage
Please refer to https://github.com/kodaimura/ddlparse/blob/main/ddlparse.go for the provided functions and more.

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
パース前に下記ルールに沿って構文チェックを行います。構文チェックに失敗した場合はValidateErrorを返し、成功した場合にのみパースが行われ、Tableオブジェクトに変換されます。  
構文チェックを通過した場合でもパース時チェックによりParseErrorを返すことがあります。  
※RDBMSの実際のエラーチェックとは完全に一致していないためパースが成功した場合でも、RDBMSではエラーとなることがあります。

#### パース時チェック
```sql
# カラム名が重複
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    id INT
);

# PKを重複して指定
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    PRIMARY KEY(id)
);

# UNIQUEを重複して指定
CREATE TABLE users (
    email VARCHAR(50) NOT NULL UNIQUE,
    UNIQUE(email)
);

# 存在しないカラムを指定
CREATE TABLE users (
    email VARCHAR(50) NOT NULL UNIQUE,
    UNIQUE(email2)
);
```
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
foreign-key-clause
```
* table-constraint
```
[CONSTRAINT name] RIMARY KEY (column-name, ...) [conflict-clause]
[CONSTRAINT name] UNIQUE (column-name, ...) [conflict-clause]
[CONSTRAINT name] CHECK (expr)
[CONSTRAINT name] FOREIGN KEY (column-name, ...) foreign-key-clause
```
* conflict-clause
```
ON CONFLICT {ROLLBACK|ABORT|FAIL|IGNORE|REPLACE}
```

* foreign-key-clause
```
REFERENCES table-name [(column-name, ...)]
  [
    ON {DELETE|UPDATE} {SET NULL|SET DEFAULT|CASCADE|RESTRICT|NO ACTION} |
    MATCH name |
    [NOT] DEFERRABLE [INITIALLY DEFERRED | INITIALLY IMMEDIATE]
  ]
```
* table-options
```
WITHOUT ROWID|STRICT
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
[CONSTRAINT name] 
{ RIMARY KEY [index_parameters]
  UNIQUE [index_parameters]
  NOT NULL
  NULL
  CHECK (expr) [NO INHERIT]
  DEFAULT {literal-value|(expr)}
  GENERATED ALWAYS AS (expr) STORED
  GENERATED {ALWAYS|BY DEFAULT} AS IDENTITY [(sequence_options)] 
  AS (expr) [STORED|VIRTUAL]
  foreign-key-clause }
[ DEFERRABLE | NOT DEFERRABLE ] [ INITIALLY DEFERRED | INITIALLY IMMEDIATE ]
```

* table-constraint
```
[CONSTRAINT name] RIMARY KEY (column-name, ...) [conflict-clause]
[CONSTRAINT name] UNIQUE (column-name, ...) [conflict-clause]
[CONSTRAINT name] CHECK (expr)
[CONSTRAINT name] FOREIGN KEY (column-name, ...) foreign-key-clause
```
* conflict-clause
```
ON CONFLICT {ROLLBACK|ABORT|FAIL|IGNORE|REPLACE}
```

* foreign-key-clause
```
REFERENCES table-name [(column-name, ...)]
  [
    ON {DELETE|UPDATE} {SET NULL|SET DEFAULT|CASCADE|RESTRICT|NO ACTION} |
    MATCH name |
    [NOT] DEFERRABLE [INITIALLY DEFERRED | INITIALLY IMMEDIATE]
  ]
```
* table-options
```
WITHOUT ROWID|STRICT
```
