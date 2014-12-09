// Copyright (c) 2014 Alex Kalyvitis

package mustache

import "testing"

func TestLexer(t *testing.T) {
	for _, test := range []struct {
		lexer    *lexer
		expected []token
	}{
		{
			newLexer("foo {{{bar}}} baz {{! this is ignored }}", "{{", "}}"),
			[]token{
				{typ: tokenText, val: "foo "},
				{typ: tokenLeftDelim, val: "{{"},
				{typ: tokenRawStart, val: "{"},
				{typ: tokenIdentifier, val: "bar"},
				{typ: tokenRawEnd, val: "}"},
				{typ: tokenRightDelim, val: "}}"},
				{typ: tokenText, val: " baz "},
				{typ: tokenLeftDelim, val: "{{"},
				{typ: tokenComment, val: "!"},
				{typ: tokenText, val: " this is ignored "},
				{typ: tokenRightDelim, val: "}}"},
				{typ: tokenEOF},
			},
		},
		{
			newLexer("foo {{bar}} baz {{=| |=}} |foo| |={{ }}=| {{bar}}", "{{", "}}"),
			[]token{
				{typ: tokenText, val: "foo "},
				{typ: tokenLeftDelim, val: "{{"},
				{typ: tokenIdentifier, val: "bar"},
				{typ: tokenRightDelim, val: "}}"},
				{typ: tokenText, val: " baz "},
				{typ: tokenText, val: " "},
				{typ: tokenLeftDelim, val: "|"},
				{typ: tokenIdentifier, val: "foo"},
				{typ: tokenRightDelim, val: "|"},
				{typ: tokenText, val: " "},
				{typ: tokenText, val: " "},
				{typ: tokenLeftDelim, val: "{{"},
				{typ: tokenIdentifier, val: "bar"},
				{typ: tokenRightDelim, val: "}}"},
				{typ: tokenEOF},
			},
		},
	} {
		var (
			lexer = test.lexer
			token = lexer.token()
			i     = 0
		)
		for token.typ > tokenEOF {
			if token.typ != test.expected[i].typ {
				t.Errorf("unexpected token %q, expected %q", token.typ, test.expected[i].typ)
			}
			if token.val != test.expected[i].val {
				t.Errorf("unexpected value %q, expected %q", token.val, test.expected[i].val)
			}
			token = lexer.token()
			i++
		}
	}
}
