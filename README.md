# Overview

This is an implementation of the mustache templating language in Go. It is
inspired by [hoisie/mustache](https://github.com/hoisie/mustache) however it's not a fork, rather a re-implementation with an improved API, support for `io.Writer` and `io.Reader` and template parsing using a lexer and parser.

It is built using lexing techniques from Rob Pike's talk on [lexical scanning in Go](http://rspace.googlecode.com/hg/slide/lex.html), and functional options as described by the same author in the blog post on [Self-referential functions and the design of options](http://commandcenter.blogspot.nl/2014/01/self-referential-functions-and-design.html).

For more information on mustache check the official documentation [here](http://mustache.github.io/).

**Warning:** as of the time of this writing, not all tests pass and this release is not thoroughly tested and verified. Use with caution!

# Installation

Install with `go get github.com/alexkappa/mustache`.

# Documentation

The API documentation is available at [godoc.org](http://godoc.org/github.com/alexkappa/mustache).

# Usage

Using mustache is just a matter of creating a new instance of `mustache.Template`, `Parse` and `Render`.

```Go
template := mustache.New()
template.Parse(strings.NewReader("Hello, {{subject}}!"))
template.Render(os.Stdout, map[string]string{"subject": "world"})
```

There are additional `Parse` and `Render` helpers to deal with different kind of input or output, such as `string`, `[]byte` or `io.Writer`/`io.Reader`. These are:

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

# Tests

Run `go test` as usual. If you want to run the spec tests against this package, make sure you've checked out the specs submodule. Otherwise spec tests will be skipped.

# Contributing

If you would like to contribute, head on to the [issues](https://github.com/alexkappa/mustache/issues) page for tasks that need help.