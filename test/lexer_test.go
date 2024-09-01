package test

import (
	"testing"
)


func TestLex_SQLite(t *testing.T) {
	tr := NewTester(SQLite, t)

	ddl := ""
	tr.LexOK(ddl, 0)

	/* -------------------------------------------------- */
	ddl = `CREATE TABLE IF NOT EXISTS users (
			"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
			'username' TEXT NOT NULL UNIQUE, * - -2 #aaaaaa
			password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
			email TEXT NOT NULL UNIQUE, /*aaa*/
			created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
			updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
		);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	tr.LexOK(ddl, 85)
	
	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE */
	);`;
	tr.LexNG(ddl, 3, "*")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT, /*
		email TEXT NOT NULL UNIQUE
	);`;
	tr.LexNG(ddl, 4, "<EOF>")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE "
	);`;
	tr.LexNG(ddl, 4, "<EOF>")

	ddl = `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE '
	);`;
	tr.LexNG(ddl, 4, "<EOF>")

	ddl = "CREATE TABLE IF NOT EXISTS `users ();"
	tr.LexNG(ddl, 1, "<EOF>")
}


func TestLex_MySQL(t *testing.T) {
	tr := NewTester(MySQL, t)
	
	ddl := `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		'username' TEXT NOT NULL UNIQUE, * - -2 #aaaaaa
		password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
		email TEXT NOT NULL UNIQUE, /*aaa*/
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
	);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	tr.LexOK(ddl, 84)

}


func TestLex_PostgreSQL(t *testing.T) {
	tr := NewTester(PostgreSQL, t)
	
	ddl := `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,
		'username' TEXT NOT NULL UNIQUE, * - -2 #aaaaaa
		password TEXT NOT NULL DEFAULT "aaaa'bbb'aaaa", --XXX
		email TEXT NOT NULL UNIQUE, /*aaa*/
		created_at TEXT NOT NULL DEFAULT (DATETIME('now', 'localtime')),
		updated_at TEXT NOT NULL DEFAULT(DATETIME('now', 'localtime'))
	);` + "CREATE TABLE IF NOT EXISTS users (`user_id` INTEGER PRIMARY KEY AUTOINCREMENT)"

	tr.LexNG(ddl, 8, "`")

}