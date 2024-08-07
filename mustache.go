// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
)

type ErrorSlice []error

func (es ErrorSlice) Error() string {
	b := strings.Builder{}
	b.WriteRune('[')
	first := true
	for _, e := range es {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(e.Error())
	}
	b.WriteRune(']')

	return b.String()
}

func maybeAddNode(seenNodes map[node]struct{}, newNode node) bool {
	if _, ok := seenNodes[newNode]; !ok {
		seenNodes[newNode] = struct{}{}
		return true
	}
	return false
}

func nodeAlreadyVisitedError(node string) error {
	return fmt.Errorf("the node %s has been already visited in the mustache node graph", node)
}

// The node type is the base type that represents a node in the parse tree.
type node interface {
	// The render function should be defined by any type wishing to satisfy the
	// node interface. Implementations should be able to render itself to the
	// w Writer with c given as context.
	render(t *Template, w *writer, seenNodes map[node]struct{}, c ...interface{}) error
}

// The textNode type represents a part of the template that is made up solely of
// text. It's an alias to string and it ignores c when rendering.
type textNode string

func (n textNode) render(t *Template, w *writer, _ map[node]struct{}, c ...interface{}) error {
	for _, r := range n {
		if !whitespace(r) {
			w.text()
		}
		err := w.write(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n textNode) String() string {
	return fmt.Sprintf("[text: %q]", string(n))
}

type escapeType int

const (
	noEscape escapeType = iota
	htmlEscape
	jsonEscape
)

func (e escapeType) String() string {
	switch e {
	case noEscape:
		return "noEscape"
	case htmlEscape:
		return "htmlEscape"
	case jsonEscape:
		return "jsonEscape"
	default:
		return "invalidEscape"
	}
}

// The varNode type represents a part of the template that needs to be replaced
// by a variable that exists within c.
type varNode struct {
	name   string
	escape escapeType
}

func (n *varNode) render(t *Template, w *writer, _ map[node]struct{}, c ...interface{}) error {
	w.text()
	v, _ := lookup(n.name, c...)
	// If the value is present but 'falsy', such as a false bool, or a zero int,
	// we still want to render that value.
	if v != nil {
		print(w, v, n.escape)
		return nil
	}
	return fmt.Errorf("failed to lookup %s", n.name)
}

func (n *varNode) String() string {
	return fmt.Sprintf("[var: %q escaped: %s]", n.name, n.escape.String())
}

// The sectionNode type is a complex node which recursively renders its child
// elements while passing along its context along with the global context.
type sectionNode struct {
	name     string
	inverted bool
	elems    []node
}

func (n *sectionNode) render(t *Template, w *writer, seenNodes map[node]struct{}, c ...interface{}) error {
	if !maybeAddNode(seenNodes, n) {
		return nodeAlreadyVisitedError(n.String())
	}
	defer delete(seenNodes, n)

	w.tag()
	defer w.tag()

	errs := ErrorSlice{}

	elemFn := func(v ...interface{}) {
		for _, elem := range n.elems {
			err := elem.render(t, w, seenNodes, append(v, c...)...)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	v, ok := lookup(n.name, c...)
	if ok != n.inverted {
		r := reflect.ValueOf(v)
		switch r.Kind() {
		case reflect.Slice, reflect.Array:
			if r.Len() > 0 {
				for i := 0; i < r.Len(); i++ {
					elemFn(r.Index(i).Interface())
				}
			} else {
				elemFn(v)
			}
		default:
			elemFn(v)
		}
	}
	if len(errs) != 0 {
		if !t.silentMiss {
			return errs
		}
	}
	return nil
}

// The testNode type is a complex node which recursively renders its child
// elements while passing along its context along with the global context.
type testNode struct {
	testIdent string
	testVal   string
	elems     []node
}

func (n *testNode) render(t *Template, w *writer, seenNodes map[node]struct{}, c ...interface{}) error {
	if !maybeAddNode(seenNodes, n) {
		return nodeAlreadyVisitedError(n.String())
	}
	defer delete(seenNodes, n)

	w.tag()
	defer w.tag()
	errs := ErrorSlice{}
	v, _ := lookup(n.testIdent, c...)
	if v != nil {
		vs := strings.Builder{}
		print(&vs, v, noEscape)
		if vs.String() == n.testVal {
			for _, elem := range n.elems {
				err := elem.render(t, w, seenNodes, c...)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	if len(errs) != 0 {
		if !t.silentMiss {
			return errs
		}
	}
	return nil
}

func (n *testNode) String() string {
	return fmt.Sprintf("[test: %q val: %s elems: %s]", n.testIdent, n.testVal, n.elems)
}

func (n *sectionNode) String() string {
	return fmt.Sprintf("[section: %q inv: %t elems: %s]", n.name, n.inverted, n.elems)
}

// The commentNode type is a part of the template which gets ignored. Perhaps it
// can be optionally enabled to print comments.
type commentNode string

func (n commentNode) render(t *Template, w *writer, _ map[node]struct{}, c ...interface{}) error {
	w.tag()
	return nil
}

func (n commentNode) String() string {
	return fmt.Sprintf("[comment: %q]", string(n))
}

// The partialNode type represents a named partial template.
type partialNode struct {
	name string
}

func (p *partialNode) render(t *Template, w *writer, seenNodes map[node]struct{}, c ...interface{}) error {
	if !maybeAddNode(seenNodes, p) {
		return nodeAlreadyVisitedError(p.String())
	}
	defer delete(seenNodes, p)

	w.tag()
	if template, ok := t.partials[p.name]; ok {
		template.partials = t.partials
		err := template.render(w, seenNodes, c...)
		if err != nil {
			if !t.silentMiss {
				return err
			}
		}
	}

	return nil
}

func (p *partialNode) String() string {
	return fmt.Sprintf("[partial: %s]", p.name)
}

type delimNode string

func (n delimNode) String() string {
	return "[delim]"
}

func (n delimNode) render(t *Template, w *writer, _ map[node]struct{}, c ...interface{}) error {
	w.tag()
	return nil
}

// The print function is able to format the interface v and write it to w using
// the best possible formatting flags.
func print(w io.Writer, v interface{}, needEscape escapeType) {
	var output string
	if s, ok := v.(fmt.Stringer); ok {
		output = s.String()
	} else {
		switch v.(type) {
		case string:
			output = fmt.Sprintf("%s", v)
		case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			output = fmt.Sprintf("%d", v)
		case float32, float64:
			output = fmt.Sprintf("%g", v)
		default:
			obj, _ := json.Marshal(v)
			output = string(obj)
		}
	}

	if needEscape == htmlEscape {
		output = escapeHtml(output)
	} else if needEscape == jsonEscape {
		output = escapeJson(output)
	}
	fmt.Fprint(w, output)
}

// The escape function replicates the text/template.HTMLEscapeString but keeps
// "&apos;" and "&quot;" for compatibility with the mustache spec.
func escapeHtml(s string) string {
	if !strings.ContainsAny(s, `'"&<>`) {
		return s
	}
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func escapeJson(s string) string {
	b := new(strings.Builder)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(s)
	if err != nil {
		panic(err)
	}
	// Skip 1 character at the beginning for the quote
	// Skip 2 characters at the end, 1 quote and 1 newline
	// which is inserted automatically by encoder.
	return b.String()[1 : b.Len()-2]
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

// Errors enables missing variable errors. This option is deprecated. Please
// use SilentMiss instead.
func Errors() Option {
	return func(t *Template) {
		t.silentMiss = false
	}
}

// SilentMiss sets the silent miss behavior of variable lookups when rendering.
// If true, missed lookups will not produce any errors. Otherwise a missed
// variable lookup will stop the rendering and return an error.
func SilentMiss(silent bool) Option {
	return func(t *Template) {
		t.silentMiss = silent
	}
}

// Default is this, when text is inserted it will be escaped
// HTML style as is default for mustache.
func HtmlEscape() Option {
	return func(t *Template) {
		t.escape = htmlEscape
	}
}

// If you specify this option, then text inserted into the template
// will be escaped using JSON escaping rules (with SetHTMLEscaping(false))
func JsonEscape() Option {
	return func(t *Template) {
		t.escape = jsonEscape
	}
}

// If you specify this option, then this Template will support
// {{#test_value ident value}} sections
func TestValueSection() Option {
	return func(t *Template) {
		t.testValueSection = true
	}
}

// The Template type represents a template and its components.
type Template struct {
	name             string
	elems            []node
	partials         map[string]*Template
	startDelim       string
	endDelim         string
	silentMiss       bool
	testValueSection bool
	escape           escapeType
}

// New returns a new Template instance.
func New(options ...Option) *Template {
	t := &Template{
		elems:            make([]node, 0),
		partials:         make(map[string]*Template),
		startDelim:       "{{",
		endDelim:         "}}",
		silentMiss:       true,
		testValueSection: false,
		escape:           htmlEscape,
	}
	t.Option(options...)
	return t
}

// Option applies options to the currrent template t.
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
	l := newLexer(string(b), t.startDelim, t.endDelim, t.testValueSection)
	p := newParser(l, t.escape)
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

func (t *Template) render(w *writer, seenNodes map[node]struct{}, context ...interface{}) error {
	for _, elem := range t.elems {
		err := elem.render(t, w, seenNodes, context...)
		if err != nil {
			if !t.silentMiss {
				return err
			}
		}
	}
	return w.flush()
}

// Render walks through the template's parse tree and writes the output to w
// replacing the values found in context.
func (t *Template) Render(w io.Writer, context ...interface{}) error {
	return t.render(newWriter(w), make(map[node]struct{}), context...)
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
