package sqlparse

import "testing"

func TestParseDdl(t *testing.T) {
	//ParseDdl()
    //t.Errorf("failed")
}

func TestTokenize(t *testing.T) {
	result := tokenize(
		`CREATE TABLE IF NOT EXISTS users (
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime'))
		);`,
	)

	for _, s := range result {
		t.Log(s)
	}
}
