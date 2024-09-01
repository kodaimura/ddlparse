package ddlparse

import (
	"runtime"
	"testing"
	"reflect"
	"encoding/json"
)

func resultCheck(result []Table, expectJson string, t *testing.T) {
	_, _, l, _ := runtime.Caller(1)

	var map1, map2 []map[string]interface{}
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	
	json.Unmarshal([]byte(expectJson), &map1)
	json.Unmarshal([]byte(string(jsonData)), &map2)

	if !reflect.DeepEqual(map1, map2) {
		t.Errorf("%d: failed: \n%s", l, string(jsonData))
	}
}

func TestParseSQlite(t *testing.T) {
	ddl := `
	CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_name TEXT NOT NULL UNIQUE,
		user_password TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
	);
	
	CREATE TRIGGER IF NOT EXISTS trg_users_upd AFTER UPDATE ON users
	BEGIN
	UPDATE users
	SET updated_at = DATETIME('now', 'localtime') 
    WHERE rowid == NEW.rowid;
	END;`

	EXPECT_JSON := `[
		{
		  "schema": "",
		  "name": "users",
		  "if_not_exists": true,
		  "columns": [
			{
			  "name": "user_id",
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
			  "name": "user_name",
			  "data_type": {
				"name": "TEXT",
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
			  "name": "user_password",
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
				"collate": "",
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
		}
	  ]`
	result, _ := Parse(ddl, SQLite)
	resultCheck(result, EXPECT_JSON, t)

	result, _ = ParseSQLite(ddl)
	resultCheck(result, EXPECT_JSON, t)

	result, _ = ParseForce(ddl)
	resultCheck(result, EXPECT_JSON, t)
}


func TestParsePostgreSQL(t *testing.T) {
	ddl := `
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		user_name TEXT NOT NULL UNIQUE,
		user_password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	create trigger trg_users_upd BEFORE UPDATE ON users FOR EACH ROW
  	execute procedure set_update_time();`

	EXPECT_JSON := `[
		{
		  "schema": "",
		  "name": "users",
		  "if_not_exists": true,
		  "columns": [
			{
			  "name": "user_id",
			  "data_type": {
				"name": "SERIAL",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": true,
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
			  "name": "user_name",
			  "data_type": {
				"name": "TEXT",
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
			  "name": "user_password",
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
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "created_at",
			  "data_type": {
				"name": "TIMESTAMP",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": true,
				"is_autoincrement": false,
				"default": "CURRENT_TIMESTAMP",
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
				"is_not_null": true,
				"is_autoincrement": false,
				"default": "CURRENT_TIMESTAMP",
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
	result, _ := Parse(ddl, PostgreSQL)
	resultCheck(result, EXPECT_JSON, t)

	result, _ = ParsePostgreSQL(ddl)
	resultCheck(result, EXPECT_JSON, t)

	result, _ = ParseForce(ddl)
	resultCheck(result, EXPECT_JSON, t)
}


func TestParseMySQL(t *testing.T) {
	ddl := `
	CREATE TABLE IF NOT EXISTS users (
		user_id INT AUTO_INCREMENT PRIMARY KEY,
		user_name VARCHAR(255) NOT NULL UNIQUE,
		user_password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);
	
	CREATE INDEX index_user_name ON users (user_name);`

	EXPECT_JSON := `[
		{
		  "schema": "",
		  "name": "users",
		  "if_not_exists": true,
		  "columns": [
			{
			  "name": "user_id",
			  "data_type": {
				"name": "INT",
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
			  "name": "user_name",
			  "data_type": {
				"name": "VARCHAR",
				"digit_n": 255,
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
			  "name": "user_password",
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
				"collate": "",
				"references": {
				  "table_name": "",
				  "column_names": null
				}
			  }
			},
			{
			  "name": "created_at",
			  "data_type": {
				"name": "TIMESTAMP",
				"digit_n": 0,
				"digit_m": 0
			  },
			  "constraint": {
				"name": "",
				"is_primary_key": false,
				"is_unique": false,
				"is_not_null": true,
				"is_autoincrement": false,
				"default": "CURRENT_TIMESTAMP",
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
				"is_not_null": true,
				"is_autoincrement": false,
				"default": "CURRENT_TIMESTAMP",
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
	result, _ := Parse(ddl, MySQL)
	resultCheck(result, EXPECT_JSON, t)

	result, _ = ParseMySQL(ddl)
	resultCheck(result, EXPECT_JSON, t)

	result, _ = ParseForce(ddl)
	resultCheck(result, EXPECT_JSON, t)
}