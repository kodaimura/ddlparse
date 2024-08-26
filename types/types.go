package types


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