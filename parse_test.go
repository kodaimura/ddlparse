package ddlparse

import (
	"fmt"
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
			password TEXT NOT NULL --XXX,
			email TEXT NOT NULL UNIQUE, /*aaa*/
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
		);`,
	)

	if len(tokens) != 81 {
		t.Errorf("failed")
	}
}

func TestInit(t *testing.T) {
	tokens := tokenize(`--XXXXX
			a`,
	)

	parser := &sqliteParser{tokens, len(tokens), 100, 100}
	parser.init()
	if parser.tokens[parser.i] != "a" {
		t.Errorf("failed")
	}
	if parser.size != len(tokens) {
		t.Errorf("failed")
	}
	if parser.line != 2 {
		t.Errorf("failed")
	}
}

func TestNext(t *testing.T) {
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

	parser := &sqliteParser{tokens, len(tokens), 100, 100}
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
	parser.next()
	if parser.line !=  8 {
		t.Errorf("failed")
	}
}

func TestValidate(t *testing.T) {
	tokens := tokenize(`CREATE TABLE IF NOT EXISTS users (

	);`,
	)

	parser := newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF EXISTS users ();`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREAT TABLE IF NOT EXISTS users ();`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
	tokens = tokenize(`CREATE TABL IF NOT EXISTS users ();`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
	tokens = tokenize(`CREATE TABLE IF NOT EXISTS "users ();`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
	tokens = tokenize(`CREATE TABLE IF NOT EXISTS AUTOINCREMENT ();`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}
}