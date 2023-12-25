package ddlparse

import (
	"regexp"
	"strings"
)

type postgresqlParser struct {
	tokens []string
	size int
	i int
	line int
}

func newPostgreSQLParser(tokens []string) parser {
	return &postgresqlParser{tokens, len(tokens), 0, 0}
}

func (p *postgresqlParser) isOutOfRange() bool {
	return p.i > p.size - 1
}

func (p *postgresqlParser) syntaxError() error {
	if p.isOutOfRange() {
		return NewValidateError(p.line, p.tokens[p.size - 1])
	}
	return NewValidateError(p.line, p.tokens[0])
}

func (p *postgresqlParser) init() {
	p.i = -1
	p.line = 0
	p.next()
}

func (p *postgresqlParser) next() error {
	p.i += 1
	if (p.isOutOfRange()) {
		return nil;
	}
	if (p.tokens[p.i] == "\n") {
		p.line += 1
		return p.next()
	} else if (p.tokens[p.i] == "--") {
		p.skipSingleLineComment()
		return p.next()
	} else if (p.tokens[p.i] == "/*") {
		if err := p.skipMultiLineComment(); err != nil {
			return err
		}
		return p.next()
	} else {
		return nil
	}
}

func (p *postgresqlParser) skipSingleLineComment() {
	if (p.tokens[p.i] != "--") {
		return
	}
	var skip func()
	skip = func() {
		p.i += 1
		if (p.isOutOfRange()) {
			return
		} else if (p.tokens[p.i] == "\n") {
			p.line += 1
		} else {
			skip()
		}
	}
	skip()
}

func (p *postgresqlParser) skipMultiLineComment() error {
	if (p.tokens[p.i] != "/*") {
		return nil
	}
	var skip func() error
	skip = func() error {
		p.i += 1
		if (p.isOutOfRange()) {
			return p.syntaxError()
		} else if (p.tokens[p.i] == "\n") {
			p.line += 1
			return skip()
		} else if (p.tokens[p.i] == "*/") {
			return nil
		} else {
			return skip()
		}
	}
	return skip()
}

func (p *postgresqlParser) isValidName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !pattern.MatchString(name) {
		return false
	}

	if name != strings.ToLower(name) {
		return !contains(ReservedWords_PostgreSQL, name)
	}
	return !contains(ReservedWords_PostgreSQL, strings.ToUpper(name))
}

