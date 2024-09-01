package test

import (
	"testing"
)


func TestConvert_PostgreSQL(t *testing.T) {
	tr := NewTester(PostgreSQL, t)

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
	TABLESPACE tsn;
	
	create table if not exists test_table2 (
		aaa1 integer
	);`

	EXPECT_JSON := `[
		{
		  "schema": "scm",
		  "name": "test_table",
		  "if_not_exists": false,
		  "columns": [
			{
			  "name": "aaa1",
			  "data_type": {
				"name": "BIT",
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
			},
			{
			  "name": "aaa2",
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
				"name": "VARBIT",
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
			},
			{
			  "name": "aaa4",
			  "data_type": {
				"name": "VARBIT",
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
				"name": "VARBIT",
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
			},
			{
			  "name": "aaa6",
			  "data_type": {
				"name": "VARBIT",
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
			  "name": "aaa7",
			  "data_type": {
				"name": "BOOLEAN",
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
			},
			{
			  "name": "aaa8",
			  "data_type": {
				"name": "BOOL",
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
			},
			{
			  "name": "aaa9",
			  "data_type": {
				"name": "BOX",
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
			},
			{
			  "name": "aa10",
			  "data_type": {
				"name": "BYTEA",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "constraint_zzzz",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": true,
				"is_autoincrement": false,
				"default": 1,
				"check": "",
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "aa11",
			  "data_type": {
				"name": "CHARACTER",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": false,
				"is_autoincrement": false,
				"default": "aaa",
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
				"name": "CHARACTER",
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
				"name": "CHAR",
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
			},
			{
			  "name": "aa14",
			  "data_type": {
				"name": "CHAR",
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
			  "name": "aa15",
			  "data_type": {
				"name": "VARCHAR",
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
			},
			{
			  "name": "aa16",
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
			  "name": "aa17",
			  "data_type": {
				"name": "NUMERIC",
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
			},
			{
			  "name": "aa18",
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
			  "name": "aa19",
			  "data_type": {
				"name": "NUMERIC",
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
			  "name": "aa20",
			  "data_type": {
				"name": "DECIMAL",
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
				  "table_name": "reftable",
				  "column_names": [
					"dddd"
				  ]
				}
			  }
			},
			{
			  "name": "aa21",
			  "data_type": {
				"name": "DECIMAL",
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
			  "name": "aa22",
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
			  "name": "aa23",
			  "data_type": {
				"name": "TIME",
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
			  "name": "aa24",
			  "data_type": {
				"name": "TIME",
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
			}
		  ],
		  "constraints": {
			"primary_key": [
			  {
				"name": "",
				"column_names": [
				  "aaa1",
				  "aaa2",
				  "aaa3"
				]
			  },
			  {
				"name": "aaaaa",
				"column_names": [
				  "aaa1",
				  "aaa2",
				  "aaa3"
				]
			  }
			],
			"unique": [
			  {
				"name": "",
				"column_names": [
				  "aaa4",
				  "aaa5",
				  "aaa6"
				]
			  },
			  {
				"name": "bbbbb",
				"column_names": [
				  "aaa4",
				  "aaa5",
				  "aaa6"
				]
			  }
			],
			"check": null,
			"foreign_key": null
		  }
		},
		{
            "schema": "",
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