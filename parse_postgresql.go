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

func (p *postgresqlParser) isOutOfRange() {
	return p.i > p.size - 1
}

func (p *postgresqlParser) syntaxError() {
	if p.isOutOfRange() {
		return NewValidateError(p.line, p.tokens[p.size - 1])
	}
	return NewValidateError(p.line, p.tokens[0])
}

func (p *postgresqlParser) init() {
	p.i = -1
	p.lines = 0
	p.next()
}

func (p *postgresqlParser) next() error {
	p.i += 1
	if (p.isOutOfRange()) {
		return nil;
	}
	if (p.tokens[p.i] == "\n") {
		p.lines += 1
		return p.next()
	} else if (p.tokens[p.i] == "--") {
		p.skipSingleLineComment()
		return p.next()
	} else if (p.tokens[p.i] == "/*") {
		if err := p.skipMultiLineComment(); err != nil {
			return err
		}
		return p.next()
	} else {
		return nil
	}
}

func (p *postgresqlParser) skipSingleLineComment() {
	if (p.tokens[p.i] != "--") {
		return
	}
	var skip func()
	skip = func() {
		p.i += 1
		if (p.isOutOfRange()) {
			return
		} else if (p.tokens[p.i] == "\n") {
			p.lines += 1
		} else {
			skip()
		}
	}
	skip()
}

func (p *postgresqlParser) skipMultiLineComment() error {
	if (p.tokens[p.i] != "/*") {
		return nil
	}
	var skip func() error
	skip = func() error {
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		} else if (p.tokens[p.i] == "\n") {
			p.lines += 1
			return skip()
		} else if (p.tokens[p.i] == "*/") {
			return nil
		} else {
			return skip()
		}
	}
	return skip()
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