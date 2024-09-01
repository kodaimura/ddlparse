package test

import (
	"fmt"
	"testing"
)


func TestValidate_SQLite(t *testing.T) {
	tr := NewTester(SQLite, t)

	ddl := ""
	tr.ValidateOK(ddl)

	/* -------------------------------------------------- */
	ddl = `CREAT TABLE IF NOT EXISTS users ();`
	tr.ValidateNG(ddl, 1, "CREAT")

	ddl = `CREATE TABLE IF NOT EXISTS users (

	);`
	tr.ValidateNG(ddl, 3, ")")

	ddl = `CREATE TABLE IF EXISTS users ();`
	tr.ValidateNG(ddl, 1, "EXISTS")

	ddl = `CREATE TABL IF NOT EXISTS users ();`
	tr.ValidateNG(ddl, 1, "TABL")

	ddl = `CREATE TABLE IF NOT EXISTS "users ();`
	tr.ValidateNG(ddl, 1, "<EOF>")

	ddl = "CREATE TABLE IF NOT EXISTS `users ();"
	tr.ValidateNG(ddl, 1, "<EOF>")

	ddl = `CREATE TABLE IF NOT EXISTS AUTOINCREMENT ();`
	tr.ValidateNG(ddl, 1, "AUTOINCREMENT")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		aaaa INTEGER,
	);`
	tr.ValidateNG(ddl, 3, ")")

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
	tr.ValidateOK(ddl)
	
	ddl = `create table users (
		aaaa integer,
		aaaa integer -comment
	);`
	tr.ValidateNG(ddl, 3, "-comment")

	ddl = `create table users (
		aaaa integer,
		aaaa integer / * aaa */
	);`
	tr.ValidateNG(ddl, 3, "*")

	ddl = `create table users (
		aaaa integer,
		aaaa integer /* aaa
	);`
	tr.ValidateNG(ddl, 4, "<EOF>")

	ddl = `create table users (
		aaaa integer,
		aaaa integer #aaa
	);`
	tr.ValidateNG(ddl, 3, "#aaa")

	ddl = `create table users (
		aaaa integer --aaa,
		aaaa integer 
	);`
	tr.ValidateNG(ddl, 3, "aaaa")

	/* -------------------------------------------------- */
	fmt.Println("Schema Name");
	ddl = `create table scm.users (
		aaaa integer
	);`
	tr.ValidateOK(ddl)

	ddl = `create table scm users (
		aaaa integer
	);`
	tr.ValidateNG(ddl, 1, "users")

	ddl = `create table scm,users (
		aaaa integer
	);`
	tr.ValidateNG(ddl, 1, ",")

	/* -------------------------------------------------- */
	fmt.Println("Identifier");
	ddl = `create table "scm"."users" (
		"aaaa" integer
	);`
	tr.ValidateOK(ddl)

	ddl = `create table 'scm'."users" (
		"aaaa" integer
	);`
	tr.ValidateNG(ddl, 1, "'scm'")

	ddl = `create table "scm".'users' (
		"aaaa" integer
	);`
	tr.ValidateNG(ddl, 1, "'users'")

	ddl = `create table "scm"."users" (
		'aaaa' integer
	);`
	tr.ValidateNG(ddl, 2, "'aaaa'")

	ddl = "create table `scm`.`users` (`aaaa` integer);"
	tr.ValidateOK(ddl)

	ddl = `create table "scm.users (
		aaaa integer
	);`
	tr.ValidateNG(ddl, 3, "<EOF>")

	ddl = "create table `scm.users (aaaa integer);"
	tr.ValidateNG(ddl, 1, "<EOF>")

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
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integerrr
	);`
	tr.ValidateNG(ddl, 2, "integerrr")
	
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
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer
	) without;`
	tr.ValidateNG(ddl, 3, ";")

	ddl = `create table users (
		aaaa integer
	) strict, without rowid, strict;`
	tr.ValidateNG(ddl, 3, ",")

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
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer,
		constraintttt check(aaaa)
	);`
	tr.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint check check(aaaa)
	);`
	tr.ValidateNG(ddl, 3, "check")

	ddl = `create table users (
		aaaa integer,
		constraint constraint_zzzz check a
	);`
	tr.ValidateNG(ddl, 3, "a")

	ddl = `create table users (
		aaaa integer,
		check(aaa) inherit,
	);`
	tr.ValidateNG(ddl, 3, "inherit")

	ddl = `create table users (
		aaaa integer,
		unique
	);`
	tr.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		unique('aaaa')
	);`
	tr.ValidateNG(ddl, 3, "'aaaa'")

	ddl = `create table users (
		aaaa integer,
		primary key
	);`
	tr.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa integer,
		primary key('aaaa')
	);`
	tr.ValidateNG(ddl, 3, "'aaaa'")

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
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer default "aaa"
	);`
	tr.ValidateNG(ddl, 2, "\"aaa\"")

	ddl = "create table users (aaaa integer default `aaa`);"
	tr.ValidateNG(ddl, 1, "`aaa`")

	ddl = `create table users (
		aaaa integer default aaa
	);`
	tr.ValidateNG(ddl, 2, "aaa")

	ddl = `create table users (
		aaaa integer default 'aaa
	);`
	tr.ValidateNG(ddl, 3, "<EOF>")

	ddl = `create table users (
		aaaa integer default - 2
	);`
	tr.ValidateNG(ddl, 2, "-")

	ddl = `create table users (
		aaaa integer unique unique
	);`
	tr.ValidateNG(ddl, 2, "unique")

	ddl = `create table users (
		aaaa integer primary key primary key
	);`
	tr.ValidateNG(ddl, 2, "primary")

	/* -------------------------------------------------- */
	fmt.Println("Create Other Than Table")
	ddl = `create table users (
		aaaa integer
	) without rowid;
	
	create temp table users (
		aaaa integer
	) strict;

	CREATE TRIGGER update_timestamp
	AFTER UPDATE ON users
	FOR EACH ROW
	BEGIN
		UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;

	CREATE VIRTUAL TABLE documents USING fts5(content);

	CREATE TEMP VIEW temp_active_users AS
	SELECT id, name, email
	FROM users
	WHERE active = 1;

	CREATE INDEX idx_active_users_email ON users(email) WHERE active = 1;
	CREATE UNIQUE INDEX idx_users_unique_email ON users(email);
	
	create table users2 (
		aaaa integer
	) without rowid, strict;`
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer
	) without rowid;

	CREATE TRIGGE update_timestamp
	AFTER UPDATE ON users
	FOR EACH ROW
	BEGIN
		UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`
	tr.ValidateNG(ddl, 5, "TRIGGE")

	/* -------------------------------------------------- */
}