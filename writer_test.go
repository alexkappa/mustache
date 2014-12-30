package mustache

import (
	"bytes"
	"testing"
)

func TestWriter(t *testing.T) {
	for _, test := range []struct {
		text     bool
		tag      bool
		input    string
		expected string
	}{
		{true, true, "some text\n", "some text\n"},
		{false, true, "  {{#standalone}}\n here.", " here."},
		{false, false, "print this\n and this", "print this\n and this"},
	} {
		b := bytes.NewBuffer(nil)
		w := newWriter(b)
		w.hasText = test.text
		w.hasTag = test.tag
		for _, r := range test.input {
			err := w.write(r)
			if err != nil {
				t.Errorf("write error %q", err)
			}
		}
		w.flush()
		if b.String() != test.expected {
			t.Errorf("unexpected output %q, expected %q", b.String(), test.expected)
		}
	}
}
