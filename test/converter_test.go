package test

import (
	"fmt"
	"testing"
	"encoding/json"
)


func TestConvert_SQLite(t *testing.T) {
	tr := NewTester(SQLite, t)

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
	tr.ConvertOK(ddl)

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
	tables, err := convert(ddl, SQLite)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(tables)
	}
}


func TestConvert_PostgreSQL(t *testing.T) {
	//tr := NewTester(PostgreSQL, t)

	ddl := `create table "scm"."test_table" (
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
		constraint aaaaa primary key(aaa1, aaa2, aaa3) using index tablespace tsn,
		unique(aaa4, aaa5, aaa6) include (bbbb, cccc),
		constraint bbbbb unique(aaa4, aaa5, aaa6) include (bbbb, cccc),
		constraint constraint_zzzz exclude (exclude_element WITH operator, exclude_element WITH operator)
	)
	WITH (aaaaa)
	WITHOUT oids
	TABLESPACE tsn;`
	tables, err := convert(ddl, PostgreSQL)
	if err != nil {
		fmt.Println(err)
	} else {
		jsonData, _ := json.MarshalIndent(tables, "", "    ")
		fmt.Println(string(jsonData))
	}
}


func TestConvert_MySQL(t *testing.T) {
	//tr := NewTester(MySQL, t)

	ddl := `create table users (
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
	tables, err := convert(ddl, MySQL)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(tables)
	}
}