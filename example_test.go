// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

func ExampleTemplate_basic() {
	template := New()
	template.ParseString(`{{#foo}}{{bar}}{{/foo}}`)

	context := map[string]interface{}{
		"foo": true,
		"bar": "bazinga!",
	}

	output, _ := template.RenderString(context)
	fmt.Println(output)
	// Output: bazinga!
}

func ExampleTemplate_partials() {
	partial := New(Name("partial"))
	partial.ParseString(`{{bar}}`)

	template := New(Partial(partial))
	template.ParseString(`{{#foo}}{{>partial}}{{/foo}}`)

	context := map[string]interface{}{
		"foo": true,
		"bar": "bazinga!",
	}

	output, _ := template.RenderString(context)
	fmt.Println(output)
	// Output: bazinga!
}

func ExampleTemplate_reader() {
	f, err := os.Open("template.mustache")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file: %s\n", err)
	}
	t, err := Parse(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse template: %s\n", err)
	}
	t.Render(os.Stdout, nil)
}

func ExampleTemplate_http() {
	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com?foo=bar&bar=one&bar=two", nil)

	template := New()
	template.ParseString(`
<ul>{{#foo}}<li>{{.}}</li>{{/foo}}</ul>
<ul>{{#bar}}<li>{{.}}</li>{{/bar}}</ul>`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		template.Render(w, r.URL.Query())
	}

	handler(writer, request)

	fmt.Println(writer.Body.String())
	// Output:
	// <ul><li>bar</li></ul>
	// <ul><li>one</li><li>two</li></ul>
}

func ExampleOption() {
	title := New(Name("header"))   // instantiate and name the template
	title.ParseString("{{title}}") // parse a template string

	body := New()
	body.Option(Name("body")) // options can be defined after we instantiate too
	body.ParseString("{{content}}")

	template := New(
		Delimiters("|", "|"), // set the mustache delimiters to | instead of {{
		Errors(),             // return an error if a variable is missing
		Partial(title),       // register a partial
		Partial(body))        // and another one...

	template.ParseString("|>header|\n|>body|")

	context := map[string]interface{}{
		"title":   "Mustache",
		"content": "Logic less templates with Mustache!",
	}

	template.Render(os.Stdout, context)
	// Output: Mustache
	// Logic less templates with Mustache!
}
