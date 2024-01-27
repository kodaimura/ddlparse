# ddlparse
SQLのCREATE TABLE文をGoオブジェクトに変換する。  
SQLite、MySQL、PostgreSQLに対応。

### Goオブジェクト
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