func (p *postgresqlParser) isValidQuotedName(name string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9_]*$`)
	return pattern.MatchString(name)
}

func (p *postgresqlParser) Validate() error {
	p.init()
	return nil
}

func (p *postgresqlParser) Parse() ([]Table, error) {
	p.init()
	var tables []Table
	return tables, nil
}



var ReservedWords_PostgreSQL = []string{
	"A",
	"ABS",
	"ADA",
	"ALIAS",
	"ALL",
	"ALLOCATE",
	"ANALYSE",
	"ANALYZE",
	"AND",
	"ANY",
	"ARE",
	"ARRAY",
	"AS",
	"ASC",
	"ASENSITIVE",
	"ASYMMETRIC",
	"ATOMIC",
	"ATTRIBUTE",
	"ATTRIBUTES",
	"AUTHORIZATION",
	"AVG",
	"BASE64",
	"BERNOULLI",
	"BETWEEN",
	"BINARY",
	"BITVAR",
	"BIT_LENGTH",
	"BLOB",
	"BOTH",
	"BREADTH",
	"C",
	"CALL",
	"CARDINALITY",
	"CASE",
	"CAST",
	"CATALOG",
	"CATALOG_NAME",
	"CEIL",
	"CEILING",
	"CHARACTERS",
	"CHARACTER_LENGTH",
	"CHARACTER_SET_CATALOG",
	"CHARACTER_SET_NAME",
	"CHARACTER_SET_SCHEMA",
	"CHAR_LENGTH",
	"CHECK",
	"CHECKED",
	"CLASS_ORIGIN",
	"CLOB",
	"COBOL",
	"COLLATE",
	"COLLATION",
	"COLLATION_CATALOG",
	"COLLATION_NAME",
	"COLLATION_SCHEMA",
	"COLLECT",
	"COLUMN",
	"COLUMN_NAME",
	"COMMAND_FUNCTION",
	"COMMAND_FUNCTION_CODE",
	"COMPLETION",
	"CONDITION",
	"CONDITION_NUMBER",
	"CONNECT",
	"CONNECTION_NAME",
	"CONSTRAINT",
	"CONSTRAINT_CATALOG",
	"CONSTRAINT_NAME",
	"CONSTRAINT_SCHEMA",
	"CONSTRUCTOR",
	"CONTAINS",
	"CONTINUE",
	"CONVERT",
	"CORR",
	"CORRESPONDING",
	"COUNT",
	"COVAR_POP",
	"COVAR_SAMP",
	"CREATE",
	"CROSS",
	"CUBE",
	"CUME_DIST",
	"CURRENT",
	"CURRENT_DATE",
	"CURRENT_DEFAULT_TRANSFORM_GROUP",
	"CURRENT_PATH",
	"CURRENT_ROLE",
	"CURRENT_TIME",
	"CURRENT_TIMESTAMP",
	"CURRENT_TRANSFORM_GROUP_FOR_TYPE",
	"CURRENT_USER",
	"CURSOR_NAME",
	"DATA",
	"DATE",
	"DATETIME_INTERVAL_CODE",
	"DATETIME_INTERVAL_PRECISION",
	"DEFAULT",
	"DEFERRABLE",
	"DEFINED",
	"DEGREE",
	"DENSE_RANK",
	"DEPTH",
	"DEREF",
	"DERIVED",
	"DESC",
	"DESCRIBE",
	"DESCRIPTOR",
	"DESTROY",
	"DESTRUCTOR",
	"DETERMINISTIC",
	"DIAGNOSTICS",
	"DISCONNECT",
	"DISPATCH",
	"DISTINCT",
	"DO",
	"DYNAMIC",
	"DYNAMIC_FUNCTION",
	"DYNAMIC_FUNCTION_CODE",
	"ELEMENT",
	"ELSE",
	"END",
	"END-EXEC",
	"EQUALS",
	"EVERY",
	"EXCEPT",
	"EXCEPTION",
	"EXCLUDE",
	"EXEC",
	"EXISTING",
	"EXP",
	"FALSE",
	"FILTER",
	"FINAL",
	"FLOOR",
	"FOLLOWING",
	"FOR",
	"FOREIGN",
	"FORTRAN",
	"FOUND",
	"FREE",
	"FREEZE",
	"FROM",
	"FULL",
	"FUSION",
	"G",
	"GENERAL",
	"GENERATED",
	"GET",
	"GO",
	"GOTO",
	"GRANT",
	"GROUP",
	"GROUPING",
	"HAVING",
	"HEX",
	"HIERARCHY",
	"HOST",
	"IDENTITY",
	"IGNORE",
	"ILIKE",
	"IMPLEMENTATION",
	"IN",
	"INDICATOR",
	"INFIX",
	"INITIALIZE",
	"INITIALLY",
	"INNER",
	"INSTANCE",
	"INSTANTIABLE",
	"INTERSECT",
	"INTERSECTION",
	"INTO",
	"IS",
	"ISNULL",
	"ITERATE",
	"JOIN",
	"K",
	"KEY_MEMBER",
	"KEY_TYPE",
	"LATERAL",
	"LEADING",
	"LEFT",
	"LENGTH",
	"LESS",
	"LIKE",
	"LIMIT",
	"LN",
	"LOCALTIME",
	"LOCALTIMESTAMP",
	"LOCATOR",
	"LOWER",
	"M",
	"MAP",
	"MATCHED",
	"MAX",
	"MEMBER",
	"MERGE",
	"MESSAGE_LENGTH",
	"MESSAGE_OCTET_LENGTH",
	"MESSAGE_TEXT",
	"METHOD",
	"MIN",
	"MOD",
	"MODIFIES",
	"MODIFY",
	"MODULE",
	"MORE",
	"MULTISET",
	"MUMPS",
	"NATURAL",
	"NCLOB",
	"NESTING",
	"NEW",
	"NORMALIZE",
	"NORMALIZED",
	"NOT",
	"NOTNULL",
	"NULL",
	"NULLABLE",
	"NUMBER",
	"OCTETS",
	"OCTET_LENGTH",
	"OFF",
	"OFFSET",
	"OLD",
	"ON",
	"ONLY",
	"OPEN",
	"OPERATION",
	"OPTIONS",
	"OR",
	"ORDER",
	"ORDERING",
	"ORDINALITY",
	"OTHERS",
	"OUTER",
	"OUTPUT",
	"OVER",
	"OVERLAPS",
	"OVERRIDING",
	"PAD",
	"PARAMETER",
	"PARAMETERS",
	"PARAMETER_MODE",
	"PARAMETER_NAME",
	"PARAMETER_ORDINAL_POSITION",
	"PARAMETER_SPECIFIC_CATALOG",
	"PARAMETER_SPECIFIC_NAME",
	"PARAMETER_SPECIFIC_SCHEMA",
	"PARTITION",
	"PASCAL",
	"PATH",
	"PERCENTILE_CONT",
	"PERCENTILE_DISC",
	"PERCENT_RANK",
	"PLACING",
	"PLI",
	"POSTFIX",
	"POWER",
	"PRECEDING",
	"PREFIX",
	"PREORDER",
	"PRIMARY",
	"PUBLIC",
	"RANGE",
	"RANK",
	"READS",
	"RECURSIVE",
	"REF",
	"REFERENCES",
	"REFERENCING",
	"REGR_AVGX",
	"REGR_AVGY",
	"REGR_COUNT",
	"REGR_INTERCEPT",
	"REGR_R2",
	"REGR_SLOPE",
	"REGR_SXX",
	"REGR_SXY",
	"REGR_SYY",
	"RESULT",
	"RETURN",
	"RETURNED_CARDINALITY",
	"RETURNED_LENGTH",
	"RETURNED_OCTET_LENGTH",
	"RETURNED_SQLSTATE",
	"RETURNING",
	"RIGHT",
	"ROLLUP",
	"ROUTINE",
	"ROUTINE_CATALOG",
	"ROUTINE_NAME",
	"ROUTINE_SCHEMA",
	"ROW_COUNT",
	"ROW_NUMBER",
	"SCALE",
	"SCHEMA_NAME",
	"SCOPE",
	"SCOPE_CATALOG",
	"SCOPE_NAME",
	"SCOPE_SCHEMA",
	"SECTION",
	"SELECT",
	"SELF",
	"SENSITIVE",
	"SERVER_NAME",
	"SESSION_USER",
	"SETS",
	"SIMILAR",
	"SIZE",
	"SOME",
	"SOURCE",
	"SPACE",
	"SPECIFIC",
	"SPECIFICTYPE",
	"SPECIFIC_NAME",
	"SQL",
	"SQLCODE",
	"SQLERROR",
	"SQLEXCEPTION",
	"SQLSTATE",
	"SQLWARNING",
	"SQRT",
	"STATE",
	"STATIC",
	"STDDEV_POP",
	"STDDEV_SAMP",
	"STRUCTURE",
	"STYLE",
	"SUBCLASS_ORIGIN",
	"SUBLIST",
	"SUBMULTISET",
	"SUM",
	"SYMMETRIC",
	"SYSTEM_USER",
	"TABLE",
	"TABLESAMPLE",
	"TABLE_NAME",
	"TERMINATE",
	"THAN",
	"THEN",
	"TIES",
	"TIMEZONE_HOUR",
	"TIMEZONE_MINUTE",
	"TO",
	"TOP_LEVEL_COUNT",
	"TRAILING",
	"TRANSACTIONS_COMMITTED",
	"TRANSACTIONS_ROLLED_BACK",
	"TRANSACTION_ACTIVE",
	"TRANSFORM",
	"TRANSFORMS",
	"TRANSLATE",
	"TRANSLATION",
	"TRIGGER_CATALOG",
	"TRIGGER_NAME",
	"TRIGGER_SCHEMA",
	"TRUE",
	"UESCAPE",
	"UNBOUNDED",
	"UNDER",
	"UNION",
	"UNIQUE",
	"UNNAMED",
	"UNNEST",
	"UPPER",
	"USAGE",
	"USER",
	"USER_DEFINED_TYPE_CATALOG",
	"USER_DEFINED_TYPE_CODE",
	"USER_DEFINED_TYPE_NAME",
	"USER_DEFINED_TYPE_SCHEMA",
	"USING",
	"VARIABLE",
	"VAR_POP",
	"VAR_SAMP",
	"VERBOSE",
	"WHEN",
	"WHENEVER",
	"WHERE",
	"WIDTH_BUCKET",
	"WINDOW",
	"WITH",
	"WITHIN",
	"XMLAGG",
	"XMLBINARY",
	"XMLCOMMENT",
	"XMLNAMESPACES",
}