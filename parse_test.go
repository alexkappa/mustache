// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"bytes"
	"testing"
)

var parserTests = []struct {
	template string
	data     interface{}
	expected string
}{
	{
		"{{#foo}}{{#foo}}hello nested{{/foo}}{{/foo}}",
		map[string]bool{"foo": true},
		"hello nested",
	},
	{
		"foo {{bar}} {{#alex}}baz{{/alex}} {{!foo}}",
		map[string]interface{}{"bar": "French Beer Factory", "alex": true},
		"foo French Beer Factory baz ",
	},
	{
		"this will{{^foo}}not{{/foo}} be rendered",
		map[string]interface{}{"foo": true},
		"this will be rendered",
	},
	{
		"{{#list}}({{.}}){{/list}}",
		map[string][]string{"list": {"a", "b", "c", "d", "e"}},
		"(a)(b)(c)(d)(e)",
	},
}

func TestParser(t *testing.T) {
	for _, test := range parserTests {
		parser := newParser(newLexer(test.template, "{{", "}}"))
		elems, err := parser.parse()
		if err != nil {
			t.Fatal(err)
		}
		output := bytes.NewBuffer(nil)
		for _, elem := range elems {
			elem.render(nil, output, test.data)
		}
		if test.expected != output.String() {
			t.Errorf("unexpected output: %q != %q", test.expected, output.String())
		}
	}
}
