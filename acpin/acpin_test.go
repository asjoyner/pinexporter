package acpin

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

var (
	quietEdge = make(chan gpio.Level)
	quietPin  = gpiotest.Pin{N: "quietpin", L: gpio.Low, EdgesChan: quietEdge}

	liveEdge  = make(chan gpio.Level)
	liveLevel gpio.Level
	wigglyPin = gpiotest.Pin{N: "wigglypin", L: liveLevel, EdgesChan: liveEdge}
)

func init() {
	gpioreg.Register(&quietPin)
	gpioreg.Register(&wigglyPin)
	go wigglePin()
}

func TestName(t *testing.T) {
	p := ByName("quietpin")
	if p == nil {
		t.Errorf("no such pin: quietpin")
	}
	name := p.Name()
	if name != "quietpin" {
		t.Errorf("want: quietpin, got: %q", name)
	}
}

func TestQuietPin(t *testing.T) {
	p := ByName("quietpin")
	if p == nil {
		t.Errorf("no such pin: quietpin")
	}
	if err := p.In(gpio.PullDown, gpio.BothEdges); err != nil {
		t.Errorf("failed to initialize: %s", err)
	}
	if res := p.Read(); res != gpio.Low {
		t.Errorf("quiet pin is not low: want: gpio.Low, got: %s", res)
	}
}

func TestQuietPinStaysLow(t *testing.T) {
	p := ByName("quietpin")
	if p == nil {
		t.Errorf("no such pin: quietpin")
	}
	if err := p.In(gpio.PullDown, gpio.BothEdges); err != nil {
		t.Errorf("failed to initialize: %s", err)
	}
	time.Sleep(time.Second)
	if res := p.Read(); res != gpio.Low {
		t.Errorf("quiet pin did not stay low: want: gpio.Low, got: %s", res)
	}
}

func TestNoisyPin(t *testing.T) {
	p := ByName("wigglypin")
	if p == nil {
		t.Errorf("no such pin: wigglypin")
	}
	if err := p.In(gpio.PullDown, gpio.BothEdges); err != nil {
		t.Errorf("failed to initialize: %s", err)
	}
	time.Sleep(time.Second)
	if res := p.Read(); res != gpio.High {
		t.Errorf("noisy pin is not high: want: gpio.High, got: %s", res)
	}
}

func TestTeardownParallel(t *testing.T) {
	if err := quietPin.Halt(); err != nil {
		t.Errorf("quietPin.Halt(): %s", err)
	}
	if err := wigglyPin.Halt(); err != nil {
		t.Errorf("wigglyPin.Halt(): %s", err)
	}

}

func wigglePin() {
	for i := 0; i < 10; i++ {
		liveEdge <- gpio.High
		liveLevel = gpio.High
		liveEdge <- gpio.Low
		liveLevel = gpio.Low
	}
}
