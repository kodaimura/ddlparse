package ddlparse


type postgresqlParser struct {
	tokens []string
	size int
	i int
	lines int
}

func newPostgreSQLParser(tokens []string) parser {
	return &postgresqlParser{tokens, len(tokens), 0, 0}
}

func (p *postgresqlParser) init() {
	p.i = 0
	p.lines = 0
}

func (p *postgresqlParser) next() {
	p.i += 1
	if (p.i <= p.size && p.tokens[p.i] == "\n") {
		p.lines += 1
		p.next()
	}
}

func (p *postgresqlParser) Validate() error {
	p.init()
	return nil
}

func (p *postgresqlParser) Parse() ([]Table, error) {
	p.init()
	var tables []Table
	return tables, nil
}