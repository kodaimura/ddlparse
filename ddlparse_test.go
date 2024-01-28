package ddlparse

import (
	"fmt"
	"runtime"
	"testing"
)

type tester struct {
	rdbms Rdbms
	t *testing.T
}

type testerI interface {
	TokenizeOK(ddl string, size int)
	TokenizeNG(ddl string, line int, near string)
	ValidateOK(ddl string)
	ValidateNG(ddl string, line int, near string)
	ParseOK(ddl string)
	ParseNG(ddl string)
} 

func newTester(rdbms Rdbms, t *testing.T) testerI {
	return &tester{rdbms, t}
}


func (te *tester) TokenizeOK(ddl string, size int) {
	_, _, l, _ := runtime.Caller(1)
	tokens, err := tokenize(ddl, te.rdbms)
	if err != nil {
		te.t.Errorf("%d: failed TokenizeOK: %s", l, err.Error())
	} else {
		if len(tokens) != size {
			te.t.Errorf("%d: failed TokenizeOK: Expected (size:%d) But (size:%d)", l, size, len(tokens))
		}
	}
}


func (te *tester) TokenizeNG(ddl string, line int, near string) {
	_, _, l, _ := runtime.Caller(1)
	_, err := tokenize(ddl, te.rdbms)
	if err != nil {
		verr, _ := err.(ValidateError)
		if (verr.Line == line && verr.Near == near) {
			fmt.Println(err.Error())
		}  else {
			te.t.Errorf(
				"%d: failed TokenizeNG: Expected (line:%d, near: %s) But (line:%d, near: %s)",
				l, line, near, verr.Line, verr.Near,
			)
		}
	} else {
		te.t.Errorf("%d: failed TokenizeNG", l)
	}
}

func (te *tester) validate(ddl string) ([]string, error) {
	tokens, err := tokenize(ddl, te.rdbms)
	if err != nil {
		return []string{}, err
	}
	tokens, err = validate(tokens, te.rdbms)
	if err != nil {
		return []string{}, err
	}
	return tokens, nil
}

func (te *tester) ValidateOK(ddl string) {
	_, _, l, _ := runtime.Caller(1)
	_, err := te.validate(ddl)
	if err != nil {
		te.t.Errorf("%d: failed ValidateOK: %s", l, err.Error())
	}
}

func (te *tester) ValidateNG(ddl string, line int, near string) {
	_, _, l, _ := runtime.Caller(1)
	_, err := te.validate(ddl)
	if err != nil {
		verr, _ := err.(ValidateError)
		if (verr.Line == line && verr.Near == near) {
			fmt.Println(err.Error())
		}  else {
			te.t.Errorf(
				"%d: failed ValidateNG: Expected (line:%d, near: %s) But (line:%d, near: %s)",
				l, line, near, verr.Line, verr.Near,
			)
		}
	} else {
		te.t.Errorf("%d: failed ValidateNG", l)
	}
}

func (te *tester) ParseOK(ddl string) {
	_, _, l, _ := runtime.Caller(1)
	tables, err := Parse(ddl, te.rdbms)
	if err != nil {
		te.t.Errorf("%d: failed ParseOK: %s", l, err.Error())
	} else {
		fmt.Println(tables)
	}
}

func (te *tester) ParseNG(ddl string) {
	_, _, l, _ := runtime.Caller(1)
	tables, err := Parse(ddl, te.rdbms)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		te.t.Errorf("%d: failed ParseNG", l)
		fmt.Println(tables)
	}
}


func TestTokenize(t *testing.T) {
	fmt.Println("--------------------------------------------------")
	fmt.Println("TestTokenize")
	fmt.Println("")
	
	test := newTester(SQLite, t)

	ddl := ""
	test.TokenizeOK(ddl, 0)

	/* -------------------------------------------------- */
	ddl = `CREATE TABLE IF NOT EXISTS users (
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
			'username' TEXT NOT NULL UNIQUE, * - -2 #aaaaaa
			password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
			email TEXT NOT NULL UNIQUE, /*aaa*/
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
		);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	test.TokenizeOK(ddl, 85)
	
	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE */
	);`;
	test.TokenizeNG(ddl, 3, "*")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT, /*
		email TEXT NOT NULL UNIQUE
	);`;
	test.TokenizeNG(ddl, 4, ";")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE "
	);`;
	test.TokenizeNG(ddl, 4, ";")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE '
	);`;
	test.TokenizeNG(ddl, 4, ";")

	ddl = "CREATE TABLE IF NOT EXISTS `users ();"
	test.TokenizeNG(ddl, 1, ";")

	/* -------------------------------------------------- */
	test = newTester(MySQL, t)
	
	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		'username' TEXT NOT NULL UNIQUE, * - -2 #aaaaaa
		password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
		email TEXT NOT NULL UNIQUE, /*aaa*/
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
	);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	test.TokenizeOK(ddl, 84)
	/* -------------------------------------------------- */
	test = newTester(PostgreSQL, t)
	
	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		'username' TEXT NOT NULL UNIQUE, * - -2 #aaaaaa
		password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
		email TEXT NOT NULL UNIQUE, /*aaa*/
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
	);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	test.TokenizeNG(ddl, 8, "`")
	/* -------------------------------------------------- */
	
}


