package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseConfig(t *testing.T) {
	config := `# Test Config
[[pin]]
GPIO=1
Name="waffles"
DetectAC=true

[[pin]]
GPIO=2
Name="pancakes"
Labels={ Foo="Bar" }
`
	want := Config{
		Pin: []pinConfig{
			{GPIO: 1, Name: "waffles", DetectAC: true},
			{GPIO: 2, Name: "pancakes", Labels: map[string]string{"Foo": "Bar"}},
		},
	}

	got, err := parseConfig(config)
	if err != nil {
		t.Errorf("parsing valid config: %s", err)
	}
	if !cmp.Equal(want, got) {
		t.Errorf("want: -, got: +:\n%s", cmp.Diff(want, got))
	}
}

// TestInvalidConfig should not go to far testing toml parsing, it just ensures
// that an invalid config does return an error.
func TestInvalidConfig(t *testing.T) {
	config := ` [testpin1] name= `
	_, err := parseConfig(config)
	if err == nil {
		t.Errorf("invalid config does not return an error")
	}
}
