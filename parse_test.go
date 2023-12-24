package ddlparse

import (
	"testing"
)

func TestParse(t *testing.T) {
	//Parse()
    //t.Errorf("failed")
}

func TestTokenize(t *testing.T) {
	tokens := tokenize(
		`CREATE TABLE IF NOT EXISTS users (
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
		);`,
	)

	if len(tokens) != 76 {
		t.Errorf("failed")
	}
}

func TestSkip(t *testing.T) {
	tokens := tokenize(`,--XXXXX
			a
			/*
			password TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			*/
			--XXXXX
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
		);`,
	)

	parser := &postgresqlParser{tokens, len(tokens), 100, 100}
	parser.init()
	parser.next()
	if parser.tokens[parser.i] != "a" {
		t.Errorf("failed")
	}
	parser.next()
	if parser.tokens[parser.i] != "created_at" {
		t.Errorf("failed")
	}
	parser.next()
	if parser.tokens[parser.i] != "TEXT" {
		t.Errorf("failed")
	}
}