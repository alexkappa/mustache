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

func TestFalsyTemplate(t *testing.T) {
	input := strings.NewReader("some text {{^foo}}{{foo}}{{/foo}} {{bar}} here")
	template := New()
	err := template.Parse(input)
	if err != nil {
		t.Error(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, map[string]interface{}{"foo": 0, "bar": false})
	if err != nil {
		t.Error(err)
	}
	expected := "some text 0 false here"
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
	w := newWriter(b)
	for _, e := range template.elems {
		err := e.render(template, w, data)
		if err != nil {
			t.Error(err)
		}
	}
	w.flush()

	expected := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Proin commodo viverra elit 0.11.`

	if expected != b.String() {
		t.Errorf("output didn't match. expected %q got %q.", expected, b.String())
		t.Log(b.String())
	}
}

func TestSnakeCase(t *testing.T) {
	tmpl := New(SilentMiss(false))
	err := tmpl.ParseString("{{test_snake_case}} {{foo}} {{Bar}}")
	if err != nil {
		t.Error(err)
	}
	type Args struct {
		TestSnakeCase string
		Foo           string
		Bar           string
	}
	str, err := tmpl.RenderString(Args{
		TestSnakeCase: "a",
		Foo:           "b",
		Bar:           "c",
	})
	if err != nil {
		t.Error(err)
	}
	if str != "a b c" {
		t.Errorf("expected %s got %s", "a b c", str)
	}
}

func TestSnakeCaseAndPtr(t *testing.T) {
	tmpl := New(SilentMiss(false))
	err := tmpl.ParseString("{{test_snake_case}} {{foo}} {{Bar}}")
	if err != nil {
		t.Error(err)
	}
	type Args struct {
		TestSnakeCase string
		Foo           string
		Bar           string
	}
	str, err := tmpl.RenderString(&Args{
		TestSnakeCase: "a",
		Foo:           "b",
		Bar:           "c",
	})
	if err != nil {
		t.Error(err)
	}
	if str != "a b c" {
		t.Errorf("expected %s got %s", "a b c", str)
	}
}
