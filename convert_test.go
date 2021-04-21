package main

import (
	"reflect"
	"testing"
)

func TestEntitiesToMap(t *testing.T) {
	e := []byte(`<!ENTITY num "numeric">
	<!ENTITY pn "pronoun">`)
	exp := make(map[string]string)
	exp["num"] = "numeric"
	exp["pn"] = "pronoun"
	got := entitiesToMap(e)
	if !reflect.DeepEqual(exp, got) {
		t.Errorf("entitiesToMap error: expected %v, got %v", exp, got)
	}
}
