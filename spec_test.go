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
	"strings"
	"testing"
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

func testSpec(t *testing.T, s Spec) {
	for _, test := range s.Tests {
		buf := &bytes.Buffer{}
		buf.WriteString(fmt.Sprintf("%q\n", test.Desc))
		buf.WriteString(fmt.Sprintf("Name    : %q\n", test.Name))
		buf.WriteString(fmt.Sprintf("Template: %q\n", test.Template))
		buf.WriteString(fmt.Sprintf("Template: %v\n", test.Data))
		template := New()
		err := template.ParseString(test.Template)
		if err != nil {
			t.Fatalf("%sParse failed on test %q: %q", buf, test.Name, err)
		}
		for n, s := range test.Partials {
			buf.WriteString(fmt.Sprintf("Partial : %s> %q\n", n, s))
			p := New(Name(n))
			if err := p.ParseString(s); err != nil {
				t.Fatalf("%sParse failed on test %q partial %s: %q", buf, test.Name, err, n)
			}
			template.Option(Partial(p))
		}
		output, err := template.RenderString(test.Data)
		if err != nil {
			t.Fatalf("%sRender failed on test %q: %q", buf, test.Name, err)
		}
		buf.WriteString(fmt.Sprintf("Tree    : %#v\n", template.elems))
		buf.WriteString(fmt.Sprintf("Expected: %q\n", test.Expected))
		buf.WriteString(fmt.Sprintf("Have    : %q\n", output))
		if output != test.Expected {
			t.Error(buf.String())
		}
	}
}

func TestSpecComments(t *testing.T) {
	testSpec(t, specs["comments"])
}

func TestSpecDelimiters(t *testing.T) {
	testSpec(t, specs["delimiters"])
}

func TestSpecInterpolation(t *testing.T) {
	// testSpec(t, specs["interpolation"])
}

func TestSpecInverted(t *testing.T) {
	// testSpec(t, specs["inverted"])
}

func TestSpecPartials(t *testing.T) {
	testSpec(t, specs["partials"])
}

func TestSpecSections(t *testing.T) {
	// testSpec(t, specs["sections"])
}

func TestSpecLambdas(t *testing.T) {
	t.Skip("It's not possible to evaluate functions in Go at runtime. Revisit this test soon")
	testSpec(t, specs["~lambdas"])
}
