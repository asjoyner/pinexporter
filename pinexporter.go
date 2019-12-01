package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

var (
	lastEdge time.Time
)

func main() {
	if _, err := host.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	p := gpioreg.ByName("5")
	if p == nil {
		fmt.Fprintln(os.Stderr, "Failed to find pin 5!")
		os.Exit(1)
	}
	if err := p.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to initialize pin 5: ", err)
		os.Exit(1)
	}

	go func() {
		for {
			glog.V(1).Infof("waiting for edge...")
			if p.WaitForEdge(time.Second) {
				glog.V(1).Infof("Edge!")
				lastEdge = time.Now()
				time.Sleep(250 * time.Millisecond)
				continue
			}
			glog.V(1).Infof("there was no edge.")
		}
	}()

	for {
		time.Sleep(time.Second)
		if lastEdge.Before(time.Now().Add(-3 * time.Second)) {
			fmt.Println("Pin 5 is: Off")
			continue
		}
		fmt.Println("Pin 5 is: On")
	}
}