func TestValidate_SQLite(t *testing.T) {
	fmt.Println("--------------------------------------------------")
	fmt.Println("TestValidate_SQLite")
	fmt.Println("")
	
	test := newTester(SQLite, t)

	ddl := ""
	test.ValidateOK(ddl)

	/* -------------------------------------------------- */
	ddl = `CREAT TABLE IF NOT EXISTS users ();`
	test.ValidateNG(ddl, 1, "CREAT")

	ddl = `CREATE TABLE IF NOT EXISTS users (

	);`
	test.ValidateNG(ddl, 3, ")")

	ddl = `CREATE TABLE IF EXISTS users ();`
	test.ValidateNG(ddl, 1, "EXISTS")

	ddl = `CREATE TABL IF NOT EXISTS users ();`
	test.ValidateNG(ddl, 1, "TABL")

	ddl = `CREATE TABLE IF NOT EXISTS "users ();`
	test.ValidateNG(ddl, 1, ";")

	ddl = "CREATE TABLE IF NOT EXISTS `users ();"
	test.ValidateNG(ddl, 1, ";")

	ddl = `CREATE TABLE IF NOT EXISTS AUTOINCREMENT ();`
	test.ValidateNG(ddl, 1, "AUTOINCREMENT")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
	);`
	test.ValidateNG(ddl, 3, ")")

	/* -------------------------------------------------- */
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
	
	ddl = `create table users (
		aaaa integer,
		aaaa integer -comment
	);`
	test.ValidateNG(ddl, 3, "-comment")

	ddl = `create table users (
		aaaa integer,
		aaaa integer / * aaa */
	);`
	test.ValidateNG(ddl, 3, "*")

	ddl = `create table users (
		aaaa integer,
		aaaa integer /* aaa
	);`
	test.ValidateNG(ddl, 4, ";")

	ddl = `create table users (
		aaaa integer,
		aaaa integer #aaa
	);`
	test.ValidateNG(ddl, 3, "#aaa")

	ddl = `create table users (
		aaaa integer --aaa,
		aaaa integer 
	);`
	test.ValidateNG(ddl, 3, "aaaa")

	/* -------------------------------------------------- */
	fmt.Println("Schema Name");
	ddl = `create table scm.users (
		aaaa integer
	);`
	test.ValidateOK(ddl)

	ddl = `create table scm users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 1, "users")

	ddl = `create table scm,users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 1, ",")

	/* -------------------------------------------------- */
	fmt.Println("Identifier");
	ddl = `create table "scm"."users" (
		"aaaa" integer
	);`
	test.ValidateOK(ddl)

	ddl = `create table 'scm'."users" (
		"aaaa" integer
	);`
	test.ValidateNG(ddl, 1, "'scm'")

	ddl = `create table "scm".'users' (
		"aaaa" integer
	);`
	test.ValidateNG(ddl, 1, "'users'")

	ddl = `create table "scm"."users" (
		'aaaa' integer
	);`
	test.ValidateNG(ddl, 2, "'aaaa'")

	ddl = "create table `scm`.`users` (`aaaa` integer);"
	test.ValidateOK(ddl)

	ddl = `create table "scm.users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 3, ";")

	ddl = "create table `scm.users (aaaa integer);"
	test.ValidateNG(ddl, 1, ";")

	/* -------------------------------------------------- */
	fmt.Println("Column Date Type")

	ddl = `create table users (
		aaaa integer,
		aaaa text,
		aaaa numeric,
		aaaa integer,
		aaaa real,
		aaaa none
	);`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integerrr
	);`
	test.ValidateNG(ddl, 2, "integerrr")
	
	/* -------------------------------------------------- */
	fmt.Println("Table Option")
	ddl = `create table users (
		aaaa integer
	) without rowid;
	
	create table users (
		aaaa integer
	) strict;

	create table users (
		aaaa integer
	) strict, without rowid;
	
	create table users (
		aaaa integer
	) without rowid, strict;`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer
	) without;`
	test.ValidateNG(ddl, 3, ";")

	ddl = `create table users (
		aaaa integer
	) strict, without rowid, strict;`
	test.ValidateNG(ddl, 3, ",")

	/* -------------------------------------------------- */
	fmt.Println("Table Constraints");
	ddl = `create table users (
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
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer,
		constraintttt check(aaaa)
	);`
	test.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint check check(aaaa)
	);`
	test.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint constraint_zzzz check a
	);`
	test.ValidateNG(ddl, 3, "a")

	ddl = `create table users (
		aaaa integer,
		check(aaa) inherit,
	);`
	test.ValidateNG(ddl, 3, "inherit")

	ddl = `create table users (
		aaaa integer,
		unique
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		unique('aaaa')
	);`
	test.ValidateNG(ddl, 3, "'aaaa'")

	ddl = `create table users (
		aaaa integer,
		primary key
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		primary key('aaaa')
	);`
	test.ValidateNG(ddl, 3, "'aaaa'")

	/* -------------------------------------------------- */
	fmt.Println("Column Constraints");
	ddl = `create table users (
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
		aaaa INTEGER PRIMARY KEY ON CONFLICT ROLLBACK,
		aaaa INTEGER NOT NULL,
		aaaa INTEGER NOT NULL ON CONFLICT ROLLBACK,
		aaaa INTEGER NOT NULL ON CONFLICT ABORT,
		aaaa INTEGER NOT NULL ON CONFLICT FAIL,
		aaaa INTEGER NOT NULL ON CONFLICT IGNORE,
		aaaa INTEGER NOT NULL ON CONFLICT REPLACE,
		aaaa integer not null on conflict rollback,
		aaaa INTEGER UNIQUE,
		aaaa INTEGER UNIQUE ON CONFLICT ROLLBACK,
		aaaa INTEGER UNIQUE ON CONFLICT ABORT,
		aaaa INTEGER UNIQUE ON CONFLICT FAIL,
		aaaa INTEGER UNIQUE ON CONFLICT IGNORE,
		aaaa INTEGER UNIQUE ON CONFLICT REPLACE,
		aaaa integer unique on conflict rollback,
		aaaa INTEGER CHECK (),
		aaaa INTEGER CHECK (aaaaaaaaa),
		aaaa INTEGER CHECK (aaa(aa(a)a())aa),
		aaaa integer check (aaaaaaaaa),
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
		aaaa integer default current_timestamp,
		aaaa INTEGER COLLATE BINARY,
		aaaa INTEGER COLLATE NOCASE,
		aaaa INTEGER COLLATE RTRIM,
		aaaa INTEGER collate binary,
		aaaa INTEGER collate nocase,
		aaaa INTEGER collate rtrim,
		aaaa INTEGER REFERENCES bbb(ccc),
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
		aaaa INTEGER REFERENCES bbb ON DELETE SET NULL MATCH SIMPLE DEFERRABLE INITIALLY IMMEDIATE,
		aaaa INTEGER NOT NULL UNIQUE DEFAULT -10
	);`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer default "aaa"
	);`
	test.ValidateNG(ddl, 2, "\"aaa\"")

	ddl = "create table users (aaaa integer default `aaa`);"
	test.ValidateNG(ddl, 1, "`aaa`")

	ddl = `create table users (
		aaaa integer default aaa
	);`
	test.ValidateNG(ddl, 2, "aaa")

	ddl = `create table users (
		aaaa integer default 'aaa
	);`
	test.ValidateNG(ddl, 3, ";")

	ddl = `create table users (
		aaaa integer default - 2
	);`
	test.ValidateNG(ddl, 2, "-")

	ddl = `create table users (
		aaaa integer unique unique
	);`
	test.ValidateNG(ddl, 2, "unique")

	ddl = `create table users (
		aaaa integer primary key primary key
	);`
	test.ValidateNG(ddl, 2, "primary")
}

func TestValidate_PostgreSQL(t *testing.T) {
	fmt.Println("--------------------------------------------------")
	fmt.Println("TestValidate_PostgreSQL")
	fmt.Println("")

	test := newTester(PostgreSQL, t)

	ddl := ""
	test.ValidateOK(ddl)
	/* -------------------------------------------------- */
	ddl = `CREAT TABLE IF NOT EXISTS users ();`
	test.ValidateNG(ddl, 1, "CREAT")

	ddl = `CREATE TABLE IF NOT EXISTS users (

	);`
	test.ValidateNG(ddl, 3, ")")

	ddl = `CREATE TABLE IF EXISTS users ();`
	test.ValidateNG(ddl, 1, "EXISTS")

	ddl = `CREATE TABL IF NOT EXISTS users ();`
	test.ValidateNG(ddl, 1, "TABL")

	ddl = `CREATE TABLE IF NOT EXISTS "users ();`
	test.ValidateNG(ddl, 1, ";")

	ddl = `CREATE TABLE IF NOT EXISTS 'users ();`
	test.ValidateNG(ddl, 1, ";")

	ddl = `CREATE TABLE IF NOT EXISTS ALLOCATE ();`
	test.ValidateNG(ddl, 1, "ALLOCATE")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
	);`
	test.ValidateNG(ddl, 3, ")")

	/* -------------------------------------------------- */
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
	
	ddl = `create table users (
		aaaa int,
		aaaa integer -comment
	);`
	test.ValidateNG(ddl, 3, "-comment")

	ddl = `create table users (
		aaaa int,
		aaaa integer / * aaa */
	);`
	test.ValidateNG(ddl, 3, "*")

	ddl = `create table users (
		aaaa int,
		aaaa integer /* aaa
	);`
	test.ValidateNG(ddl, 4, ";")

	ddl = `create table users (
		aaaa int,
		aaaa integer #aaa
	);`
	test.ValidateNG(ddl, 3, "#aaa")

	ddl = `create table users (
		aaaa int --aaa,
		aaaa integer 
	);`
	test.ValidateNG(ddl, 3, "aaaa")

	/* -------------------------------------------------- */
	fmt.Println("Schema Name");
	ddl = `create table scm.users (
		aaaa integer
	);`
	test.ValidateOK(ddl)

	ddl = `create table scm users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 1, "users")

	ddl = `create table scm,users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 1, ",")

	/* -------------------------------------------------- */
	fmt.Println("Identifier");
	ddl = `create table "scm"."users" (
		"aaaa" integer
	);`
	test.ValidateOK(ddl)

	ddl = `create table 'scm'."users" (
		"aaaa" integer
	);`
	test.ValidateNG(ddl, 1, "'scm'")

	ddl = `create table "scm".'users' (
		"aaaa" integer
	);`
	test.ValidateNG(ddl, 1, "'users'")

	ddl = `create table "scm"."users" (
		'aaaa' integer
	);`
	test.ValidateNG(ddl, 2, "'aaaa'")

	ddl = "create table `scm`.`users` (`aaaa` integer);"
	test.ValidateNG(ddl, 1, "`")

	ddl = `create table "scm.users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 3, ";")

	/* -------------------------------------------------- */
	fmt.Println("Column Date Type")
	ddl = `create table users (
		aaaa bigint,
		aaaa int8,
		aaaa bigserial,
		aaaa serial8,
		aaaa bit,
		aaaa bit(10),
		aaaa bit varying,
		aaaa bit varying(10),
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
		aaaa numeric(10),
		aaaa numeric(10, 5),
		aaaa decimal,
		aaaa decimal(10),
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

	ddl = `create table users (
		aaaa int,
		aaaa bit var (10)
	);`
	test.ValidateNG(ddl, 3, "var")

	ddl = `create table users (
		aaaa int,
		aaaa numeric ()
	);`
	test.ValidateNG(ddl, 3, ")")

	ddl = `create table users (
		aaaa int,
		aaaa numeric (10, '5')
	);`
	test.ValidateNG(ddl, 3, "'5'")

	ddl = `create table users (
		aaaa int,
		aaaa numeric (10, 5, 2)
	);`
	test.ValidateNG(ddl, 3, ",")

	ddl = `create table users (
		aaaa int,
		aaaa int (10)
	);`
	test.ValidateNG(ddl, 3, "(")

	ddl = `create table users (
		aaaa int,
		aaaa time (10, 2)
	);`
	test.ValidateNG(ddl, 3, ",")

	ddl = `create table users (
		aaaa int,
		aaaa time (10) without time
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa int,
		aaaa time (10) with
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa int,
		aaaa time (10) without time zon
	);`
	test.ValidateNG(ddl, 3, "zon")
	/* -------------------------------------------------- */
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

	ddl = `create table users (
		aaaa integer
	)
	with (aaaaa),
	without oids,
	tablespace tsn;`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer
	),
	with aaaaa,
	without oids,
	tablespaceeee tsn;`
	test.ValidateNG(ddl, 4, "aaaaa")

	ddl = `create table users (
		aaaa integer
	),
	with (aaaaa),
	without oids aaa,
	tablespace tsn;`
	test.ValidateNG(ddl, 5, "aaa")

	ddl = `create table users (
		aaaa integer
	),
	with (aaaaa),
	without oids,
	tablespaceeee tsn;`
	test.ValidateNG(ddl, 6, "tablespaceeee")

	/* -------------------------------------------------- */
	fmt.Println("Table Constraints");
	ddl = `create table users (
		aaaa integer,
		bbbb integer,
		cccc text,
		constraint constraint_zzzz check(aaa),
		check(aaa()'bbb'(aaa)),
		check(aaa) no inherit,
		constraint constraint_zzzz unique(aaaa),
		unique(aaaa),
		unique(aaaa, bbbb, "cccc"),
		unique(aaaa) include (bbbb, cccc),
		unique(aaaa) with (aaaa = value, bbbb = 1),
		unique(aaaa) using index tablespace tsn,
		constraint constraint_zzzz primary key(aaaa),
		primary key(aaaa),
		primary key(aaaa, bbbb, "cccc"),
		primary key(aaaa) include (bbbb, cccc),
		primary key(aaaa) with (aaaa = value, bbbb = 1),
		primary key(aaaa) using index tablespace tsn,
		constraint constraint_zzzz exclude (exclude_element WITH operator, exclude_element WITH operator),
		exclude (exclude_element WITH operator, exclude_element WITH operator),
		exclude using index_method (exclude_element WITH operator),
		exclude using index_method (exclude_element WITH operator) include (bbbb, cccc),
		exclude using index_method (exclude_element WITH operator) include (bbbb, cccc) where (predicate),
		constraint constraint_zzzz foreign key(aaaa) references reftable,
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

	ddl = `create table users (
		aaaa integer,
		constraintttt check(aaaa)
	);`
	test.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint check check(aaaa)
	);`
	test.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint constraint_zzzz check a
	);`
	test.ValidateNG(ddl, 3, "a")

	ddl = `create table users (
		aaaa integer,
		check(aaa) inherit,
	);`
	test.ValidateNG(ddl, 3, "inherit")

	ddl = `create table users (
		aaaa integer,
		unique
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		unique('aaaa')
	);`
	test.ValidateNG(ddl, 3, "'aaaa'")

	ddl = `create table users (
		aaaa integer,
		primary key
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		primary key('aaaa')
	);`
	test.ValidateNG(ddl, 3, "'aaaa'")

	/* -------------------------------------------------- */
	fmt.Println("Column Constraints");
	ddl = `create table users (
		aaaa integer,
		aaaa integer constraint constraint_zzzz not null,
		aaaa integer not null,
		aaaa integer constraint constraint_zzzz null,
		aaaa integer null,
		aaaa integer constraint constraint_zzzz default 1,
		aaaa integer default -1,
		aaaa integer default 'a',
		aaaa integer default null,
		aaaa integer default true,
		aaaa integer default false,
		aaaa integer default current_date,
		aaaa integer default current_time,
		aaaa integer default current_timestamp,
		aaaa integer constraint constraint_zzzz generated always as (generation_expr) stored,
		aaaa integer generated always as (generation_expr) stored,
		aaaa integer generated as identity,
		aaaa integer generated as identity (sequence_options),
		aaaa integer generated always as identity,
		aaaa integer generated always as identity (sequence_options),
		aaaa integer generated by default as identity,
		aaaa integer generated by default as identity (sequence_options),
		aaaa integer constraint constraint_zzzz check(aaa),
		aaaa integer check(aaa()'bbb'(aaa)),
		aaaa integer check(aaa) no inherit,
		aaaa integer constraint constraint_zzzz unique,
		aaaa integer unique,
		aaaa integer unique include (bbbb, cccc),
		aaaa integer unique with (aaaa = value, bbbb = 1),
		aaaa integer unique using index tablespace tsn,
		aaaa integer constraint constraint_zzzz primary key,
		aaaa integer primary key,
		aaaa integer primary key include (bbbb, cccc),
		aaaa integer primary key with (aaaa = value, bbbb = 1),
		aaaa integer primary key using index tablespace tsn,
		aaaa integer constraint constraint_zzzz references reftable,
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

	ddl = `create table users (
		aaaa integer null not null
	);`
	test.ValidateNG(ddl, 2, "not")

	ddl = `create table users (
		aaaa integer not null null
	);`
	test.ValidateNG(ddl, 2, "null")

	ddl = `create table users (
		aaaa integer default "aaa"
	);`
	test.ValidateNG(ddl, 2, "\"aaa\"")

	ddl = `create table users (
		aaaa integer default aaa
	);`
	test.ValidateNG(ddl, 2, "aaa")

	ddl = `create table users (
		aaaa integer default 'aaa
	);`
	test.ValidateNG(ddl, 3, ";")

	ddl = `create table users (
		aaaa integer default - 2
	);`
	test.ValidateNG(ddl, 2, "-")

	ddl = `create table users (
		aaaa integer unique unique
	);`
	test.ValidateNG(ddl, 2, "unique")

	ddl = `create table users (
		aaaa integer primary key primary key
	);`
	test.ValidateNG(ddl, 2, "primary")
}

func TestValidate_MySQL(t *testing.T) {
	fmt.Println("--------------------------------------------------")
	fmt.Println("TestValidate_MySQL")
	fmt.Println("")

	test := newTester(MySQL, t)

	ddl := ""
	test.ValidateOK(ddl)
	/* -------------------------------------------------- */
	ddl = `CREAT TABLE IF NOT EXISTS users ();`
	test.ValidateNG(ddl, 1, "CREAT")

	ddl = `CREATE TABLE IF NOT EXISTS users (

	);`
	test.ValidateNG(ddl, 3, ")")

	ddl = `CREATE TABLE IF EXISTS users ();`
	test.ValidateNG(ddl, 1, "EXISTS")

	ddl = `CREATE TABL IF NOT EXISTS users ();`
	test.ValidateNG(ddl, 1, "TABL")

	ddl = `CREATE TABLE IF NOT EXISTS "users ();`
	test.ValidateNG(ddl, 1, ";")

	ddl = "CREATE TABLE IF NOT EXISTS `users ();"
	test.ValidateNG(ddl, 1, ";")

	ddl = `CREATE TABLE IF NOT EXISTS AUTO_INCREMENT ();`
	test.ValidateNG(ddl, 1, "AUTO_INCREMENT")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
	);`
	test.ValidateNG(ddl, 3, ")")

	/* -------------------------------------------------- */
	fmt.Println("Comment Out");
	ddl = `create table users (
		aaaa integer, --comment
		aaaa integer, /* commen
		comment
		comment
		*/
		aaaa integer #aaa
	);`
	test.ValidateOK(ddl)
	
	ddl = `create table users (
		aaaa int,
		aaaa integer -comment
	);`
	test.ValidateNG(ddl, 3, "-comment")

	ddl = `create table users (
		aaaa int,
		aaaa integer / * aaa */
	);`
	test.ValidateNG(ddl, 3, "*")

	ddl = `create table users (
		aaaa int,
		aaaa integer /* aaa
	);`
	test.ValidateNG(ddl, 4, ";")

	ddl = `create table users (
		aaaa int --aaa,
		aaaa integer 
	);`
	test.ValidateNG(ddl, 3, "aaaa")

	/* -------------------------------------------------- */
	fmt.Println("Schema Name");
	ddl = `create table scm.users (
		aaaa integer
	);`
	test.ValidateOK(ddl)

	ddl = `create table scm users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 1, "users")

	ddl = `create table scm,users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 1, ",")

	/* -------------------------------------------------- */
	fmt.Println("Identifier");
	ddl = "create table `scm`.`users` (`aaaa` integer);"
	test.ValidateOK(ddl)

	ddl = "create table 'scm'.`users` (" + `
		"aaaa" integer
	);`
	test.ValidateNG(ddl, 1, "'scm'")

	ddl = "create table `scm`.'users' (" + `
		"aaaa" integer
	);`
	test.ValidateNG(ddl, 1, "'users'")

	ddl = `create table "scm"."users" (
		"aaaa" integer
	);`
	test.ValidateNG(ddl, 1, "\"scm\"")

	ddl = "create table `scm`.`users` (" + `
		'aaaa' integer
	);`
	test.ValidateNG(ddl, 2, "'aaaa'")

	ddl = "create table `scm`.`users` (\"aaaa\" integer);"
	test.ValidateNG(ddl, 1, "\"aaaa\"")

	ddl = `create table "scm.users (
		aaaa integer
	);`
	test.ValidateNG(ddl, 3, ";")

	/* -------------------------------------------------- */
	fmt.Println("Column Date Type")
	ddl = `create table users (
		aaaa bool,
		aaaa boolean,
		aaaa integer,
		aaaa integer (10),
		aaaa int,
		aaaa int (10),
		aaaa smallint,
		aaaa smallint (10),
		aaaa tinyint,
		aaaa tinyint (10),
		aaaa mediumint,
		aaaa mediumint (10),
		aaaa bigint,
		aaaa bigint (10),
		aaaa numeric,
		aaaa numeric(10),
		aaaa numeric(10, 5),
		aaaa decimal,
		aaaa decimal(10),
		aaaa decimal(10, 5),
		aaaa float,
		aaaa float (10),
		aaaa float (10, 5),
		aaaa real,
		aaaa real (10),
		aaaa real (10, 5),
		aaaa double,
		aaaa double (10),
		aaaa double (10, 5),
		aaaa bit,
		aaaa bit (10),
		aaaa date,
		aaaa datetime,
		aaaa datetime (3),
		aaaa timestamp,
		aaaa timestamp (3),
		aaaa time,
		aaaa time (3),
		aaaa year,
		aaaa year (4),
		aaaa char,
		aaaa char(10),
		aaaa varchar,
		aaaa varchar(10),
		aaaa binary (100),
		aaaa varbinary (100),
		aaaa blob,
		aaaa blob (10),
		aaaa text,
		aaaa text (10),
		aaaa geometry,
		aaaa point,
		aaaa linestring,
		aaaa polygon,
		aaaa multipoint,
		aaaa multilinestring,
		aaaa multipolygon,
		aaaa geometrycollection,
		aaaa json
	);`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa int,
		aaaa bigin
	);`
	test.ValidateNG(ddl, 3, "bigin")

	ddl = `create table users (
		aaaa int,
		aaaa bit var (10)
	);`
	test.ValidateNG(ddl, 3, "var")

	ddl = `create table users (
		aaaa int,
		aaaa numeric ()
	);`
	test.ValidateNG(ddl, 3, ")")

	ddl = `create table users (
		aaaa int,
		aaaa numeric (10, '5')
	);`
	test.ValidateNG(ddl, 3, "'5'")

	ddl = `create table users (
		aaaa int,
		aaaa numeric (10, 5, 2)
	);`
	test.ValidateNG(ddl, 3, ",")

	ddl = `create table users (
		aaaa int,
		aaaa int (10 , 5)
	);`
	test.ValidateNG(ddl, 3, ",")

	ddl = `create table users (
		aaaa int,
		aaaa time (10, 2)
	);`
	test.ValidateNG(ddl, 3, ",")

	/* -------------------------------------------------- */
	fmt.Println("Table Options");
	ddl = `create table users (
		aaaa integer
	)
	AUTOEXTEND_SIZE = 1
	AUTO_INCREMENT = 1
	AVG_ROW_LENGTH = 1
	DEFAULT CHARACTER SET = charset_zzzz
	CHARACTER SET = charset_zzzz
	CHECKSUM = 0
	CHECKSUM = 1
	DEFAULT COLLATE = collation_zzzz
	COLLATE = collation_zzzz
	COMMENT = 'string'
	COMPRESSION = 'ZLIB'
	COMPRESSION = 'LZ4'
	COMPRESSION = 'NONE'
	CONNECTION = 'connect_string'
	DATA DIRECTORY = 'absolute path to directory'
	INDEX DIRECTORY = 'absolute path to directory'
	DELAY_KEY_WRITE = 0
	DELAY_KEY_WRITE = 1
	ENCRYPTION = 'Y' 
	ENCRYPTION = 'N'
	ENGINE = engine_zzzz
	ENGINE_ATTRIBUTE = 'string'
	INSERT_METHOD = NO
	INSERT_METHOD = FIRST
	INSERT_METHOD = LAST
	KEY_BLOCK_SIZE = 1
	MAX_ROWS = 1
	MIN_ROWS = 1
	PACK_KEYS = 0
	PACK_KEYS = 1
	PACK_KEYS = DEFAULT
	PASSWORD = 'string'
	ROW_FORMAT = DEFAULT 
	ROW_FORMAT = DYNAMIC
	ROW_FORMAT = FIXED
	ROW_FORMAT = COMPRESSED
	ROW_FORMAT = REDUNDANT
	ROW_FORMAT = COMPACT
	SECONDARY_ENGINE_ATTRIBUTE = 'string'
	STATS_AUTO_RECALC = DEFAULT 
	STATS_AUTO_RECALC = 0
	STATS_AUTO_RECALC = 1
	STATS_PERSISTENT = DEFAULT
	STATS_PERSISTENT = 0
	STATS_PERSISTENT = 1
	STATS_SAMPLE_PAGES = 1
	TABLESPACE tablespace_zzzz
	TABLESPACE tablespace_zzzz STORAGE DISK
	TABLESPACE tablespace_zzzz STORAGE MEMORY
	UNION = (tbl_yyyy, tbl_zzzz);

	create table users (
		aaaa integer
	)
	AUTOEXTEND_SIZE 1
	AUTO_INCREMENT 1
	AVG_ROW_LENGTH 1
	DEFAULT CHARACTER SET charset_zzzz
	CHARACTER SET charset_zzzz
	CHECKSUM 0
	CHECKSUM 1
	DEFAULT COLLATE collation_zzzz
	COLLATE collation_zzzz
	COMMENT 'string'
	COMPRESSION 'ZLIB'
	COMPRESSION 'LZ4'
	COMPRESSION 'NONE'
	CONNECTION 'connect_string'
	DATA DIRECTORY 'absolute path to directory'
	INDEX DIRECTORY 'absolute path to directory'
	DELAY_KEY_WRITE 0
	DELAY_KEY_WRITE 1
	ENCRYPTION 'Y' 
	ENCRYPTION 'N'
	ENGINE engine_zzzz
	ENGINE_ATTRIBUTE 'string'
	INSERT_METHOD NO
	INSERT_METHOD FIRST
	INSERT_METHOD LAST
	KEY_BLOCK_SIZE 1
	MAX_ROWS 1
	MIN_ROWS 1
	PACK_KEYS 0
	PACK_KEYS 1
	PACK_KEYS DEFAULT
	PASSWORD 'string'
	ROW_FORMAT DEFAULT 
	ROW_FORMAT DYNAMIC
	ROW_FORMAT FIXED
	ROW_FORMAT COMPRESSED
	ROW_FORMAT REDUNDANT
	ROW_FORMAT COMPACT
	SECONDARY_ENGINE_ATTRIBUTE 'string'
	STATS_AUTO_RECALC DEFAULT 
	STATS_AUTO_RECALC 0
	STATS_AUTO_RECALC 1
	STATS_PERSISTENT DEFAULT
	STATS_PERSISTENT 0
	STATS_PERSISTENT 1
	STATS_SAMPLE_PAGES = 1
	TABLESPACE tablespace_zzzz
	TABLESPACE tablespace_zzzz STORAGE DISK
	TABLESPACE tablespace_zzzz STORAGE MEMORY
	UNION (tbl_yyyy, tbl_zzzz);

	create table users (
		aaaa integer
	),
	AUTOEXTEND_SIZE = 1,
	AUTO_INCREMENT = 1,
	AVG_ROW_LENGTH = 1,
	DEFAULT CHARACTER SET = charset_zzzz,
	CHARACTER SET = charset_zzzz;
	
	create table users (
		aaaa integer
	)
	AUTOEXTEND_SIZE = 1 AUTO_INCREMENT = 1 AVG_ROW_LENGTH = 1 DEFAULT CHARACTER SET = charset_zzzz CHARACTER SET = charset_zzzz;`
	test.ValidateOK(ddl)

	/* -------------------------------------------------- */
	fmt.Println("Table Constraints");
	ddl = `create table users (
		aaaa integer,
		bbbb integer,
		cccc text,
		index index_zzzz using btree (aaaa (10) asc, bbbb (10) desc, cccc (10), dddd) KEY_BLOCK_SIZE = 'value' COMMENT 'string',
		index index_zzzz using hash ((expr(zzzz)), (expr(zzzz)) asc, (expr(zzzz)) desc),
		index index_zzzz using hash (aaaa) KEY_BLOCK_SIZE 1 WITH PARSER parser_zzzz VISIBLE INVISIBLE ENGINE_ATTRIBUTE = 'string' SECONDARY_ENGINE_ATTRIBUTE = 'string',
		index index_zzzz (aaaa (10)) using hash,
		index index_zzzz (aaaa (10)),
		index index_zzzz (aaaa),
		index (aaaa (10)) using hash,
		index (aaaa (10)),
		key index_zzzz using btree (aaaa (10) asc, bbbb (10) desc, cccc (10), dddd) KEY_BLOCK_SIZE = 'value' COMMENT 'string',
		key index_zzzz using hash ((expr(zzzz)), (expr(zzzz)) asc, (expr(zzzz)) desc),
		key index_zzzz using hash (aaaa) KEY_BLOCK_SIZE 1 WITH PARSER parser_zzzz VISIBLE INVISIBLE ENGINE_ATTRIBUTE = 'string' SECONDARY_ENGINE_ATTRIBUTE = 'string',
		key index_zzzz (aaaa (10)) using hash,
		key index_zzzz (aaaa (10)),
		key index_zzzz (aaaa),
		key (aaaa (10)) using hash,
		key (aaaa (10)),
		fulltext index index_zzzz (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		fulltext key index_zzzz (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		fulltext index (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		fulltext key (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		fulltext index (aaaa (10)),
		fulltext key (aaaa (10)),
		spatial index index_zzzz (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		spatial key index_zzzz (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		spatial index (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		spatial key (aaaa (10)) KEY_BLOCK_SIZE = 'value',
		spatial index (aaaa (10)),
		spatial key (aaaa (10)),
		constraint constraint_zzzz primary key(aaaa),
		constraint primary key(aaaa),
		primary key using btree (aaaa (10) asc, bbbb (10) desc, cccc (10), dddd) KEY_BLOCK_SIZE = 'value' COMMENT 'string',
		primary key using btree (aaaa (10) asc, bbbb (10) desc, cccc (10)),
		primary key (aaaa (10) asc, bbbb (10) desc, cccc (10)),
		primary key (aaaa, bbbb, cccc),
		constraint constraint_zzzz unique (aaaa),
		constraint unique (aaaa),
		unique index index_zzzz using btree (aaaa (10) asc, bbbb (10) desc, cccc (10), dddd) KEY_BLOCK_SIZE = 'value' COMMENT 'string',
		unique index index_zzzz using hash ((expr(zzzz)), (expr(zzzz)) asc, (expr(zzzz)) desc),
		unique index index_zzzz using hash (aaaa) KEY_BLOCK_SIZE 1 WITH PARSER parser_zzzz VISIBLE INVISIBLE ENGINE_ATTRIBUTE = 'string' SECONDARY_ENGINE_ATTRIBUTE ='string',
		unique index index_zzzz (aaaa (10)) using hash,
		unique index index_zzzz (aaaa (10)),
		unique index index_zzzz (aaaa),
		unique index (aaaa (10)) using hash,
		unique index (aaaa (10)),
		unique key (aaaa (10)),
		unique (aaaa, bbbb, cccc),
		foreign key(aaaa, bbbb, cccc) references reftable (aaaa (10) asc, bbbb (10) desc, cccc (10), dddd),
		foreign key(aaaa, bbbb, dddd) references reftable ((expr(zzzz)), (expr(zzzz)) asc, (expr(zzzz)) desc),
		foreign key(aaaa) references reftable (dddd) match full,
		foreign key(aaaa) references reftable (dddd) match partial,
		foreign key(aaaa) references reftable (dddd) match simple,
		foreign key(aaaa) references reftable (dddd) match full on delete NO ACTION,
		foreign key(aaaa) references reftable (dddd) match full on update RESTRICT,
		foreign key(aaaa) references reftable (dddd) match full on update SET DEFAULT,
		foreign key(aaaa) references reftable (dddd) match full on delete CASCADE on update SET NULL,
		foreign key index_zzzz (aaaa) references reftable (dddd),
		constraint constraint_zzzz check(aaa),
		constraint check(aaa),
		check(aaa()'bbb'(aaa)),
		check(aaa) not enforced,
		check(aaa) enforced
	);`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer,
		constraintttt check(aaaa)
	);`
	test.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint check check(aaaa)
	);`
	test.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint constraint_zzzz check a
	);`
	test.ValidateNG(ddl, 3, "a")

	ddl = `create table users (
		aaaa integer,
		check(aaa) inherit,
	);`
	test.ValidateNG(ddl, 3, "inherit")

	ddl = `create table users (
		aaaa integer,
		unique
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		unique('aaaa')
	);`
	test.ValidateNG(ddl, 3, "'aaaa'")

	ddl = `create table users (
		aaaa integer,
		primary key
	);`
	test.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		primary key('aaaa')
	);`
	test.ValidateNG(ddl, 3, "'aaaa'")

	/* -------------------------------------------------- */
	fmt.Println("Column Constraints");
	ddl = `create table users (
		aaaa integer,
		aaaa integer not null,
		aaaa integer null,
		aaaa integer default -1,
		aaaa integer default 'a',
		aaaa integer default "a",
		aaaa integer default null,
		aaaa integer default true,
		aaaa integer default false,
		aaaa date default current_date,
		aaaa date default current_date on update current_date,
		aaaa time default current_time,
		aaaa time default current_time on update current_time,
		aaaa timestamp default current_timestamp,
		aaaa timestamp default current_timestamp on update current_timestamp,
		aaaa integer default (expr(aaa)),
		aaaa integer visible,
		aaaa integer invisible,
		aaaa integer auto_increment,
		aaaa integer unique,
		aaaa integer unique key,
		aaaa integer primary key,
		aaaa integer key,
		aaaa integer comment 'string',
		aaaa integer comment "string",
		aaaa integer collate collation_zzzz,
		aaaa integer column_format fixed,
		aaaa integer column_format dynamic,
		aaaa integer column_format default,
		aaaa integer engine_attribute = 'string',
		aaaa integer engine_attribute 'string',
		aaaa integer secondary_engine_attribute = 'string',
		aaaa integer secondary_engine_attribute 'string',
		aaaa integer storage disk,
		aaaa integer storage memory,
		aaaa integer references reftable (dddd) match full,
		aaaa integer references reftable (dddd) match partial,
		aaaa integer references reftable (dddd) match simple,
		aaaa integer references reftable (dddd) match full on delete NO ACTION,
		aaaa integer references reftable (dddd) match full on update RESTRICT,
		aaaa integer references reftable (dddd) match full on update SET DEFAULT,
		aaaa integer references reftable (dddd) match full on delete CASCADE on update SET NULL,
		aaaa integer references reftable (dddd),
		aaaa integer constraint constraint_zzzz check(aaa),
		aaaa integer constraint check(aaa),
		aaaa integer check(aaa()'bbb'(aaa)),
		aaaa integer check(aaa) not enforced,
		aaaa integer check(aaa) enforced,
		aaaa integer generated always as (generation_expr),
		aaaa integer as (generation_expr),
		aaaa integer virtual,
		aaaa integer stored,
		aaaa integer virtual,
		aaaa integer not null default -1 visible key
	);`
	test.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer null not null
	);`
	test.ValidateNG(ddl, 2, "not")

	ddl = `create table users (
		aaaa integer not null null
	);`
	test.ValidateNG(ddl, 2, "null")

	ddl = "create table users (aaaa integer default `aaa`);"
	test.ValidateNG(ddl, 1, "`aaa`")

	ddl = `create table users (
		aaaa integer default aaa
	);`
	test.ValidateNG(ddl, 2, "aaa")

	ddl = `create table users (
		aaaa integer default 'aaa
	);`
	test.ValidateNG(ddl, 3, ";")

	ddl = `create table users (
		aaaa integer default - 2
	);`
	test.ValidateNG(ddl, 2, "-")

	ddl = `create table users (
		aaaa integer unique unique
	);`
	test.ValidateNG(ddl, 2, "unique")

	ddl = `create table users (
		aaaa integer primary key primary key
	);`
	test.ValidateNG(ddl, 2, "primary")
}


func TestParse(t *testing.T) {
	test := newTester(SQLite, t)

	ddl := ""
	ddl = `CREATE TABLE IF NOT EXISTS "sch"."project" (
		project_id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_name TEXT NOT NULL,
		project_memo TEXT DEFAULT 'aaaaa"bbb"aaaaa',
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		UNIQUE(project_name, username)
	);`
	test.ParseOK(ddl)

	ddl = `create table scm.test_table (
		id integer primary key asc autoincrement,
		aaa integer not null on conflict fail unique,
		bbb integer default -10,
		ccc none default true,
		ddd none default false,
		eee none default null,
		fff text default (DATETIME('now', 'localtime')),
		ggg text default 'AAA',
		hhh numeric check (aaa(aa(a)a())aa),
		iii real,
		constraint const_name foreign key (a, b, "c") references bbb(ccc) on delete set null
	);`
	tables, err := Parse(ddl, SQLite)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(tables)
	}

	ddl = `create table "scm"."test_table" (
		aaa1 bit,
		aaa2 bit(10),
		aaa3 bit varying,
		aaa4 bit varying(10),
		aaa5 varbit,
		aaa6 varbit(10),
		aaa7 boolean,
		aaa8 bool,
		aaa9 box,
		aa10 bytea constraint constraint_zzzz not null default 1,
		aa11 character default 'aaa',
		aa12 character(10) default null,
		aa13 char default current_timestamp,
		aa14 char(10) default true,
		aa15 character varying,
		aa16 character varying(10),
		aa17 numeric,
		aa18 numeric(10),
		aa19 numeric(10, 5),
		aa20 decimal references reftable (dddd),
		aa21 decimal(10) check(aaa()'bbb'(aaa)),
		aa22 decimal(10, 5) generated always as (generation_expr) stored,
		"aa23" time(10) without time zone,
		aa24 time(10) with time zone,
		primary key(aaa1, aaa2, aaa3) using index tablespace tsn,
		unique(aaa4, aaa5, aaa6) include (bbbb, cccc),
		constraint constraint_zzzz exclude (exclude_element WITH operator, exclude_element WITH operator)
	)
	WITH (aaaaa)
	WITHOUT oids
	TABLESPACE tsn;`
	tables, err = Parse(ddl, PostgreSQL)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(tables)
	}

	ddl = `create table users (
		aaa1 integer primary key auto_increment,
		aaa2 integer (10) key visible not null,
		aaa3 int (10) not null default -1,
		aaa4 smallint (10) null default null,
		aaa5 tinyint (10) default 'a',
		aaa6 mediumint (10) default "a",
		aaa7 bigint (10) default true,
		aaa8 numeric(10) default (expr(aaa)),
		aaa9 numeric(10, 5) unique comment 'string',
		aa10 decimal(10) unique key collate collation_zzzz,
		aa11 decimal(10, 5) column_format fixed,
		aa12 float (10) engine_attribute = 'string',
		aa13 float (10, 5) engine_attribute 'string',
		aa14 real (10) secondary_engine_attribute = 'string',
		aa15 real (10, 5) storage disk,
		aa16 double (10) references reftable (aaaa),
		aa17 double (10, 5) references reftable (dddd) match full on delete CASCADE on update SET NULL,
		aa18 bit (10) check(aaa()'bbb'(aaa)),
		aa19 datetime (3) check(aaa) not enforced,
		aa20 timestamp (3) generated always as (generation_expr),
		aa21 time (3) as (generation_expr),
		aa22 year (4) virtual,
		aa23 char(10) not null default -1 visible key,
		aa24 varchar(10),
		aa25 binary (100),
		aa26 varbinary (100), --aaaaa
		aa27 blob (10),
		aa28 text (10),
		updated_at timestamp with time zone default current_timestamp on update current_timestamp,
		primary key using btree (aa25 (10) asc, aa26 (10) desc, aa27 (10), aa28) KEY_BLOCK_SIZE = 'value' COMMENT 'string',
		unique index index_zzzz using btree (aa25 (10) asc, aa26 (10) desc, aa27 (10), aa28) KEY_BLOCK_SIZE = 'value' COMMENT 'string',
		foreign key(aaaa, bbbb) references reftable (aaaa (10) asc, bbbb (10) desc, cccc (10), dddd),
		constraint constraint_zzzz check(aaa)
	)AUTOEXTEND_SIZE 1
	AUTO_INCREMENT 1
	AVG_ROW_LENGTH 1
	DEFAULT CHARACTER SET charset_zzzz
	CHARACTER SET charset_zzzz
	CHECKSUM 0
	CHECKSUM 1
	DEFAULT COLLATE collation_zzzz
	COLLATE collation_zzzz
	COMMENT 'string'
	COMPRESSION 'ZLIB'
	COMPRESSION 'LZ4'
	COMPRESSION 'NONE'
	CONNECTION 'connect_string'
	DATA DIRECTORY 'absolute path to directory'
	INDEX DIRECTORY 'absolute path to directory'
	DELAY_KEY_WRITE 0
	DELAY_KEY_WRITE 1
	ENCRYPTION 'Y' 
	ENCRYPTION 'N'
	ENGINE engine_zzzz
	ENGINE_ATTRIBUTE 'string'
	INSERT_METHOD NO
	INSERT_METHOD FIRST
	INSERT_METHOD LAST
	KEY_BLOCK_SIZE 1
	MAX_ROWS 1
	MIN_ROWS 1
	PACK_KEYS 0
	PACK_KEYS 1
	PACK_KEYS DEFAULT
	PASSWORD 'string'
	ROW_FORMAT DEFAULT 
	ROW_FORMAT DYNAMIC
	ROW_FORMAT FIXED
	ROW_FORMAT COMPRESSED
	ROW_FORMAT REDUNDANT
	ROW_FORMAT COMPACT
	SECONDARY_ENGINE_ATTRIBUTE 'string'
	STATS_AUTO_RECALC DEFAULT 
	STATS_AUTO_RECALC 0
	STATS_AUTO_RECALC 1
	STATS_PERSISTENT DEFAULT
	STATS_PERSISTENT 0
	STATS_PERSISTENT 1
	STATS_SAMPLE_PAGES = 1
	TABLESPACE tablespace_zzzz
	TABLESPACE tablespace_zzzz STORAGE DISK
	TABLESPACE tablespace_zzzz STORAGE MEMORY
	UNION (tbl_yyyy, tbl_zzzz);`
	tables, err = Parse(ddl, MySQL)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(tables)
	}
}

func TestEnd(t *testing.T) {
	fmt.Println("--------------------------------------------------")
}


func TestWork(t *testing.T) {
	ddl := ""
	ddl = ``

	tokens, err := tokenize(ddl, SQLite)
	if err != nil {
		fmt.Println(err.Error())
	} 
	validatedTokens, err := validate(tokens, SQLite)
	if err != nil {
		fmt.Println(err.Error())
	} 
	fmt.Println(validatedTokens)
}