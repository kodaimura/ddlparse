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

	tokens = tokenize(`CREATE TABLE users (
		user_id INTEGER
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`create table if not exists users (
		user_id INTEGER
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE "users" (
		user_id INTEGER
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE aaaa.users (
		user_id INTEGER
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE "aaaa"."users" (
		user_id INTEGER
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE aaaa. (
		user_id INTEGER
	);`)
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

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
		bbbb TEXT,
		cccc NUMERIC,
		dddd INTEGER,
		eeee REAL,
		ffff NONE,
		aaaa integer,
		bbbb text,
		cccc numeric,
		dddd integer,
		eeee real,
		ffff none
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
		bbbb TEXT,
		cccc NUMERICCC
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		aaaa INTEGER PRIMARY KEY AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ASC AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY DESC AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON CONFLICT ROLLBACK AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON CONFLICT ABORT AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON CONFLICT FAIL AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON CONFLICT IGNORE AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON CONFLICT REPLACE AUTOINCREMENT,
		aaaa integer primary key on conflict rollback autoincrement,
		aaaa INTEGER PRIMARY KEY ON CONFLICT ROLLBACK
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		aaaa INTEGER PRIMARY AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ASC AUTOINCREMENT
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		aaaa INTEGER PRIMARY KEY AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON ROLLBACK
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		aaaa INTEGER PRIMARY KEY AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON CONFLICT ROLLBACKKK
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER NOT NULL,
		aaaa INTEGER NOT NULL ON CONFLICT ROLLBACK,
		aaaa INTEGER NOT NULL ON CONFLICT ABORT,
		aaaa INTEGER NOT NULL ON CONFLICT FAIL,
		aaaa INTEGER NOT NULL ON CONFLICT IGNORE,
		aaaa INTEGER NOT NULL ON CONFLICT REPLACE,
		aaaa integer not null on conflict rollback
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER NOT NULL,
		aaaa INTEGER NOT ON CONFLICT ROLLBACK
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER NOT NULL,
		aaaa INTEGER NOT NULL IN CONFLICT ROLLBACK
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER UNIQUE,
		aaaa INTEGER UNIQUE ON CONFLICT ROLLBACK,
		aaaa INTEGER UNIQUE ON CONFLICT ABORT,
		aaaa INTEGER UNIQUE ON CONFLICT FAIL,
		aaaa INTEGER UNIQUE ON CONFLICT IGNORE,
		aaaa INTEGER UNIQUE ON CONFLICT REPLACE,
		aaaa integer unique on conflict rollback
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER UNIQUE,
		aaaa INTEGER UNIQUEEEE ON CONFLICT ROLLBACK
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CHECK (),
		aaaa INTEGER CHECK (aaaaaaaaa),
		aaaa integer check (aaaaaaaaa)
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CHECK (),
		aaaa INTEGER CHECKKK (aaaaaaaaa)
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULT +10,
		aaaa INTEGER DEFAULT -10,
		aaaa INTEGER DEFAULT 10,
		aaaa INTEGER DEFAULT 'aaaaa',
		aaaa INTEGER DEFAULT NULL,
		aaaa INTEGER DEFAULT TRUE,
		aaaa INTEGER DEFAULT FALSE,
		aaaa INTEGER DEFAULT CURRENT_TIME,
		aaaa INTEGER DEFAULT CURRENT_DATE,
		aaaa INTEGER DEFAULT CURRENT_TIMESTAMP,
		aaaa integer default null,
		aaaa integer default true,
		aaaa integer default false,
		aaaa integer default current_time,
		aaaa integer default current_date,
		aaaa integer default current_timestamp
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULTTT +10
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULT =10
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULT aaa
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY,
		aaaa INTEGER CONSTRAINT const_uq UNIQUE,
		aaaa INTEGER CONSTRAINT const_nn NOT NULL,
		aaaa INTEGER CONSTRAINT const_de DEFAULT 10,
		aaaa INTEGER CONSTRAINT const_ch CHECK (aaaa),
		aaaa integer constraint const_ch primary key,
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY UNIQUE NOT NULL,
		aaaa INTEGER NOT NULL DEFAULT 10
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY,
		aaaa INTEGER CONSTRAINT const_uq UNIQUE,
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY PRIMARY KEY
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	tokens = tokenize(`CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY,
		aaaa INTEGER CONSTRAINT const_uq UNIQUE,
		aaaa INTEGER CONSTRAINT const_pk NOT NULL NOT NULL
	);`)
	parser = newSQLiteParser(tokens)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
}