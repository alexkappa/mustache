// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	for _, test := range []struct {
		template string
		expected []node
	}{
		{
			"{{#foo}}{{#foo}}hello nested{{/foo}}{{/foo}}",
			[]node{
				&sectionNode{"foo", false, []node{
					&sectionNode{"foo", false, []node{
						textNode("hello nested"),
					}},
				}},
			},
		},
		{
			"foo {{bar}} {{#alex}}baz{{/alex}} {{!foo}}",
			[]node{
				textNode("foo "),
				&varNode{"bar", true},
				textNode(" "),
				&sectionNode{"alex", false, []node{
					textNode("baz"),
				}},
				textNode(" "),
				commentNode("foo"),
			},
		},
		{
			"this will{{^foo}}not{{/foo}} be rendered",
			[]node{
				textNode("this will"),
				&sectionNode{"foo", true, []node{
					textNode("not"),
				}},
				textNode(" be rendered"),
			},
		},
		{
			"{{#list}}({{.}}){{/list}}",
			[]node{
				&sectionNode{"list", false, []node{
					textNode("("),
					&varNode{".", true},
					textNode(")"),
				}},
			},
		},
	} {
		parser := newParser(newLexer(test.template, "{{", "}}"))
		elems, err := parser.parse()
		if err != nil {
			t.Fatal(err)
		}
		for i, elem := range elems {
			if !reflect.DeepEqual(elem, test.expected[i]) {
				t.Errorf("elements are not equal %v != %v", elem, test.expected[i])
			}
		}
	}
}
