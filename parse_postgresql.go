package ddlparse


type postgresqlParser struct {
	tokens []string
}

func NewPostgreSQLParser(tokens []string) parser {
	return &postgresqlParser{tokens}
}


func (p *postgresqlParser) validate() []string {
	var errs []string
	return errs
}

func (p *postgresqlParser) parse() ([]Table, error) {
	var tables []Table
	return tables, nil
}