package ddlparse


type mysqlParser struct {
	tokens []string
}

func NewMySQLParser(tokens []string) parser {
	return &mysqlParser{tokens}
}


func (p *mysqlParser) validate() []string {
	var errs []string
	return errs
}

func (p *mysqlParser) parse() ([]Table, error) {
	var tables []Table
	return tables, nil
}