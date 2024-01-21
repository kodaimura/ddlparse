package ddlparse

import (
	"fmt"
	"testing"
)

func newTestParser(ddl string) *sqliteParser {
	return &sqliteParser{ddl: ddl}
}

type tester struct {
	rdbms Rdbms
	t *testing.T
}

type testerI interface {
	ValidateOK(ddl string)
	ValidateNG(ddl string, line int, near string)
} 

func newTester(rdbms Rdbms, t *testing.T) testerI {
	return &tester{rdbms, t}
}

func (te *tester) getParser(ddl string) parser {
	if te.rdbms == PostgreSQL {
		return newPostgreSQLParser(ddl)
	} else if te.rdbms == MySQL {
		return newMySQLParser(ddl)
	}
	return newSQLiteParser(ddl)
}

func (te *tester) ValidateOK(ddl string) {
	parser := te.getParser(ddl)
	if err := parser.Validate(); err != nil {
		fmt.Println(fmt.Sprintf("failed validateOK: %s", err.Error()))
		te.t.Errorf("failed")
	}
}

func (te *tester) ValidateNG(ddl string, line int, near string) {
	parser := te.getParser(ddl)
	if err := parser.Validate(); err != nil {
		verr, _ := err.(ValidateError)
		if (verr.Line == line && verr.Near == near) {
			fmt.Println(err.Error())
		}  else {
			te.t.Errorf(
				fmt.Sprintf(
					"failed validateNG: Expected (line:%d, near: %s) But (line:%d, near: %s)",
					line, near, verr.Line, verr.Near,
				))
		}
	} else {
		te.t.Errorf("failed validateNG")
	}
}


