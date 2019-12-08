package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/host"
)

var (
	quietEdge = make(chan gpio.Level)
	quietPin  = gpiotest.Pin{N: "quietpin", L: gpio.Low, EdgesChan: quietEdge}
)

func init() {
	gpioreg.Register(&quietPin)
	gpioreg.RegisterAlias("424242", "quietpin")
	gpioreg.RegisterAlias("424243", "quietpin")
}

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
	if _, err := host.Init(); err != nil {
		t.Fatal(err)
	}
	config := ` [testpin1] name= `
	_, err := parseConfig(config)
	if err == nil {
		t.Errorf("invalid config does not return an error")
	}
}

func TestConfigurePins(t *testing.T) {
	conf := Config{
		Pin: []pinConfig{
			{GPIO: 424242, Name: "quietpin"},
			{GPIO: 424243, Name: "quietpin", DetectAC: true},
		},
	}
	pins, err := configurePins(conf)
	if err != nil {
		t.Errorf("configuring test pins: %s", err)
	}
	for i, pin := range pins {
		if pin == nil {
			t.Errorf("pin %d is nil!?", i)
			continue
		}
		t.Logf("testing read from pin %s", pin.Name())
		got := pin.Read()
		if got != gpio.Low {
			t.Errorf("pin %d, want: gpio.High, got: %s\n", i, got)
		}
	}
}
