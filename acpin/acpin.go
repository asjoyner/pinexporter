// Package acpin presents a similar interface to the gpio.PinIn package it
// includes, but handles consistent detection of AC voltage on the GPIO pin in
// software.
package acpin

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

// PinIn represents a AC GPIO pin
type PinIn struct {
	name     string
	gpiopin  gpio.PinIn
	lastEdge time.Time
	mu       sync.RWMutex // protects lastEdge
}

// New initializes and returns a PinIn struct which is initialized to detect
// voltage on the pin defined by "name".
func New(name string) (*PinIn, error) {
	pin := PinIn{name: name}
	if _, err := host.Init(); err != nil {
		return nil, err
	}
	pin.gpiopin = gpioreg.ByName(name)
	if pin.gpiopin == nil {
		return nil, fmt.Errorf("no such pin: %s", name)
	}
	if err := pin.gpiopin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		return nil, fmt.Errorf("failed to initialize pin %s: %s", name, err)
	}

	go pin.watchPin()
	return &pin, nil
}

// Read returns the Level of the pin.  It is High if the pin has seen an AC
// cycle in the last 3 seconds.
func (p *PinIn) Read() gpio.Level {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.lastEdge.Before(time.Now().Add(-3 * time.Second)) {
		return gpio.High
	}
	return gpio.Low
}

func (p *PinIn) watchPin() {
	for {
		// TODO: build a way to close this goroutine
		glog.V(1).Infof("waiting for edge on pin %s...", p.name)
		if p.gpiopin.WaitForEdge(time.Second) {
			glog.V(1).Infof("Found Edge on pin %s!", p.name)
			p.mu.Lock()
			p.lastEdge = time.Now()
			p.mu.Unlock()
			time.Sleep(250 * time.Millisecond)
			continue
		}
		glog.V(1).Infof("timed out with no edge on pin %s", p.name)
	}
}
