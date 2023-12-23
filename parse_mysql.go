package ddlparse


type mysqlParser struct {
	tokens []string
	size int
	i int
	lines int
}

func newMySQLParser(tokens []string) parser {
	return &mysqlParser{tokens, len(tokens), 0, 0}
}

func (p *mysqlParser) init() {
	p.i = 0
	p.lines = 0
}

func (p *mysqlParser) next() {
	p.i += 1
	if (p.i <= p.size && p.tokens[p.i] == "\n") {
		p.lines += 1
		p.next()
	}
}

func (p *mysqlParser) Validate() error {
	p.init()
	return nil
}

func (p *mysqlParser) Parse() ([]Table, error) {
	p.init()
	var tables []Table
	return tables, nil
}