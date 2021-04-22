package main

import (
	"reflect"
	"testing"
)

var b = []byte(`<!ENTITY num "numeric"><!ENTITY pn "pronoun">`)

var m = map[string]string{
	"num": "numeric",
	"pn":  "pronoun",
}

func TestEntityBytesToMap(t *testing.T) {
	a := entityBytesToMap(b)
	if !reflect.DeepEqual(m, a) {
		t.Errorf("entityBytesToMap error: expected %v, got %v", m, a)
	}
}
