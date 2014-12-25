// Copyright (c) 2014 Alex Kalyvitis
// Portions Copyright (c) 2011 The Go Authors

package mustache

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// token represents a token or text string returned from the scanner.
type token struct {
	typ  tokenType
	val  string
	line int
	col  int
}

// String satisfies the fmt.Stringer interface making it easier to print tokens.
func (i token) String() string {
	return fmt.Sprintf("%s:%q", i.typ, i.val)
}

// tokenType identifies the type of lex tokens.
type tokenType int

const (
	tokenError tokenType = iota // error occurred; value is text of error
	tokenEOF
	tokenIdentifier     // alphanumeric identifier
	tokenLeftDelim      // {{ left action delimiter
	tokenRightDelim     // }} right action delimiter
	tokenText           // plain text
	tokenComment        // {{! this is a comment and is ignored}}
	tokenSectionStart   // {{#foo}} denotes a section start
	tokenSectionInverse // {{^foo}} denotes an inverse section start
	tokenSectionEnd     // {{/foo}} denotes the closing of a section
	tokenRawStart       // { denotes the beginning of an unencoded identifier
	tokenRawEnd         // } denotes the end of an unencoded identifier
	tokenRawAlt         // {{&foo}} is an alternative way to define raw tags
	tokenPartial        // {{>foo}} denotes a partial
	tokenSetDelim       // {{={% %}=}} sets delimiters to {% and %}
	tokenSetLeftDelim   // denotes a custom left delimiter
	tokenSetRightDelim  // denotes a custom right delimiter
)

// Make the types prettyprint.
var tokenName = map[tokenType]string{
	tokenError:          "t_error",
	tokenEOF:            "t_eof",
	tokenIdentifier:     "t_ident",
	tokenLeftDelim:      "t_left_delim",
	tokenRightDelim:     "t_right_delim",
	tokenText:           "t_text",
	tokenComment:        "t_comment",
	tokenSectionStart:   "t_section_start",
	tokenSectionInverse: "t_section_inverse",
	tokenSectionEnd:     "t_section_end",
	tokenRawStart:       "t_raw_start",
	tokenRawEnd:         "t_raw_end",
	tokenRawAlt:         "t_raw_alt",
	tokenPartial:        "t_partial",
	tokenSetDelim:       "t_set_delim",
	tokenSetLeftDelim:   "t_set_left_delim",
	tokenSetRightDelim:  "t_set_right_delim",
}

// String satisfies the fmt.Stringer interface making it easier to print tokens.
func (i tokenType) String() string {
	s := tokenName[i]
	if s == "" {
		return fmt.Sprintf("t_unknown_%d", int(i))
	}
	return s
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the
// next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string     // the name of the input; used only for error reports.
	input      string     // the string being scanned.
	leftDelim  string     // start of action.
	rightDelim string     // end of action.
	state      stateFn    // the next lexing function to enter.
	pos        int        // current position in the input.
	start      int        // start position of this token.
	width      int        // width of last rune read from input.
	tokens     chan token // channel of scanned tokens.
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an token back to the client.
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{
		t,
		l.input[l.start:l.pos],
		l.lineNum(),
		l.columnNum(),
	}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// lineNum reports which line we're on. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNum() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

// columnNum reports the character of the current line we're on.
func (l *lexer) columnNum() int {
	if lf := strings.LastIndex(l.input[:l.pos], "\n"); lf != -1 {
		return len(l.input[lf+1 : l.pos])
	}
	return len(l.input[:l.pos])
}

// error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.token.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{
		tokenError,
		fmt.Sprintf(format, args...),
		l.lineNum(),
		l.columnNum(),
	}
	return nil
}

// token returns the next token from the input.
func (l *lexer) token() token {
	for {
		select {
		case token := <-l.tokens:
			return token
		default:
			l.state = l.state(l)
		}
	}
	panic("not reached")
}

// state functions.

// stateText scans until an opening action delimiter, "{{".
func stateText(l *lexer) stateFn {
	for {
		// Lookahead for {{ which should switch to lexing an open tag instead of
		// regular text tokens.
		if strings.HasPrefix(l.input[l.pos:], l.leftDelim) {
			if l.pos > l.start {
				l.emit(tokenText)
			}
			return stateLeftDelim
		}
		// Produce a token and exit the loop if we have reached the end of file.
		if l.next() == eof {
			break
		}
	}
	// Emit whatever we gathered so far as text.
	if l.pos > l.start {
		l.emit(tokenText)
	}
	// Always end with EOF token. The parser will keep asking for tokens until
	// an tokenEOF or tokenError token are encountered.
	l.emit(tokenEOF)
	return nil
}

