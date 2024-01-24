package ddlparse

/*
////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

  VALIDATE: 
    Check the syntax of DDL (tokens). 
    And eliminate unnecessary tokens during parsing.
	Return an ValidateError if validation fails.

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

Example:

***** tokens *****
[CREATE TABLE IF NOT users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

***** validatedTokens *****
[CREATE TABLE users ( \n id INTEGER PRIMARY KEY AUTOINCREMENT , \n
 name TEXT NOT NULL , \n password TEXT NOT NULL , \n created_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n updated_at TEXT NOT NULL 
 DEFAULT ( DATETIME ( 'now' , 'localtime' ) ) , \n UNIQUE ( name ) \n ) ;]

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
*/

func validate (tokens []string, rdbms Rdbms) ([]string, error) {
	v := newValidator(rdbms, tokens)
	return v.Validate()
}


type validator interface {
	Validate() ([]string, error)
}

func newValidator(rdbms Rdbms, tokens []string) validator {
	if rdbms == PostgreSQL {
		return newPostgreSQLValidator(tokens)
	} else if rdbms == MySQL {
		return newMySQLValidator(tokens)
	}
	return newSQLiteValidator(tokens)
}