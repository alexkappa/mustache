// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
		[]token{},
	},
}

func TestLexer(t *testing.T) {
	for _, test := range lexerTests {
		t.Logf("test: %q\n", test.name)
		lex := newLexer(test.template, test.leftDelim, test.rightDelim)
		// i := 0
		for {
			token := lex.token() // get the next token from the lexer

			// if token.typ != test.tokens[i].typ {
			// 	t.Errorf("parse error on test %q\n", test.name)
			// 	t.Errorf("expected token %q but instead got %q\n", token.typ, test.tokens[i].typ)
			// }
			// i++

			t.Logf("%-15s: %q\n", token.typ, token.val)

			// the lexer emitted an EOF or error
			if token.typ <= tokenEOF {
				break
			}
		}
	}
}
