// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"text/template"
)

// The node type is the base type that represents a node in the parse tree.
type node interface {
	// The render function should be defined by any type wishing to satisfy the
	// node interface. Implementations should be able to render itself to the
	// w Writer with c given as context.
	render(t *Template, w io.Writer, c ...interface{}) error
}

// The textNode type represents a part of the template that is made up solely of
// text. It's an alias to string and it ignores c when rendering.
type textNode string

func (n textNode) render(t *Template, w io.Writer, c ...interface{}) error {
	_, err := w.Write([]byte(n))
	return err
}

func (n textNode) String() string {
	return fmt.Sprintf("text_node: %q", string(n))
}

// The varNode type represents a part of the template that needs to be replaced
// by a variable that exists within c.
type varNode struct {
	name   string
	escape bool
}

func (n *varNode) render(t *Template, w io.Writer, c ...interface{}) error {
	if v, ok := lookup(n.name, c...); ok {
		if n.escape {
			v = template.HTMLEscapeString(fmt.Sprintf("%v", v))
		}
		print(w, v)
		return nil
	}
	return fmt.Errorf("failed to lookup %s", n.name)
}

func (n *varNode) String() string {
	if n.escape {
		return fmt.Sprintf("var_node: {{%s}}", n.name)
	}
	return fmt.Sprintf("var_node: {{{%s}}}", n.name)
}

// The sectionNode type is a complex node which recursively renders its child
// elements while passing along its context along with the global context.
type sectionNode struct {
	name     string
	inverted bool
	elems    []node
}

func (n *sectionNode) render(t *Template, w io.Writer, c ...interface{}) error {
	if v, ok := lookup(n.name, c...); ok {
		for _, elem := range n.elems {
			elem.render(t, w, append(c, v)...)
		}
		return nil
	}
	return fmt.Errorf("failed to lookup %s", n.name)
}

type commentNode string

func (n commentNode) render(t *Template, w io.Writer, c ...interface{}) error {
	return nil
}

func (n commentNode) String() string {
	return fmt.Sprintf("comment_node: %q", string(n))
}

func (n *sectionNode) String() string {
	var buf bytes.Buffer
	buf.WriteString("section_node: {{")
	if n.inverted {
		buf.WriteByte('^')
	} else {
		buf.WriteByte('#')
	}
	buf.WriteString(n.name)
	buf.WriteString("}}")
	for _, elem := range n.elems {
		buf.WriteString(fmt.Sprintf("%s", elem))
	}
	buf.WriteString(fmt.Sprintf("{{/%s}} ", n.name))
	return buf.String()
}

type partialNode struct {
	name string
}

func (p *partialNode) render(t *Template, w io.Writer, c ...interface{}) error {
	if partial, ok := t.partials[p.name]; ok {
		partial.Render(w, c...)
	}
	return nil
}

// The lookup function searches for name inside the v slice using reflection.
func lookup(name string, v ...interface{}) (interface{}, bool) {
	for _, i := range v {
		r := reflect.ValueOf(i)
		switch r.Kind() {
		case reflect.Map:
			mapValue := r.MapIndex(reflect.ValueOf(name))
			if mapValue.IsValid() {
				return mapValue.Interface(), true
			}
		case reflect.Struct:
			fieldValue := r.FieldByName(name)
			if fieldValue.IsValid() {
				return fieldValue.Interface(), truth(fieldValue.Interface())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return r.Int(), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return r.Uint(), true
		case reflect.String:
			return r.String(), true
		}
	}
	return nil, false
}

func truth(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
	case float32:
		return t > 0
	}
	return false
}

// The print function is able to format the interface v and write it to w using
// the best possible formatting flags.
func print(w io.Writer, v interface{}) {
	switch v.(type) {
	case string:
		fmt.Fprintf(w, "%s", v)
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		fmt.Fprintf(w, "%d", v)
	case float32, float64:
		fmt.Fprintf(w, "%f", v)
	default:
		fmt.Fprintf(w, "%v", v)
	}
}

// The Option type describes functional options used with Templates. Check out
// Dave Cheney's talk on functional options http://bit.ly/1x9WWPi.
type Option func(*Template)

// Name sets the name of the template.
func Name(n string) Option {
	return func(t *Template) {
		t.name = n
	}
}

// Delimiters sets the start and end delimiters of the template.
func Delimiters(start, end string) Option {
	return func(t *Template) {
		t.startDelim = start
		t.endDelim = end
	}
}

// Partial sets p as a partial to the template. It is important to set the name
// of p so that it may be looked up by the parent template.
func Partial(p *Template) Option {
	return func(t *Template) {
		t.partials[p.name] = p
	}
}

// Errors enables missing variable errors.
func Errors() Option {
	return func(t *Template) {
		t.silentMiss = false
	}
}

// The Template type represents a template and its components.
type Template struct {
	name       string
	elems      []node
	partials   map[string]*Template
	startDelim string
	endDelim   string
	silentMiss bool
}

// New returns a new Template instance.
func New(options ...Option) *Template {
	t := &Template{
		elems:      make([]node, 0),
		partials:   make(map[string]*Template),
		startDelim: "{{",
		endDelim:   "}}",
		silentMiss: true,
	}
	t.Option(options...)
	return t
}

func (t *Template) Option(options ...Option) {
	for _, optionFn := range options {
		optionFn(t)
	}
}

// Parse parses a stream of bytes read from r and creates a parse tree that
// represents the template.
func (t *Template) Parse(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	l := newLexer(string(b), t.startDelim, t.endDelim)
	p := newParser(l)
	elems, err := p.parse()
	if err != nil {
		return err
	}
	t.elems = elems
	return nil
}

// ParseString is a helper function that uses a string as input.
func (t *Template) ParseString(s string) error {
	return t.Parse(strings.NewReader(s))
}

// ParseBytes is a helper function that uses a byte array as input.
func (t *Template) ParseBytes(b []byte) error {
	return t.Parse(bytes.NewReader(b))
}

// Render walks through the template's parse tree and writes the output to w
// replacing the values found in context.
func (t *Template) Render(w io.Writer, context ...interface{}) error {
	for _, elem := range t.elems {
		err := elem.render(t, w, context...)
		if err != nil {
			if !t.silentMiss {
				return err
			}
		}
	}
	return nil
}

// RenderString is a helper function that renders the template as a string.
func (t *Template) RenderString(context ...interface{}) (string, error) {
	b := &bytes.Buffer{}
	err := t.Render(b, context...)
	return b.String(), err
}

// RenderBytes is a helper function that renders the template as a byte slice.
func (t *Template) RenderBytes(context ...interface{}) ([]byte, error) {
	var b *bytes.Buffer
	err := t.Render(b, context...)
	return b.Bytes(), err
}

// Parse wraps the creation of a new template and parsing from r in one go.
func Parse(r io.Reader) (*Template, error) {
	t := New()
	err := t.Parse(r)
	return t, err
}

// Render wraps the parsing and rendering into a single function.
func Render(r io.Reader, w io.Writer, context ...interface{}) error {
	t, err := Parse(r)
	if err != nil {
		return err
	}
	return t.Render(w, context...)
}
