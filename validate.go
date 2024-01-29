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