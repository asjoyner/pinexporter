package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"

	"github.com/BurntSushi/toml"
	"github.com/asjoyner/pinexporter/acpin"
)

var (
	configPath = flag.String("config", "config.toml", "Path to the configuration file.")
)

// Pin provides a uniform interface for AC and DC GPIO pins
type Pin interface {
	Name() string
	Read() gpio.Level
}

// Config represents the pinexporter TOML configuration format
type Config struct {
	Pin []pinConfig
}

// pinConfig represents the configuration of each individual pin
type pinConfig struct {
	GPIO     int
	Name     string
	DetectAC bool
	Labels   map[string]string
}

func parseConfig(tomlString string) (Config, error) {
	var conf Config
	if _, err := toml.Decode(tomlString, &conf); err != nil {
		return Config{}, err
	}
	return conf, nil
}

func configurePins(conf Config) ([]Pin, error) {
	if len(conf.Pin) == 0 {
		return nil, fmt.Errorf("no pins specified")
	}
	pins := make([]Pin, len(conf.Pin))
	for _, p := range conf.Pin {
		// TODO: build a DC pin receiver
		/*
			newPin := gpioreg.ByName
			if p.DetectAC {
				newPin = acpin.New
			}
		*/
		pin, err := acpin.New(fmt.Sprintf("%d", p.GPIO))
		if err != nil {
			return nil, err
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

func main() {
	if _, err := host.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tomlBytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	conf, err := parseConfig(string(tomlBytes))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	pins, err := configurePins(conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	for {
		time.Sleep(time.Second)
		for _, pin := range pins {
			fmt.Printf("Pin %s is: %s\n", pin.Name(), pin.Read())
		}
		fmt.Println()
	}
}
