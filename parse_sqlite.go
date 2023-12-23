package ddlparse

import (
	"errors"
)

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

func (p *sqliteParser) next() error {
	p.i += 1
	if (p.i > p.size) {
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
		if (p.i > p.size) {
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
		if (p.i > p.size) {
			return errors.New("expected '*/' but not found")
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
	return nil
}

func (p *sqliteParser) Parse() ([]Table, error) {
	p.init()
	var tables []Table
	return tables, nil
}