// stateLeftDelim scans the left delimiter, which is known to be present.
func stateLeftDelim(l *lexer) stateFn {
	l.pos += len(l.leftDelim)
	if l.input[l.pos] == '=' {
		// When the lexer encounters "{{=" it proceeds to the set delimiter
		// state which alters the left and right delimiters. This operation is
		// hidden from the parser and no tokens are emited.
		l.pos++
		return stateSetDelim
	}
	l.emit(tokenLeftDelim)
	return stateTag
}

// stateRightDelim scans the right delimiter, which is known to be present.
func stateRightDelim(l *lexer) stateFn {
	l.pos += len(l.rightDelim)
	l.emit(tokenRightDelim)
	return stateText
}

// stateTag scans the elements inside action delimiters.
func stateTag(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], "}"+l.rightDelim) {
		l.pos++
		l.emit(tokenRawEnd)
		return stateRightDelim
	}
	if strings.HasPrefix(l.input[l.pos:], l.rightDelim) {
		return stateRightDelim
	}
	switch r := l.next(); {
	case r == eof || r == '\n':
		return l.errorf("unclosed action")
	case whitespace(r):
		l.ignore()
	case r == '!':
		l.emit(tokenComment)
		return stateComment
	case r == '#':
		l.emit(tokenSectionStart)
	case r == '^':
		l.emit(tokenSectionInverse)
	case r == '/':
		l.emit(tokenSectionEnd)
	case r == '&':
		l.emit(tokenRawAlt)
	case r == '>':
		l.emit(tokenPartial)
	case r == '{':
		l.emit(tokenRawStart)
	case alphanum(r):
		l.backup()
		return stateIdent
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
	return stateTag
}

// stateIdent scans an alphanumeric or field.
func stateIdent(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case alphanum(r):
			// absorb.
		default:
			l.backup()
			l.emit(tokenIdentifier)
			break Loop
		}
	}
	return stateTag
}

// stateComment scans a comment. The left comment marker is known to be present.
func stateComment(l *lexer) stateFn {
	i := strings.Index(l.input[l.pos:], l.rightDelim)
	if i < 0 {
		return l.errorf("unclosed tag")
	}
	l.pos += i
	l.emit(tokenText)
	l.pos += len(l.rightDelim)
	l.emit(tokenRightDelim)
	return stateText
}

// stateSetDelim scans a set of set delimiter tags and replaces the lexers left
// and right delimiters to new values.
func stateSetDelim(l *lexer) stateFn {
	end := "=" + l.rightDelim
	i := strings.Index(l.input[l.pos:], end)
	if i < 0 {
		return l.errorf("unclosed tag")
	}
	delims := strings.Split(l.input[l.pos:l.pos+i], " ") // " | | "
	if len(delims) < 2 {
		l.errorf("set delimiters should be separated by a space")
	}
	delimFn := leftFn
	for _, delim := range delims {
		if delim != "" {
			if delimFn != nil {
				delimFn = delimFn(l, delim)
			}
		}
	}
	l.pos += i + len(end)
	l.ignore()
	return stateText
}

// delimFn is a self referencing function which helps with setting the right
// delimiter in the right order.
type delimFn func(l *lexer, s string) delimFn

// leftFn sets the left delimiter to s and returns a rightFn.
func leftFn(l *lexer, s string) delimFn {
	l.leftDelim = s
	return rightFn
}

// rightFn sets the right delimiter to s.
func rightFn(l *lexer, s string) delimFn {
	l.rightDelim = s
	return nil
}

// newLexer creates a new scanner for the input string.
func newLexer(input, left, right string) *lexer {
	l := &lexer{
		input:      input,
		leftDelim:  left,
		rightDelim: right,
		tokens:     make(chan token, 2),
	}
	l.state = stateText // initial state
	return l
}

// whitespace reports whether r is a space character.
func whitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

// alphanum reports whether r is an alphabetic, digit, or underscore.
func alphanum(r rune) bool {
	return r == '_' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
