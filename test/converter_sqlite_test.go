package test

import (
	"testing"
)


func TestConvert_SQLite(t *testing.T) {
	tr := NewTester(SQLite, t)

	ddl := ""
	ddl = `CREATE TABLE IF NOT EXISTS sch.table_name (
		column_name1 INTEGER PRIMARY KEY AUTOINCREMENT,
		column_name2 NUMERIC NOT NULL UNIQUE,
		column_name3 REAL NOT NULL DEFAULT 10,
		"column_name4" NONE REFERENCES table2(col_name),
		column_name5 TEXT NOT NULL COLLATE BINARY,
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
	);
	
	create table "scm"."table_name2" (
		column_name1 integer primary key asc autoincrement,
		column_name2 numeric not null on conflict fail unique,
		column_name3 real default -10,
		column_name4 none default true,
		column_name5 text default false,
		column_name6 text default null,
		column_name7 text default (DATETIME('now', 'localtime')),
		column_name8 text default 'AAA',
		column_name9 text check (aaa(aa(a)a())aa),
		column_name10 text,
		constraint const_name foreign key (a, b, "c") references bbb(ccc) on delete set null,
		primary key (column_name1, column_name2),
		unique (column_name3, column_name4)
	);
	`

	EXPECT_JSON := `[
          {
            "schema": "sch",
            "name": "table_name",
            "if_not_exists": true,
            "columns": [
              {
                "name": "column_name1",
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
                "name": "column_name2",
                "data_type": {
                  "name": "NUMERIC",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": true,
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
                "name": "column_name3",
                "data_type": {
                  "name": "REAL",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": true,
                  "is_autoincrement": false,
                  "default": 10,
                  "check": "",
                  "collate": "",
                  "references": {
                    "table_name": "",
                    "column_names": null
                  }
                }
              },
              {
                "name": "column_name4",
                "data_type": {
                  "name": "NONE",
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
                    "table_name": "table2",
                    "column_names": [
                      "col_name"
                    ]
                  }
                }
              },
              {
                "name": "column_name5",
                "data_type": {
                  "name": "TEXT",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": true,
                  "is_autoincrement": false,
                  "default": null,
                  "check": "",
                  "collate": "BINARY",
                  "references": {
                    "table_name": "",
                    "column_names": null
                  }
                }
              },
              {
                "name": "created_at",
                "data_type": {
                  "name": "TEXT",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": true,
                  "is_autoincrement": false,
                  "default": "(DATETIME('now','localtime'))",
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
                  "name": "TEXT",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": true,
                  "is_autoincrement": false,
                  "default": "(DATETIME('now','localtime'))",
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
          },
          {
            "schema": "scm",
            "name": "table_name2",
            "if_not_exists": false,
            "columns": [
              {
                "name": "column_name1",
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
                "name": "column_name2",
                "data_type": {
                  "name": "NUMERIC",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": true,
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
                "name": "column_name3",
                "data_type": {
                  "name": "REAL",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": false,
                  "is_autoincrement": false,
                  "default": -10,
                  "check": "",
                  "collate": "",
                  "references": {
                    "table_name": "",
                    "column_names": null
                  }
                }
              },
              {
                "name": "column_name4",
                "data_type": {
                  "name": "NONE",
                  "digit_n": 0,
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
                "name": "column_name5",
                "data_type": {
                  "name": "TEXT",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": false,
                  "is_autoincrement": false,
                  "default": false,
                  "check": "",
                  "collate": "",
                  "references": {
                    "table_name": "",
                    "column_names": null
                  }
                }
              },
              {
                "name": "column_name6",
                "data_type": {
                  "name": "TEXT",
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
                "name": "column_name7",
                "data_type": {
                  "name": "TEXT",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": false,
                  "is_autoincrement": false,
                  "default": "(DATETIME('now','localtime'))",
                  "check": "",
                  "collate": "",
                  "references": {
                    "table_name": "",
                    "column_names": null
                  }
                }
              },
              {
                "name": "column_name8",
                "data_type": {
                  "name": "TEXT",
                  "digit_n": 0,
                  "digit_m": 0
                },
                "constraint": {
                  "name": "",
                  "is_primary_key": false,
                  "is_unique": false,
                  "is_not_null": false,
                  "is_autoincrement": false,
                  "default": "AAA",
                  "check": "",
                  "collate": "",
                  "references": {
                    "table_name": "",
                    "column_names": null
                  }
                }
              },
              {
                "name": "column_name9",
                "data_type": {
                  "name": "TEXT",
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
                  "check": "(aaa(aa(a)a())aa)",
                  "collate": "",
                  "references": {
                    "table_name": "",
                    "column_names": null
                  }
                }
              },
              {
                "name": "column_name10",
                "data_type": {
                  "name": "TEXT",
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
              "primary_key": [
                {
                  "name": "",
                  "column_names": [
                    "column_name1",
                    "column_name2"
                  ]
                }
              ],
              "unique": [
                {
                  "name": "",
                  "column_names": [
                    "column_name3",
                    "column_name4"
                  ]
                }
              ],
              "check": null,
              "foreign_key": [
                {
                  "name": "const_name",
                  "column_names": [
                    "a",
                    "b",
                    "\"c\""
                  ],
                  "references": {
                    "table_name": "bbb",
                    "column_names": [
                      "ccc"
                    ]
                  }
                }
              ]
            }
          }
        ]`

	tr.ConvertOK(ddl, EXPECT_JSON)
}

