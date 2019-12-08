// Package acpin presents the gpio.PinIn interface to handle consistent
// software detection of AC voltage on a named GPIO pin.
package acpin

import (
	"sync"
	"time"

	"github.com/golang/glog"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

// PinIn implements gpio.PinIn to represent voltage on an AC GPIO pin
type PinIn struct {
	gpiopin  gpio.PinIO
	lastEdge time.Time
	mu       sync.RWMutex // protects lastEdge
	halt     bool
}

// Name returns the name of the pin.
func (p *PinIn) Name() string {
	return p.gpiopin.Name()
}

// ByName returns a PinIn associated with GPIO pin described by name.  It will
// return gpio.High when it has detected AC voltage.
//
// This function duplicates gpioreg.ByName signature to simplify code that
// interacts with both types of pins.
func ByName(name string) *PinIn {
	pin := PinIn{N: name}
	pin.gpiopin = gpioreg.ByName(name)
	if pin.gpiopin == nil {
		return nil
	}
	return &pin
}

// In configures the pin for detecting AC voltage.
// It always sets edge detection only to RisingEdge, to reduce interrupt rate.
func (p *PinIn) In(pull gpio.Pull, _ gpio.Edge) error {
	if err := p.gpiopin.In(pull, gpio.RisingEdge); err != nil {
		return err
	}
	go p.watchPin()
	return nil
}

// Read returns the Level of the pin.  It is High if the pin has seen an AC
// cycle in the last 3 seconds.
func (p *PinIn) Read() gpio.Level {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.lastEdge.Before(time.Now().Add(-3 * time.Second)) {
		return gpio.Low
	}
	return gpio.High
}

// Halt stops the gorouting watching the pin for AC voltage.
func (p *PinIn) Halt() error {
	p.halt = true
	return nil
}

func (p *PinIn) watchPin() {
	for {
		if p.halt {
			return
		}
		glog.V(1).Infof("waiting for edge on pin %s...", p.gpiopin.Name())
		if p.gpiopin.WaitForEdge(time.Second) {
			glog.V(1).Infof("Found Edge on pin %s!", p.gpiopin.Name)
			p.mu.Lock()
			p.lastEdge = time.Now()
			p.mu.Unlock()
			time.Sleep(250 * time.Millisecond)
			continue
		}
		glog.V(1).Infof("timed out with no edge on pin %s", p.gpiopin.Name())
	}
}
