package test

import (
	"fmt"
	"runtime"
	"testing"
	"reflect"
	"encoding/json"

	"github.com/kodaimura/ddlparse/internal/types"
	"github.com/kodaimura/ddlparse/internal/common"
	"github.com/kodaimura/ddlparse/internal/lexer"
	"github.com/kodaimura/ddlparse/internal/validator"
	"github.com/kodaimura/ddlparse/internal/converter"
)

type (
	Table = types.Table
	Column = types.Column
	DataType = types.DataType
	Constraint = types.Constraint
	Reference = types.Reference
	TableConstraint = types.TableConstraint
	PrimaryKey = types.PrimaryKey
	Unique = types.Unique
	Check = types.Check
	ForeignKey = types.ForeignKey
)

type (
	Rdbms = common.Rdbms
	ValidateError = common.ValidateError
)

const (
	PostgreSQL = common.PostgreSQL
	MySQL = common.MySQL
	SQLite = common.SQLite
)


func lex (ddl string, rdbms Rdbms) ([]string, error) {
	l := lexer.NewLexer(rdbms)
	return l.Lex(ddl)
}

func validate (ddl string, rdbms Rdbms) ([]string, error) {
	tokens, err := lex(ddl, rdbms)
	if err != nil {
		return []string{}, err
	}

	v := validator.NewValidator(rdbms)
	return v.Validate(tokens)
}

func convert (ddl string, rdbms Rdbms) ([]Table, error) {
	tokens, err := validate(ddl, rdbms)
	if err != nil {
		return []Table{}, err
	}

	c := converter.NewConverter(rdbms)
	return c.Convert(tokens), nil
}


type tester struct {
	rdbms Rdbms
	t *testing.T
}

type Tester interface {
	LexOK(ddl string, size int)
	LexNG(ddl string, line int, near string)
	ValidateOK(ddl string)
	ValidateNG(ddl string, line int, near string)
	ConvertOK(ddl string, expectJson string)
} 

func NewTester(rdbms Rdbms, t *testing.T) Tester {
	return &tester{rdbms, t}
}


func (te *tester) LexOK(ddl string, size int) {
	_, _, l, _ := runtime.Caller(1)
	tokens, err := lex(ddl, te.rdbms)
	if err != nil {
		te.t.Errorf("%d: failed LexOK: %s", l, err.Error())
	} else {
		if len(tokens) != size {
			te.t.Errorf("%d: failed LexOK: Expected (size:%d) But (size:%d)", l, size, len(tokens))
		}
	}
}


func (te *tester) LexNG(ddl string, line int, near string) {
	_, _, l, _ := runtime.Caller(1)
	_, err :=  lex(ddl, te.rdbms)
	if err != nil {
		verr, _ := err.(ValidateError)
		if (verr.Line == line && verr.Near == near) {
			fmt.Println(err.Error())
		}  else {
			te.t.Errorf(
				"%d: failed LexNG: Expected (line:%d, near: %s) But (line:%d, near: %s)",
				l, line, near, verr.Line, verr.Near,
			)
		}
	} else {
		te.t.Errorf("%d: failed LexNG", l)
	}
}

func (te *tester) ValidateOK(ddl string) {
	_, _, l, _ := runtime.Caller(1)
	_, err := validate(ddl, te.rdbms)
	if err != nil {
		te.t.Errorf("%d: failed ValidateOK: %s", l, err.Error())
	}
}

func (te *tester) ValidateNG(ddl string, line int, near string) {
	_, _, l, _ := runtime.Caller(1)
	_, err := validate(ddl, te.rdbms)
	if err != nil {
		verr, _ := err.(ValidateError)
		if (verr.Line == line && verr.Near == near) {
			fmt.Println(err.Error())
		}  else {
			te.t.Errorf(
				"%d: failed ValidateNG: Expected (line:%d, near: %s) But (line:%d, near: %s)",
				l, line, near, verr.Line, verr.Near,
			)
		}
	} else {
		te.t.Errorf("%d: failed ValidateNG", l)
	}
}

func (te *tester) ConvertOK(ddl string, expectJson string) {
	_, _, l, _ := runtime.Caller(1)
	tables, err := convert(ddl, te.rdbms)
	if err != nil {
		te.t.Errorf("%d: failed ConvertOK: %s", l, err.Error())
	} else {
		var map1, map2 []map[string]interface{}
		jsonData, _ := json.MarshalIndent(tables, "", "  ")
		
		json.Unmarshal([]byte(expectJson), &map1)
		json.Unmarshal([]byte(string(jsonData)), &map2)
		if !reflect.DeepEqual(map1, map2) {
			te.t.Errorf("%d: failed ConvertOK: \n%s", l, string(jsonData))
		}
	}
}