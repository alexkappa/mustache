# Overview

This is an implementation of the mustache templating language in Go. It is
inspired by [hoisie/mustache](https://github.com/hoisie/mustache) but it's not a fork. It is also inspired by Rob Pike's talk on [lexical scanning in Go](http://rspace.googlecode.com/hg/slide/lex.html). For more information on mustache check the official documentation [here](http://mustache.github.io/).

**Warning:** as of the time of this writing, not all tests pass and this release is not thoroughly tested and verified. Use with caution!

# Installation

Install as usual with `go get github.com/alexkappa/go-mustache`.

# Documentation

The API documentation is available at [godoc.org](http://godoc.org/github.com/alexkappa/go-mustache).

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

To run the tests make sure you checked out the git submodules (which include the specification tests), and run `go test`.

# Contributing

Any sort of contribution is more than welcome.