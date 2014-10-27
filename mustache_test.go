// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"bytes"
	"strings"

	"testing"
)

var tests = map[string]interface{}{
	"some text {{foo}} here":                   map[string]string{"foo": "bar"},
	"{{#foo}} foo is defined {{bar}} {{/foo}}": map[string]map[string]string{"foo": {"bar": "baz"}},
	"{{^foo}} foo is defined {{bar}} {{/foo}}": map[string]map[string]string{"foo": {"bar": "baz"}},
}

func TestTemplate(t *testing.T) {
	input := strings.NewReader("some text {{foo}} here")
	template := New()
	err := template.Parse(input)
	if err != nil {
		t.Error(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, map[string]string{"foo": "bar"})
	if err != nil {
		t.Error(err)
	}
	expected := "some text bar here"
	if output.String() != expected {
		t.Errorf("expected %q got %q", expected, output.String())
	}
}

func TestParseTree(t *testing.T) {
	template := New()
	template.elems = []node{
		textNode("Lorem ipsum dolor sit "),
		&varNode{"foo", false},
		textNode(", "),
		&sectionNode{"bar", false, []node{
			&varNode{"baz", true},
			textNode(" adipiscing"),
		}},
		textNode(" elit. Proin commodo viverra elit "),
		&varNode{"zer", false},
		textNode("."),
	}
	data := map[string]interface{}{
		"foo": "amet",
		"bar": map[string]string{"baz": "consectetur"},
		"zer": 0.11,
	}
	b := bytes.NewBuffer(nil)
	for _, e := range template.elems {
		err := e.render(template, b, data)
		if err != nil {
			t.Error(err)
		}
	}

	expected := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Proin commodo viverra elit 0.110000.`

	if expected != b.String() {
		t.Errorf("output didn't match. expected %s got %s.")
		t.Log(b.String())
	}
}
