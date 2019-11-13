// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"text/tabwriter"
)

const specDir = "./spec/specs"

type Spec struct {
	Overview string `json:"overview"`
	Tests    []struct {
		Name     string            `json:"name"`
		Data     interface{}       `json:"data"`
		Expected string            `json:"expected"`
		Template string            `json:"template"`
		Partials map[string]string `json:"partials"`
		Desc     string            `json:"desc"`
	} `json:"tests"`
}

var specs = make(map[string]Spec)

func init() {
	files, err := ioutil.ReadDir(specDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.Contains(file.Name(), ".json") {
			var spec Spec
			f, err := os.Open(filepath.Join(specDir, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			err = json.NewDecoder(f).Decode(&spec)
			if err != nil {
				log.Fatal(err)
			}
			specs[strings.TrimSuffix(file.Name(), ".json")] = spec
		}
	}
}

type testWriter struct {
	b *bytes.Buffer
	w *tabwriter.Writer
}

func newTestWriter() *testWriter {
	t := &testWriter{}
	t.b = bytes.NewBuffer(nil)
	t.w = tabwriter.NewWriter(t.b, 0, 8, 0, ':', 0)
	return t
}

func (t *testWriter) Write(buf []byte) (n int, err error) { return t.w.Write(buf) }
func (t *testWriter) String() string                      { t.w.Flush(); return t.b.String() }
func (t *testWriter) Reset()                              { t.w.Flush(); t.b.Reset() }

func testSpecFunc(t *testing.T, s Spec) func(t *testing.T) {
	return func(t *testing.T) {

		t.Parallel() // run each spec test concurrently.

		// Create a new writer to write information about the test, and in
		// the event the test fails we print it to the screen.
		w := newTestWriter()

		for _, test := range s.Tests {

			fmt.Fprintf(w, "%s\n", test.Desc)
			fmt.Fprintf(w, "Name: %q\n", test.Name)
			fmt.Fprintf(w, "Template: %q\n", test.Template)
			fmt.Fprintf(w, "Data: %q\n", test.Data)

			// Handle and recover from panics.
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(w, "Error: %v\n", r)
					fmt.Fprintf(w, "Stack: %s\n", debug.Stack())
					t.Fatal(w.String())
				}
			}()

			defer w.Reset() // Reset writer after each test.

			// Parse the template and report errors.
			template := New()
			if err := template.ParseString(test.Template); err != nil {
				fmt.Fprintf(w, "Error: %s\n", err)
				t.Fatal(w.String())
			}

			// If partials were present in the spec test iterate and test each
			// one.
			for n, s := range test.Partials {
				fmt.Fprintf(w, "Partial : %s> %q\n", n, s)
				p := New(Name(n))
				if err := p.ParseString(s); err != nil {
					fmt.Fprintf(w, "Error: %s\n", err)
					t.Fatal(w.String())
				}
				template.Option(Partial(p))
			}

			output, err := template.RenderString(test.Data)
			if err != nil {
				fmt.Fprintf(w, "Error: %s\n", err)
				t.Fatal(w.String())
			}

			fmt.Fprintf(w, "Tree: %+v\n", template.elems)
			fmt.Fprintf(w, "Expected: %q\n", test.Expected)
			fmt.Fprintf(w, "Have: %q\n", output)

			if output != test.Expected {
				t.Fatal(w.String())
			}
		}
	}
}

func TestSpec(t *testing.T) {
	t.Run("Comments", testSpecFunc(t, specs["comments"]))
	t.Run("Delimiters", testSpecFunc(t, specs["delimiters"]))
	t.Run("Interpolation", testSpecFunc(t, specs["interpolation"]))
	t.Run("Inverted", testSpecFunc(t, specs["inverted"]))
	t.Run("Partials", func(t *testing.T) {
		t.Skip("skip partials as they don't conform fully to the standard")
		testSpecFunc(t, specs["partials"])(t)
	})
	t.Run("Sections", testSpecFunc(t, specs["sections"]))
}
