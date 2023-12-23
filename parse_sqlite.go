package ddlparse


type sqliteParser struct {
	tokens []string
}

func NewSQLiteParser(tokens []string) parser {
	return &sqliteParser{tokens}
}


func (p *sqliteParser) validate() []string {
	var errs []string
	return errs
}

func (p *sqliteParser) parse() ([]Table, error) {
	var tables []Table
	return tables, nil
}