func TestTokenize(t *testing.T) {
	ddl := `CREATE TABLE IF NOT EXISTS users (
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
			'username' TEXT NOT NULL UNIQUE, * - -2
			password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
			email TEXT NOT NULL UNIQUE, /*aaa*/
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
		);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	tokens, err := Tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("failed")
	}
	fmt.Println(tokens)
	
	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE */
	);`;

	_, err = Tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT, /*
		email TEXT NOT NULL UNIQUE
	);`;

	_, err = Tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE "
	);`;

	_, err = Tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE '
	);`;

	_, err = Tokenize(ddl, SQLite);
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.Errorf("failed")
	}
}


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
	
	CREATE TABLE IF NOT EXISTS "sch"."project" (
		project_id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_name TEXT NOT NULL,
		project_memo TEXT DEFAULT 'aaaaa"bbb"aaaaa',
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		UNIQUE(project_name, username)
	);`
	parser := newTestParser(ddl)
	tables, err := parser.Parse();
	if err != nil {
		fmt.Println(err.Error())
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


func TestValidate_PostgreSQL(t *testing.T) {
	fmt.Println("--------------------------------------------------")
	fmt.Println("TestValidate_PostgreSQL")
	fmt.Println("")

	test := newTester(PostgreSQL, t)

	fmt.Println("Column Date Type")
	ddl := `create table users (
		aaaa bigint,
		aaaa int8,
		aaaa bigserial,
		aaaa serial8,
		aaaa bit,
		aaaa bit(10),
		aaaa bit varying,
		aaaa varbit,
		aaaa varbit(10),
		aaaa boolean,
		aaaa bool,
		aaaa box,
		aaaa bytea,
		aaaa character,
		aaaa character(10),
		aaaa char,
		aaaa char(10),
		aaaa character varying,
		aaaa character varying(10),
		aaaa varchar,
		aaaa varchar(10),
		aaaa cidr,
		aaaa circle,
		aaaa double precision,
		aaaa float8,
		aaaa inet,
		aaaa integer,
		aaaa int,
		aaaa int4,
		aaaa json,
		aaaa jsonb,
		aaaa line,
		aaaa lseg,
		aaaa macaddr,
		aaaa macaddr8,
		aaaa money,
		aaaa numeric,
		aaaa numeric(10, 5),
		aaaa decimal,
		aaaa decimal(10, 5),
		aaaa path,
		aaaa pg_lsn,
		aaaa pg_snapshot,
		aaaa point,
		aaaa polygon,
		aaaa real,
		aaaa float4,
		aaaa smallint,
		aaaa int2,
		aaaa smallserial,
		aaaa serial2,
		aaaa serial,
		aaaa serial4,
		aaaa text,
		aaaa time,
		aaaa time(10),
		aaaa time(10) without time zone,
		aaaa time(10) with time zone,
		aaaa timetz,
		aaaa timestamp,
		aaaa timestamp(10),
		aaaa timestamp(10) without time zone,
		aaaa timestamp(10) with time zone,
		aaaa timestamptz,
		aaaa tsquery,
		aaaa tsvector,
		aaaa txid_snapshot,
		aaaa uuid,
		aaaa xml
	);`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa int,
		aaaa bigin
	);`
	test.ValidateNG(ddl, 3, "bigin")

	fmt.Println("Comment Out");
	ddl = `create table users (
		aaaa integer, --comment
		aaaa integer, /* commen
		comment
		comment
		*/
		aaaa integer
	);`
	test.ValidateOK(ddl)

	fmt.Println("Schema Name");
	ddl = `create table scm.users (
		aaaa integer
	);`
	test.ValidateOK(ddl)

	fmt.Println("Identifier");
	ddl = `create table "scm"."users" (
		"aaaa" integer
	);`
	test.ValidateOK(ddl)

	fmt.Println("Table Options");
	ddl = `create table users (
		aaaa integer
	),
	with (aaaaa),
	without oids,
	tablespace tsn;
	
	create table users (
		aaaa integer
	)
	with (aaaaa)
	without oids
	tablespace tsn;
	
	CREATE TABLE users (
		aaaa integer
	)
	WITH (aaaaa)
	WITHOUT oids
	TABLESPACE tsn;`
	test.ValidateOK(ddl)

	fmt.Println("Table Constraints");
	ddl = `create table users (
		aaaa integer,
		bbbb integer,
		cccc text,
		constraint constraint_name check(aaa),
		check(aaa()'bbb'(aaa)),
		check(aaa) no inherit,
		constraint constraint_name unique(aaaa),
		unique(aaaa),
		unique(aaaa, bbbb, "cccc"),
		unique(aaaa) include (bbbb, cccc),
		unique(aaaa) with (aaaa = value, bbbb = 1),
		unique(aaaa) using index tablespace tsn,
		constraint constraint_name primary key(aaaa),
		primary key(aaaa),
		primary key(aaaa, bbbb, "cccc"),
		primary key(aaaa) include (bbbb, cccc),
		primary key(aaaa) with (aaaa = value, bbbb = 1),
		primary key(aaaa) using index tablespace tsn,
		constraint constraint_name exclude (exclude_element WITH operator, exclude_element WITH operator),
		exclude (exclude_element WITH operator, exclude_element WITH operator),
		exclude using index_method (exclude_element WITH operator),
		exclude using index_method (exclude_element WITH operator) include (bbbb, cccc),
		exclude using index_method (exclude_element WITH operator) include (bbbb, cccc) where (predicate),
		constraint constraint_name foreign key(aaaa) references reftable,
		foreign key(aaaa, bbbb, "cccc") references reftable,
		foreign key(aaaa, bbbb) references reftable (dddd, eeee),
		foreign key(aaaa) references reftable (dddd) match full,
		foreign key(aaaa) references reftable (dddd) match partial,
		foreign key(aaaa) references reftable (dddd) match simple,
		foreign key(aaaa) references reftable (dddd) match full on delete NO ACTION,
		foreign key(aaaa) references reftable (dddd) match full on update RESTRICT,
		foreign key(aaaa) references reftable (dddd) match full on update SET DEFAULT,
		foreign key(aaaa) references reftable (dddd) match full on delete CASCADE on update SET NULL,
		foreign key(aaaa) references reftable (dddd) match full DEFERRABLE,
		foreign key(aaaa) references reftable (dddd) match full DEFERRABLE INITIALLY DEFERRED,
		foreign key(aaaa) references reftable (dddd) match full NOT DEFERRABLE INITIALLY IMMEDIATE
	);`
	test.ValidateOK(ddl)

	fmt.Println("Column Constraints");
	ddl = `create table users (
		aaaa integer,
		aaaa integer constraint constraint_name not null,
		aaaa integer not null,
		aaaa integer constraint constraint_name null,
		aaaa integer null,
		aaaa integer constraint constraint_name default 1,
		aaaa integer default -1,
		aaaa integer default 'a',
		aaaa integer default null,
		aaaa integer default true,
		aaaa integer default false,
		aaaa integer default current_date,
		aaaa integer default current_time,
		aaaa integer default current_timestamp,
		aaaa integer constraint constraint_name generated always as (generation_expr) stored,
		aaaa integer generated always as (generation_expr) stored,
		aaaa integer generated as identity,
		aaaa integer generated as identity (sequence_options),
		aaaa integer generated always as identity,
		aaaa integer generated always as identity (sequence_options),
		aaaa integer generated by default as identity,
		aaaa integer generated by default as identity (sequence_options),
		aaaa integer constraint constraint_name check(aaa),
		aaaa integer check(aaa()'bbb'(aaa)),
		aaaa integer check(aaa) no inherit,
		aaaa integer constraint constraint_name unique,
		aaaa integer unique,
		aaaa integer unique include (bbbb, cccc),
		aaaa integer unique with (aaaa = value, bbbb = 1),
		aaaa integer unique using index tablespace tsn,
		aaaa integer constraint constraint_name primary key,
		aaaa integer primary key,
		aaaa integer primary key include (bbbb, cccc),
		aaaa integer primary key with (aaaa = value, bbbb = 1),
		aaaa integer primary key using index tablespace tsn,
		aaaa integer constraint constraint_name references reftable,
		aaaa integer references reftable (bbbb),
		aaaa integer references reftable (dddd),
		aaaa integer references reftable (dddd) match full,
		aaaa integer references reftable (dddd) match partial,
		aaaa integer references reftable (dddd) match simple,
		aaaa integer references reftable (dddd) match full on delete NO ACTION,
		aaaa integer references reftable (dddd) match full on update RESTRICT,
		aaaa integer references reftable (dddd) match full on update SET DEFAULT,
		aaaa integer references reftable (dddd) match full on delete CASCADE on update SET NULL,
		aaaa integer references reftable (dddd) match full DEFERRABLE,
		aaaa integer references reftable (dddd) match full DEFERRABLE INITIALLY DEFERRED,
		aaaa integer references reftable (dddd) match full NOT DEFERRABLE INITIALLY IMMEDIATE
	);`
	test.ValidateOK(ddl)
}

func TestEnd(t *testing.T) {
	fmt.Println("--------------------------------------------------")
}