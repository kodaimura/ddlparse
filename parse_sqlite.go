package ddlparse


type sqliteParser struct {
	tokens []string
	size int
	i int
	lines int
}

func newSQLiteParser(tokens []string) parser {
	return &sqliteParser{tokens, len(tokens), 0, 0}
}

func (p *sqliteParser) init() {
	p.i = 0
	p.lines = 0
}

func (p *sqliteParser) next() {
	p.i += 1
	if (p.i <= p.size && p.tokens[p.i] == "\n") {
		p.lines += 1
		p.next()
	}
}

func (p *sqliteParser) Validate() error {
	p.init()
	return nil
}

func (p *sqliteParser) Parse() ([]Table, error) {
	p.init()
	var tables []Table
	return tables, nil
}