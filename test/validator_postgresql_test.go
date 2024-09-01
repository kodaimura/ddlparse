package test

import (
	"fmt"
	"testing"
)


func TestValidate_PostgreSQL(t *testing.T) {
	tr := NewTester(PostgreSQL, t)

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

	ddl = `CREATE TABLE IF NOT EXISTS 'users ();`
	tr.ValidateNG(ddl, 1, "<EOF>")

	ddl = `CREATE TABLE IF NOT EXISTS ALLOCATE ();`
	tr.ValidateNG(ddl, 1, "ALLOCATE")

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
		aaaa int,
		aaaa integer -comment
	);`
	tr.ValidateNG(ddl, 3, "-comment")

	ddl = `create table users (
		aaaa int,
		aaaa integer / * aaa */
	);`
	tr.ValidateNG(ddl, 3, "*")

	ddl = `create table users (
		aaaa int,
		aaaa integer /* aaa
	);`
	tr.ValidateNG(ddl, 4, "<EOF>")

	ddl = `create table users (
		aaaa int,
		aaaa integer #aaa
	);`
	tr.ValidateNG(ddl, 3, "#aaa")

	ddl = `create table users (
		aaaa int --aaa,
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
	tr.ValidateNG(ddl, 1, "`")

	ddl = `create table "scm.users (
		aaaa integer
	);`
	tr.ValidateNG(ddl, 3, "<EOF>")

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
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa int,
		aaaa bigin
	);`
	tr.ValidateNG(ddl, 3, "bigin")

	ddl = `create table users (
		aaaa int,
		aaaa bit var (10)
	);`
	tr.ValidateNG(ddl, 3, "var")

	ddl = `create table users (
		aaaa int,
		aaaa numeric ()
	);`
	tr.ValidateNG(ddl, 3, ")")

	ddl = `create table users (
		aaaa int,
		aaaa numeric (10, '5')
	);`
	tr.ValidateNG(ddl, 3, "'5'")

	ddl = `create table users (
		aaaa int,
		aaaa numeric (10, 5, 2)
	);`
	tr.ValidateNG(ddl, 3, ",")

	ddl = `create table users (
		aaaa int,
		aaaa int (10)
	);`
	tr.ValidateNG(ddl, 3, "(")

	ddl = `create table users (
		aaaa int,
		aaaa time (10, 2)
	);`
	tr.ValidateNG(ddl, 3, ",")

	ddl = `create table users (
		aaaa int,
		aaaa time (10) without time
	);`
	tr.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa int,
		aaaa time (10) with
	);`
	tr.ValidateNG(ddl, 4, ")")

	ddl = `create table users (
		aaaa int,
		aaaa time (10) without time zon
	);`
	tr.ValidateNG(ddl, 3, "zon")
	/* -------------------------------------------------- */
	fmt.Println("Table Options");
	ddl = `create table users (
		aaaa integer
	),
	with (aaaaa),
	without oids,
	tablespace tsn,
	inherits (parent_table1, parent_table2),
	partition by range (aaaa aaaa aaaa),
	partition by list (aaaa aaaa aaaa),
	partition by hash (aaaa aaaa aaaa),
	using aaaa;
	
	create table users (
		aaaa integer
	)
	with (aaaaa)
	without oids
	tablespace tsn
	PARTITION BY RANGE ( { column_name | ( expression ) } [ COLLATE collation ] [ opclass ] [, ... ] )
	USING aaaa;
	
	CREATE TABLE users (
		aaaa integer
	)
	WITH (aaaaa)
	WITHOUT oids
	TABLESPACE tsn
	INHERITS ( parent_table [, ... ] ) ;`
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer
	)
	with (aaaaa),
	without oids,
	tablespace tsn;`
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer
	),
	with aaaaa,
	without oids,
	tablespaceeee tsn;`
	tr.ValidateNG(ddl, 4, "aaaaa")

	ddl = `create table users (
		aaaa integer
	),
	with (aaaaa),
	without oids aaa,
	tablespace tsn;`
	tr.ValidateNG(ddl, 5, "aaa")

	ddl = `create table users (
		aaaa integer
	),
	with (aaaaa),
	without oids,
	tablespaceeee tsn;`
	tr.ValidateNG(ddl, 6, "tablespaceeee")

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
		foreign key(aaaa) references reftable (dddd) match full on delete CASCADE on update SET NULL
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
		aaaa integer references reftable (dddd) match full on delete CASCADE on update SET NULL
	);`
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer null not null
	);`
	tr.ValidateNG(ddl, 2, "not")

	ddl = `create table users (
		aaaa integer not null null
	);`
	tr.ValidateNG(ddl, 2, "null")

	ddl = `create table users (
		aaaa integer default "aaa"
	);`
	tr.ValidateNG(ddl, 2, "\"aaa\"")

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
	ddl = `create table scm.users (
		aaaa integer
	);

	CREATE TRIGGER update_timestamp
	AFTER UPDATE ON users
	FOR EACH ROW
	BEGIN
		UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;

	CREATE VIEW temp_active_users AS
	SELECT id, name, email
	FROM users
	WHERE active = 1;

	CREATE INDEX idx_active_users_email ON users(email) WHERE active = 1;
	CREATE AGGREGATE aggregate_name (datatype) (
		SFUNC = state_function,
		STYPE = state_data_type,
		FINALFUNC = final_function,
		INITCOND = initial_condition
	);
	
	create table scm.users2 (
		aaaa integer
	);`
	tr.ValidateOK(ddl)

	ddl = `create table scm.users (
		aaaa integer
	);

	CREATE TRIGGE update_timestamp
	AFTER UPDATE ON users
	FOR EACH ROW
	BEGIN
		UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`
	tr.ValidateNG(ddl, 5, "TRIGGE")

	/* -------------------------------------------------- */
}