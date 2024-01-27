# ddlparse
SQLのCREATE TABLE文をGoオブジェクトに変換する。  
下記のRDBMSに対応。
* SQLite
* MySQL
* PostgreSQL

Goオブジェクト
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
