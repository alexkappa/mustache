// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"fmt"
	"os"
)

func ExampleTemplate_basic() {
	template := New()
	template.ParseString(`{{#foo}}{{bar}}{{/foo}}`)

	context := map[string]interface{}{"foo": true, "bar": "bazinga!"}

	output, _ := template.RenderString(context)
	fmt.Println(output)
	// Output: bazinga!
}

func ExampleTemplate_partials() {
	partial := New(Name("partial"))
	partial.ParseString(`{{bar}}`)

	template := New(Partial(partial))
	template.ParseString(`{{#foo}}{{>partial}}{{/foo}}`)

	context := map[string]interface{}{"foo": true, "bar": "bazinga!"}

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

func ExampleOption() {
	title := New(Name("header"))   // instantiate and name the template
	title.ParseString("{{title}}") // parse a template string

	body := New()
	body.Option(Name("body"))
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

	output, _ := template.RenderString(context)
	fmt.Println(output)
	// Output: Mustache
	// Logic less templates with Mustache!
}
