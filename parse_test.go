package ddlparse

import (
	"fmt"
	"testing"
)

func newTestParser(ddl string) *sqliteParser {
	return &sqliteParser{ddl: ddl, ddlr: []rune(ddl)}
}


func TestTokenize(t *testing.T) {
	ddl := `CREATE TABLE IF NOT EXISTS users (
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
			'username' TEXT NOT NULL UNIQUE, * -
			password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
			email TEXT NOT NULL UNIQUE, /*aaa*/
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
		);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	tokens, err := tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("failed")
	}
	fmt.Println(tokens)
	
	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE */
	);`;

	_, err = tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT, /*
		email TEXT NOT NULL UNIQUE
	);`;

	_, err = tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE "
	);`;

	_, err = tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE '
	);`;

	_, err = tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
}

//func TestTokenize(t *testing.T) {
//	ddl := `CREATE TABLE IF NOT EXISTS users (
//			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
//			'username' TEXT NOT NULL UNIQUE, * -
//			password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
//			email TEXT NOT NULL UNIQUE, /*aaa*/
//			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
//			updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
//		);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"
//
//	parser := newTestParser(ddl)
//	if err := parser.tokenize(); err != nil {
//		t.Errorf("failed")
//	}
//	fmt.Println(parser.tokens)
//	
//	ddl = `CREATE TABLE IF NOT EXISTS users (
//		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
//		email TEXT NOT NULL UNIQUE */
//	);`;
//
//	parser = newTestParser(ddl)
//	if err := parser.tokenize(); err != nil {
//		fmt.Println(err.Error())
//	} else {
//		t.Errorf("failed")
//	}
//
//	ddl = `CREATE TABLE IF NOT EXISTS users (
//		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT, /*
//		email TEXT NOT NULL UNIQUE
//	);`;
//
//	parser = newTestParser(ddl)
//	if err := parser.tokenize(); err != nil {
//		fmt.Println(err.Error())
//	} else {
//		t.Errorf("failed")
//	}
//
//	ddl = `CREATE TABLE IF NOT EXISTS users (
//		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
//		email TEXT NOT NULL UNIQUE "
//	);`;
//
//	parser = newTestParser(ddl)
//	if err := parser.tokenize(); err != nil {
//		fmt.Println(err.Error())
//	} else {
//		t.Errorf("failed")
//	}
//
//	ddl = `CREATE TABLE IF NOT EXISTS users (
//		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
//		email TEXT NOT NULL UNIQUE '
//	);`;
//
//	parser = newTestParser(ddl)
//	if err := parser.tokenize(); err != nil {
//		fmt.Println(err.Error())
//	} else {
//		t.Errorf("failed")
//	}
//}


func TestValidate(t *testing.T) {
	ddl := `CREATE TABLE IF NOT EXISTS users (

	);`

	parser := newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF EXISTS users ();`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREAT TABLE IF NOT EXISTS users ();`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
	ddl = `CREATE TABL IF NOT EXISTS users ();`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
	ddl = `CREATE TABLE IF NOT EXISTS "users ();`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
	ddl = `CREATE TABLE IF NOT EXISTS AUTOINCREMENT ();`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE users (
		user_id INTEGER
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `create table if not exists users (
		user_id INTEGER
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE "users" (
		user_id INTEGER
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE aaaa.users (
		user_id INTEGER
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE "aaaa"."users" (
		user_id INTEGER
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE aaaa. (
		user_id INTEGER
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
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
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
		bbbb TEXT,
		cccc NUMERICCC
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
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
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		aaaa INTEGER PRIMARY AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ASC AUTOINCREMENT
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		aaaa INTEGER PRIMARY KEY AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON ROLLBACK
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		aaaa INTEGER PRIMARY KEY AUTOINCREMENT,
		aaaa INTEGER PRIMARY KEY ON CONFLICT ROLLBACKKK
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER NOT NULL,
		aaaa INTEGER NOT NULL ON CONFLICT ROLLBACK,
		aaaa INTEGER NOT NULL ON CONFLICT ABORT,
		aaaa INTEGER NOT NULL ON CONFLICT FAIL,
		aaaa INTEGER NOT NULL ON CONFLICT IGNORE,
		aaaa INTEGER NOT NULL ON CONFLICT REPLACE,
		aaaa integer not null on conflict rollback
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER NOT NULL,
		aaaa INTEGER NOT ON CONFLICT ROLLBACK
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER NOT NULL,
		aaaa INTEGER NOT NULL IN CONFLICT ROLLBACK
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER UNIQUE,
		aaaa INTEGER UNIQUE ON CONFLICT ROLLBACK,
		aaaa INTEGER UNIQUE ON CONFLICT ABORT,
		aaaa INTEGER UNIQUE ON CONFLICT FAIL,
		aaaa INTEGER UNIQUE ON CONFLICT IGNORE,
		aaaa INTEGER UNIQUE ON CONFLICT REPLACE,
		aaaa integer unique on conflict rollback
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER UNIQUE,
		aaaa INTEGER UNIQUEEEE ON CONFLICT ROLLBACK
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CHECK (),
		aaaa INTEGER CHECK (aaaaaaaaa),
		aaaa INTEGER CHECK (aaa(aa(a)a())aa),
		aaaa integer check (aaaaaaaaa)
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CHECK (),
		aaaa INTEGER CHECKKK (aaaaaaaaa)
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULT (aaa(aa(a)a())aa),
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
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULTTT +10
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULT =10
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER DEFAULT (),
		aaaa INTEGER DEFAULT (aaaaaaaaa),
		aaaa INTEGER DEFAULT aaa
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER COLLATE BINARY,
		aaaa INTEGER COLLATE NOCASE,
		aaaa INTEGER COLLATE RTRIM,
		aaaa INTEGER collate binary,
		aaaa INTEGER collate nocase,
		aaaa INTEGER collate rtrim
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER COLLATE BINARY,
		aaaa INTEGER COLLATEEE NOCASE
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER COLLATE BINARY,
		aaaa INTEGER COLLATE NOCASEEE
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER REFERENCES bbb(ccc),
		aaaa INTEGER REFERENCES bbb(ccc, ddd),
		aaaa INTEGER REFERENCES bbb(ccc) ON DELETE SET NULL,
		aaaa INTEGER REFERENCES bbb(ccc) ON DELETE SET DEFAULT,
		aaaa INTEGER REFERENCES bbb(ccc) ON UPDATE CASCADE,
		aaaa INTEGER REFERENCES bbb(ccc) ON UPDATE RESTRICT,
		aaaa INTEGER REFERENCES bbb(ccc) ON UPDATE NO ACTION,
		aaaa INTEGER REFERENCES bbb(ccc) MATCH SIMPLE,
		aaaa INTEGER REFERENCES bbb(ccc) MATCH PARTIAL,
		aaaa INTEGER REFERENCES bbb(ccc) MATCH FULL,
		aaaa INTEGER REFERENCES bbb(ccc) DEFERRABLE,
		aaaa INTEGER REFERENCES bbb(ccc) NOT DEFERRABLE,
		aaaa INTEGER REFERENCES bbb(ccc) NOT DEFERRABLE INITIALLY DEFERRED,
		aaaa INTEGER REFERENCES bbb(ccc) NOT DEFERRABLE INITIALLY IMMEDIATE,
		aaaa INTEGER REFERENCES bbb(ccc) ON DELETE SET NULL MATCH SIMPLE DEFERRABLE INITIALLY IMMEDIATE,
		aaaa INTEGER REFERENCES bbb ON DELETE SET NULL MATCH SIMPLE DEFERRABLE INITIALLY IMMEDIATE
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER GENERATED ALWAYS AS (aaa),
		aaaa INTEGER GENERATED ALWAYS AS (aaa) STORED,
		aaaa INTEGER GENERATED ALWAYS AS (aaa) VIRTUAL,
		aaaa INTEGER AS (aaa)
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY,
		aaaa INTEGER CONSTRAINT const_uq UNIQUE,
		aaaa INTEGER CONSTRAINT const_nn NOT NULL,
		aaaa INTEGER CONSTRAINT const_de DEFAULT 10,
		aaaa INTEGER CONSTRAINT const_ch CHECK (aaaa),
		aaaa INTEGER CONSTRAINT const_ch  COLLATE BINARY,
		aaaa integer constraint const_ch primary key,
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY UNIQUE NOT NULL COLLATE BINARY,
		aaaa INTEGER NOT NULL DEFAULT 10
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY,
		aaaa INTEGER CONSTRAINT const_uq UNIQUE,
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY PRIMARY KEY
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER CONSTRAINT const_pk PRIMARY KEY,
		aaaa INTEGER CONSTRAINT const_uq UNIQUE,
		aaaa INTEGER CONSTRAINT const_pk NOT NULL NOT NULL
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		PRIMARY KEY (a),
		CONSTRAINT const_name PRIMARY KEY (a, b, "c"),
		CONSTRAINT const_name PRIMARY KEY (a, b, "c") ON CONFLICT ROLLBACK,
		constraint const_name primary key (a, b, "c") on conflict rollback,
		UNIQUE (a),
		CONSTRAINT const_name UNIQUE (a, b, "c"),
		CONSTRAINT const_name UNIQUE (a, b, "c") ON CONFLICT ROLLBACK,
		constraint const_name unique (a, b, "c") on conflict rollback,
		CHECK (a),
		CONSTRAINT const_name CHECK (aaa(aa(a)a())aa),
		CONSTRAINT const_name check (aaa(aa(a)a())aa),
		FOREIGN KEY (a) REFERENCES bbb(ccc) ON DELETE SET NULL,
		CONSTRAINT const_name FOREIGN KEY (a, b, "c") REFERENCES bbb(ccc) ON DELETE SET NULL,
		constraint const_name foreign key (a, b, "c") references bbb(ccc) on delete set null
	);`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT
	) WITHOUT ROWID;`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT
	) STRICT;`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT
	) STRICT, WITHOUT ROWID;`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT
	) WITHOUT ROWID, STRICT;`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT
	) WITHOUT ROWID`
	parser = newTestParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
}


func TestParse(t *testing.T) {
	ddl := `CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
	);
	
	CREATE TABLE IF NOT EXISTS 'sch'."project" (
		project_id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_name TEXT NOT NULL,
		project_memo TEXT DEFAULT "aaaaa'bbb'aaaaa",
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		UNIQUE(project_name, username)
	);`
	parser := newTestParser(ddl)
	tables, err := parser.Parse();
	if err != nil {
		t.Errorf("failed")
	}
	fmt.Println(tables)

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
		aaaa INTEGER
	);`
	parser = newTestParser(ddl)
	_, err = parser.Parse();
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER PRIMARY KEY,
		PRIMARY KEY(aaaa)
	);`
	parser = newTestParser(ddl)
	_, err = parser.Parse();
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER UNIQUE,
		UNIQUE(aaaa)
	);`
	parser = newTestParser(ddl)
	_, err = parser.Parse();
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER UNIQUE,
		UNIQUE(bbbb)
	);`
	parser = newTestParser(ddl)
	_, err = parser.Parse();
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
}