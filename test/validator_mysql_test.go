package test

import (
	"fmt"
	"testing"
)


func TestValidate_MySQL(t *testing.T) {
	tr := NewTester(MySQL, t)

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

	ddl = `CREATE TABLE IF NOT EXISTS AUTO_INCREMENT ();`
	tr.ValidateNG(ddl, 1, "AUTO_INCREMENT")

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
		aaaa integer #aaa
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
	ddl = "create table `scm`.`users` (`aaaa` integer);"
	tr.ValidateOK(ddl)

	ddl = "create table 'scm'.`users` (" + `
		"aaaa" integer
	);`
	tr.ValidateNG(ddl, 1, "'scm'")

	ddl = "create table `scm`.'users' (" + `
		"aaaa" integer
	);`
	tr.ValidateNG(ddl, 1, "'users'")

	ddl = `create table "scm"."users" (
		"aaaa" integer
	);`
	tr.ValidateNG(ddl, 1, "\"scm\"")

	ddl = "create table `scm`.`users` (" + `
		'aaaa' integer
	);`
	tr.ValidateNG(ddl, 2, "'aaaa'")

	ddl = "create table `scm`.`users` (\"aaaa\" integer);"
	tr.ValidateNG(ddl, 1, "\"aaaa\"")

	ddl = `create table "scm.users (
		aaaa integer
	);`
	tr.ValidateNG(ddl, 3, "<EOF>")

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
		aaaa int (10 , 5)
	);`
	tr.ValidateNG(ddl, 3, ",")

	ddl = `create table users (
		aaaa int,
		aaaa time (10, 2)
	);`
	tr.ValidateNG(ddl, 3, ",")

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
	tr.ValidateOK(ddl)

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
	tr.ValidateOK(ddl)

	ddl = `create table users (
		aaaa integer null not null
	);`
	tr.ValidateNG(ddl, 2, "not")

	ddl = `create table users (
		aaaa integer not null null
	);`
	tr.ValidateNG(ddl, 2, "null")

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
	CREATE DATABASE LINK dblink_name
	CONNECT TO user_name IDENTIFIED BY password
	USING 'service_name';
	
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