// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"bytes"
	"os"
	"strings"

	"testing"
)

func TestTemplate(t *testing.T) {
	input := strings.NewReader("some text {{foo}} here")
	template := New()
	err := template.Parse(input)
	if err != nil {
		t.Error(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, map[string]string{"foo": "bar %2B"})
	if err != nil {
		t.Error(err)
	}
	expected := "some text bar %2B here"
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
		&varNode{"foo", noEscape},
		textNode(", "),
		&sectionNode{"bar", false, []node{
			&varNode{"baz", htmlEscape},
			textNode(" adipiscing"),
		}},
		textNode(" elit. Proin commodo viverra elit "),
		&varNode{"zer", noEscape},
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
		err := e.render(template, w, map[node]struct{}{}, data)
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

func TestTemplateJsonEscaped(t *testing.T) {
	input := strings.NewReader("some text {{foo}} here")
	template := New(JsonEscape())
	err := template.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, map[string]string{"foo": "\"bar\"\n<baz> %2B"})
	if err != nil {
		t.Fatal(err)
	}
	expected := "some text \\\"bar\\\"\\n<baz> %2B here"
	if output.String() != expected {
		t.Errorf("expected %q got %q", expected, output.String())
	}
}
func TestObjectOutput(t *testing.T) {
	inputTemplate := strings.NewReader("Raw output here: {{.}}")
	inputData := map[string]map[string]string{"foo": {"bar": "baz"}}
	template := New()
	err := template.Parse(inputTemplate)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, inputData)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Raw output here: {&quot;foo&quot;:{&quot;bar&quot;:&quot;baz&quot;}}"
	if output.String() != expected {
		t.Errorf("expected %q got %q", expected, output.String())
	}
}

func TestObjectOutputJsonEscaped(t *testing.T) {
	inputTemplate := strings.NewReader("Raw output here: {{.}}")
	//inputData := map[string]map[string]string{"foo": {"bår": "baz"}}
	inputData := map[string]map[string]string{"foo": {"bar": "baz %2B"}}
	template := New(JsonEscape())
	err := template.Parse(inputTemplate)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, inputData)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Raw output here: {\\\"foo\\\":{\\\"bar\\\":\\\"baz %2B\\\"}}"
	if output.String() != expected {
		t.Errorf("expected %q got %q", expected, output.String())
	}
}

func TestObjectOutputUnescaped(t *testing.T) {
	inputTemplate := strings.NewReader("Raw output here: {{{.}}}")
	inputData := map[string]map[string]string{"foo": {"bar": "baz %2B"}}
	template := New()
	err := template.Parse(inputTemplate)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, inputData)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Raw output here: {\"foo\":{\"bar\":\"baz %2B\"}}"
	if output.String() != expected {
		t.Errorf("expected %q got %q", expected, output.String())
	}
}

func TestInterpolationWithWhitespace(t *testing.T) {
	input := strings.NewReader("some text {{foo.bar baz.foo}} here")
	template := New()
	err := template.Parse(input)
	if err != nil {
		t.Error(err)
	}
	var output bytes.Buffer
	err = template.Render(&output, map[string]map[string]map[string]string{"foo": {"bar baz": {"foo": "bar %2B"}}})
	if err != nil {
		t.Error(err)
	}
	expected := "some text bar %2B here"
	if output.String() != expected {
		t.Errorf("expected %q got %q", expected, output.String())
	}
}

type templateTest struct {
	template string
	payload  interface{}
	expect   string
}

func TestSectionTestValue(t *testing.T) {
	tests := []templateTest{
		{ // test basic
			`some text {{#test_value {{a}} "value"}}hidden{{/test_value}} here`,
			map[string]string{"a": "value"},
			`some text hidden here`,
		},
		{ // Expect top level context.
			`some text {{#test_value {{a}} "value"}}{{b}}{{/test_value}} here`,
			map[string]string{"a": "value", "b": "hidden"},
			`some text hidden here`,
		},
		{ // Test nesting.
			`some text {{#test_value {{a}} "value"}}{{#test_value {{b}} "hidden"}}thing{{/test_value}}{{/test_value}} here`,
			map[string]string{"a": "value", "b": "hidden"},
			`some text thing here`,
		},
		{ // test nil lookup
			`some text {{#test_value {{aa}} "value"}}hidden{{/test_value}} here`,
			map[string]string{"a": "value"},
			`some text  here`,
		},
		{ // Test nesting normal section
			`some text {{#test_value {{a}} "value"}}{{#b}}{{.}}{{/b}}{{/test_value}} here`,
			map[string]interface{}{"a": "value", "b": []int{1, 2, 3}},
			`some text 123 here`,
		},
	}
	for _, test := range tests {
		input := strings.NewReader(test.template)
		template := New(TestValueSection())
		err := template.Parse(input)
		if err != nil {
			t.Error(err)
		}
		var output bytes.Buffer
		err = template.Render(&output, test.payload)
		if err != nil {
			t.Error(err)
		}
		if output.String() != test.expect {
			t.Errorf("expected %q got %q", test.expect, output.String())
		}

	}

}

func TestInvalidGraph(t *testing.T) {
	template := New(Errors())
	template.name = "Partial"
	p := &partialNode{name: "Partial"}
	template.elems = append(template.elems, p)
	template.partials[template.name] = template

	recursiveErr := template.Render(os.Stdout, nil)
	if recursiveErr == nil {
		t.Error("Partial should've encountered an endless loop bug")
	}
}

func TestValidGraph(t *testing.T) {
	template := New(Errors())
	template.name = "Partial"
	p := &partialNode{name: "Partial"}
	template.elems = append(template.elems, p, p, p)

	recursiveErr := template.Render(os.Stdout, nil)
	if recursiveErr != nil {
		t.Error("Partial shouldn't have encountered an endless loop bug")
	}
}
