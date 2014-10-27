// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"bytes"
	"testing"
)

func TestParser(t *testing.T) {
	lexer := newLexer("foo {{bar}} {{#alex}}baz{{/alex}} {{!foo}}", "{{", "}}")
	parser := newParser(lexer)
	elems, err := parser.parse()
	if err != nil {
		t.Fatal(err)
	}
	output := bytes.NewBuffer(nil)
	for _, elem := range elems {
		t.Logf("%s\n", elem)
		elem.render(nil, output, map[string]interface{}{"bar": "French Beer Factory", "alex": true})
	}
	t.Logf("%s\n", output.String())
}
