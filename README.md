# Mustache

[![Go Reference](https://pkg.go.dev/badge/github.com/alexkappa/mustache.svg)](https://pkg.go.dev/github.com/alexkappa/mustache)

![](https://github.com/alexkappa/mustache/actions/workflows/go.yml/badge.svg)

This is an implementation of the mustache templating language in Go.

It is inspired by [hoisie/mustache](https://github.com/hoisie/mustache) however
it's not a fork, but rather a re-implementation with improved spec conformance,
and a more flexible API (e.g. support for `io.Writer` and `io.Reader`).

It is built using lexing techniques described in the slides on [lexical scanning
in Go](https://talks.golang.org/2011/lex.slide), and functional options as
described in the blog post on [self-referential functions and the design of
options](http://commandcenter.blogspot.nl/2014/01/self-referential-functions-and-design.html).

This package aims to cover 100% of the mustache specification tests, however, by
the time of this writing, it is not complete.

For more information on mustache check the [official
documentation](http://mustache.github.io/) and the [mustache
spec](http://github.com/mustache/spec).

# Installation

Install with `go get github.com/alexkappa/mustache`.

# Documentation

The API documentation is available at
[godoc.org](https://pkg.go.dev/github.com/alexkappa/mustache).

# Usage

The core of this package is the `Template`, and its `Parse` and `Render`
functions.

```Go
template := mustache.New()
template.Parse(strings.NewReader("Hello, {{subject}}!"))
template.Render(os.Stdout, map[string]string{"subject": "world"})
```

## Helpers

There are additional `Parse` and `Render` helpers to deal with different kinds
of input or output, such as `string`, `[]byte` or `io.Writer`/`io.Reader`.

```Go
Parse(r io.Reader) error
ParseString(s string) error
ParseBytes(b []byte) error
```

```Go
Render(w io.Writer, context interface{}) error
RenderString(context interface{}) (string, error)
RenderBytes(context interface{}) ([]byte, error)
```

### Reader/Writer

```Go
f, err := os.Open("template.mustache")
if err != nil {
    fmt.Fprintf(os.Stderr, "failed to open file: %s\n", err)
}
t, err := Parse(f)
if err != nil {
    fmt.Fprintf(os.Stderr, "failed to parse template: %s\n", err)
}
t.Render(os.Stdout, nil)
```

**Note:** in the example above, we used
[Parse](https://pkg.go.dev/github.com/alexkappa/mustache#Parse) which wraps the
`t := New()` and `t.Parse()` functions for conciseness.

### String

```Go
t := mustache.New()
err := t.ParseString("Hello, {{subject}}!")
if err != nil {
    // handle error
}
s, _ := t.RenderString(map[string]string{"subject": "world"})
if err != nil {
    // handle error
}
fmt.Println(s)
```

## Options

It is possible to define some options on the template, which will alter the way
the template will parse, render or fail.

The options are:

- `Name(n string) Option` sets the name of the template. This option is useful
  when using the template as a partial to another template.
- `Delimiters(start, end string) Option` sets the start and end delimiters of
  the template.
- `Partial(p *Template) Option` sets p as a partial to the template. It is
  important to set the name of p so that it may be looked up by the parent
  template.
- `SilentMiss(silent bool) Option` sets missing variable lookup behavior.

Options can be defined either as arguments to
[New](https://pkg.go.dev/github.com/alexkappa/mustache#New) or using the
[Option](https://pkg.go.dev/github.com/alexkappa/mustache#Template.Option)
function.

## Partials

Partials are templates themselves and can be defined using the
[Partial](https://pkg.go.dev/github.com/alexkappa/mustache#Partial) option.

**Note:** It is important to name the partial using the
[Name](https://pkg.go.dev/github.com/alexkappa/mustache#Name) option which should
match the mustache partial tag `{{>name}}` in the parent template.

```Go
title := New(
    Name("header")        // instantiate and name the template
    Delimiters("|", "|")) // set the mustache delimiters to | instead of {{

title.ParseString("|title|") // parse a template string

body := New()
body.Option(Name("body"))
body.ParseString("{{content}}")

template := New(
    SilentMiss(false), // return an error if a variable lookup fails
    Partial(title),    // register a partial
    Partial(body))     // and another one...

template.ParseString("{{>header}}\n{{>body}}")

context := map[string]interface{}{
    "title":   "Mustache",
    "content": "Logic less templates with Mustache!",
}

template.Render(os.Stdout, context)
```

## Context

When rendering, context can be either a `map` or a `struct`. Following are some
examples of valid context arguments.

```Go
ctx := map[string]interface{}{
    "foo": "Hello",
    "bar": map[string]string{
        "baz": "World",
    }
}
mustache.Render("{{foo}} {{bar.baz}}", ctx) // Hello World
```

```Go
type Foo struct { Bar string }
ctx := &Foo{ Bar: "Hi, from a struct!" }
mustache.Render("{{Bar}}", ctx) // Hi, from a struct!
```

```Go
type Foo struct { bar string }
func (f *Foo) Bar() string { return f.bar }
ctx := &Foo{"Hi, from a method!"}
mustache.Render("{{Bar}}", ctx) // Hi, from a method!
```

```Go
type Foo struct { Bar string `tag:"bar"` }
ctx := &Foo{ Bar: "Hi, from a struct tag!" }
mustache.Render("{{bar}}", ctx) // Hi, from a struct tag!
```

# Tests

Run `go test` as usual. If you want to run the spec tests against this package,
make sure you've checked out the specs submodule. Otherwise, spec tests will be
skipped.

Currently, certain spec tests are skipped as they fail due to an issue with how
standalone tags and empty lines are being handled. Inspecting them manually, one
can see that the templates render correctly but with some additional `\n` which
should have been omitted. See issue
[#1](http://github.com/alexkappa/mustache/issues/1).

See [SPEC.md](https://github.com/alexkappa/mustache/blob/master/SPEC.md) for a
breakdown of which spec tests pass and fail.

# Contributing

If you would like to contribute, head on to the
[issues](https://github.com/alexkappa/mustache/issues) page for tasks that need
help.
