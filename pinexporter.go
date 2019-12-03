package main

import (
	"fmt"
	"os"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"

	"github.com/asjoyner/pinexporter/acpin"
)

// Pin provides a uniform interface for AC and DC GPIO pins
type Pin interface {
	Read() gpio.Level
}

func main() {
	if _, err := host.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	pin, err := acpin.New("5")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for {
		time.Sleep(time.Second)
		fmt.Println("Pin 5 is: ", pin.Read())
	}
}
