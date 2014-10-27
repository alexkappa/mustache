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

type Spec struct {
	Overview string `json:"overview"`
	Tests    []struct {
		Name     string      `json:"name"`
		Data     interface{} `json:"data"`
		Expected string      `json:"expected"`
		Template string      `json:"template"`
		Desc     string      `json:"desc"`
	} `json:"tests"`
}

var specs = make(map[string]Spec)

func init() {
	files, err := ioutil.ReadDir("./spec/specs")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.Contains(file.Name(), ".json") {
			var spec Spec
			f, err := os.Open(filepath.Join("./spec/specs", file.Name()))
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

func spec(t *testing.T, s Spec) {
	for _, test := range s.Tests {
		buf := &bytes.Buffer{}
		buf.WriteString(fmt.Sprintf("%q\n", test.Desc))
		buf.WriteString(fmt.Sprintf("Template: %q\n", test.Template))
		buf.WriteString(fmt.Sprintf("Expected: %q\n", test.Expected))
		template := New()
		err := template.ParseString(test.Template)
		if err != nil {
			t.Fatalf("%sParse failed on test %q: %q", buf, test.Name, err)
		}
		output, err := template.RenderString(test.Data)
		if err != nil {
			t.Fatalf("%sRender failed on test %q: %q", buf, test.Name, err)
		}
		buf.WriteString(fmt.Sprintf("Have    : %q\n", output))
		if output != test.Expected {
			t.Errorf("%sExpected %q got %q\n", buf, test.Expected, output)
		}
	}
}

func TestSpecComments(t *testing.T) {
	spec(t, specs["comments"])
}

func TestSpecDelimiters(t *testing.T) {
	spec(t, specs["delimiters"])
}

func TestSpecInterpolation(t *testing.T) {
	spec(t, specs["interpolation"])
}

func TestSpecInverted(t *testing.T) {
	spec(t, specs["inverted"])
}

func TestSpecPartials(t *testing.T) {
	spec(t, specs["partials"])
}

func TestSpecSections(t *testing.T) {
	spec(t, specs["sections"])
}

func TestSpecLambdas(t *testing.T) {
	t.Skip("first fix the lookup func")
	spec(t, specs["~lambdas"])
}
