// Copyright (c) 2014 Alex Kalyvitis

package mustache

import "testing"

var lexerTests = []struct {
	name       string
	template   string
	leftDelim  string
	rightDelim string
	tokens     []token
}{
	{
		"simple",
		"foo {{{bar}}} baz {{! this is ignored }}",
		"{{",
		"}}",
		[]token{
			{typ: tokenText},
			{typ: tokenLeftDelim},
			{typ: tokenRawStart},
			{typ: tokenIdentifier},
			{typ: tokenRawEnd},
			{typ: tokenRightDelim},
			{typ: tokenText},
			{typ: tokenLeftDelim},
			{typ: tokenComment},
			{typ: tokenText},
			{typ: tokenRightDelim},
			{typ: tokenEOF},
		},
	},
	{
		"set delimiters",
		"foo {{bar}} baz {{=| |=}} |foo| |={{ }}=| {{bar}}",
		"{{",
		"}}",
		[]token{
			{typ: tokenText},
			{typ: tokenLeftDelim},
			{typ: tokenIdentifier},
			{typ: tokenRightDelim},
			{typ: tokenText},
			{typ: tokenText},
			{typ: tokenLeftDelim},
			{typ: tokenIdentifier},
			{typ: tokenRightDelim},
			{typ: tokenText},
			{typ: tokenText},
			{typ: tokenLeftDelim},
			{typ: tokenIdentifier},
			{typ: tokenRightDelim},
			{typ: tokenEOF},
		},
	},
}

func TestLexer(t *testing.T) {
	for _, test := range lexerTests {
		lexer := newLexer(test.template, test.leftDelim, test.rightDelim)
		for token, i := lexer.token(), 0; token.typ <= tokenEOF; token, i = lexer.token(), i+1 {
			if token.typ != test.tokens[i].typ {
				t.Errorf("unexpected token %q, expected %q", token.typ, test.tokens[i].typ)
			}
		}
	}
}
