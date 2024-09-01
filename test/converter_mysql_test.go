package test

import (
	"testing"
)


func TestConvert_MySQL(t *testing.T) {
	tr := NewTester(MySQL, t)

	ddl := `create table test_table (
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
	UNION (tbl_yyyy, tbl_zzzz);
	` + "create table if not exists scm." + "`" + "test_table2" + "`" + "(" + "`" + "aaa1" + "`" + "integer);"

	EXPECT_JSON := `[
		{
		  "schema": "",
		  "name": "test_table",
		  "if_not_exists": false,
		  "columns": [
			{
			  "name": "aaa1",
			  "data_type": {
				"name": "INTEGER",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": true,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": true,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa2",
			  "data_type": {
				"name": "INTEGER",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": true,
				"is_unique": false,
				"is_not_null": true,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa3",
			  "data_type": {
				"name": "INT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": true,
				"is_autoincrement": false,
				"default": -1,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa4",
			  "data_type": {
				"name": "SMALLINT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa5",
			  "data_type": {
				"name": "TINYINT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": "a",
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa6",
			  "data_type": {
				"name": "MEDIUMINT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": "a",
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa7",
			  "data_type": {
				"name": "BIGINT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": true,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa8",
			  "data_type": {
				"name": "NUMERIC",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": "(expr(aaa))",
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aaa9",
			  "data_type": {
				"name": "NUMERIC",
				"digit_n": 10,
				"digit_m": 5
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": true,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa10",
			  "data_type": {
				"name": "DECIMAL",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": true,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "collation_zzzz",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa11",
			  "data_type": {
				"name": "DECIMAL",
				"digit_n": 10,
				"digit_m": 5
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa12",
			  "data_type": {
				"name": "FLOAT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa13",
			  "data_type": {
				"name": "FLOAT",
				"digit_n": 10,
				"digit_m": 5
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa14",
			  "data_type": {
				"name": "REAL",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa15",
			  "data_type": {
				"name": "REAL",
				"digit_n": 10,
				"digit_m": 5
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa16",
			  "data_type": {
				"name": "DOUBLE",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "reftable",
				  "column_names": [
					"aaaa"
				  ]
				}
			  }
			},
			{
			  "name": "aa17",
			  "data_type": {
				"name": "DOUBLE",
				"digit_n": 10,
				"digit_m": 5
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "reftable",
				  "column_names": [
					"dddd"
				  ]
				}
			  }
			},
			{
			  "name": "aa18",
			  "data_type": {
				"name": "BIT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "(aaa()'bbb'(aaa))",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa19",
			  "data_type": {
				"name": "DATETIME",
				"digit_n": 3,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "(aaa)",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa20",
			  "data_type": {
				"name": "TIMESTAMP",
				"digit_n": 3,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa21",
			  "data_type": {
				"name": "TIME",
				"digit_n": 3,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa22",
			  "data_type": {
				"name": "YEAR",
				"digit_n": 4,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa23",
			  "data_type": {
				"name": "CHAR",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": true,
				"is_unique": false,
				"is_not_null": true,
				"is_autoincrement": false,
				"default": -1,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa24",
			  "data_type": {
				"name": "VARCHAR",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa25",
			  "data_type": {
				"name": "BINARY",
				"digit_n": 100,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa26",
			  "data_type": {
				"name": "VARBINARY",
				"digit_n": 100,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa27",
			  "data_type": {
				"name": "BLOB",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa28",
			  "data_type": {
				"name": "TEXT",
				"digit_n": 10,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "updated_at",
			  "data_type": {
				"name": "TIMESTAMP",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": "current_timestamp",
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			}
		  ],
		  "constraints": {
			"primary_key": [
			  {
				"name": "",
				"column_names": [
				  "aa25",
				  "aa26",
				  "aa27",
				  "aa28"
				]
			  }
			],
			"unique": [
			  {
				"name": "",
				"column_names": [
				  "aa25",
				  "aa26",
				  "aa27",
				  "aa28"
				]
			  }
			],
			"check": [
			  {
				"name": "constraint_zzzz",
				"expr": "(aaa)"
			  }
			],
			"foreign_key": [
			  {
				"name": "",
				"column_names": [
				  "aaaa",
				  "bbbb"
				],
				"references": {
				  "table_name": "reftable",
				  "column_names": [
					"aaaa",
					"bbbb",
					"cccc",
					"dddd"
				  ]
				}
			  }
			]
		  }
		},
		{
		  "schema": "scm",
		  "name": "test_table2",
		  "if_not_exists": true,
		  "columns": [
			{
			  "name": "aaa1",
			  "data_type": {
				"name": "INTEGER",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": null,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			}
		  ],
		  "constraints": {
			"primary_key": null,
			"unique": null,
			"check": null,
			"foreign_key": null
		  }
		}
	  ]`

	tr.ConvertOK(ddl, EXPECT_JSON)
}