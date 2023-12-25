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

func (p *sqliteParser) isOutOfRange() {
	return p.i > p.size - 1
}

func (p *sqliteParser) syntaxError() {
	if p.isOutOfRange() {
		return NewValidateError(p.line, p.tokens[p.size - 1])
	}
	return NewValidateError(p.line, p.tokens[0])
}

func (p *sqliteParser) init() {
	p.i = -1
	p.lines = 0
	p.next()
}

func (p *sqliteParser) next() error {
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

func (p *sqliteParser) skipSingleLineComment() {
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

func (p *sqliteParser) skipMultiLineComment() error {
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

func (p *sqliteParser) Validate() error {
	p.init()
	return p.validate()
}

func (p *sqliteParser) validate() error {
	if (p.isOutOfRange()) {
		return nil
	}
	if err := p.validateCreateTable(); err != nil {
		return err
	}
	return p.validate()
}

func (p *sqliteParser) validateCreateTable() error {
	if (p.tokens[p.i] != "create" && p.tokens[p.i] != "CREATE") {
		return p.syntaxError()
	}
	p.next()
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if (p.tokens[p.i] != "table" && p.tokens[p.i] != "TABLE") {
		return p.syntaxError()
	}
	p.next()
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if (p.tokens[p.i] == "if" || p.tokens[p.i] == "IF") {
		p.next()
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
		if (p.tokens[p.i] != "not" && p.tokens[p.i] != "NOT") {
			return p.syntaxError()
		}
		p.next()
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
		if (p.tokens[p.i] != "exists" && p.tokens[p.i] != "EXISTS") {
			return p.syntaxError()
		}
	}
	p.next()
	if (p.isOutOfRange()) {
		return p.syntaxError()
	}
	if (p.tokens[p.i] = "") {
		p.validateTableName
	}
}

func (p *sqliteParser) validateTableName() error {
	if (p.tokens[p.i] == "\"") {
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
		if !isValidTableName(p.tokens[p.i]) {
			return p.syntaxError()
		}
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		}
	}

}

func (p *sqliteParser) Parse() ([]Table, error) {
	p.init()
	var tables []Table
	return tables, nil
}