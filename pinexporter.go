package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/BurntSushi/toml"
	"github.com/asjoyner/pinexporter/acpin"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configPath = flag.String("config", "config.toml", "Path to the configuration file.")
	addr       = flag.String(
		"addr",
		":4746",
		"The hostname and port to bind to and serve /metrics.",
	)
	pinExp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pin",
			Help: "Status of the GPIO pin identified by the label.",
		},
		[]string{
			// name is the name of the pin from the config.toml file
			"name"},
	)
)

// Pin provides a limited uniform interface for AC and DC GPIO pins
type Pin interface {
	Name() string
	In(gpio.Pull, gpio.Edge) error
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
	pins := make([]Pin, 0)
	for _, p := range conf.Pin {
		var pin Pin
		if p.DetectAC {
			pin = acpin.ByName(fmt.Sprintf("%d", p.GPIO))
		} else {
			pin = gpioreg.ByName(fmt.Sprintf("%d", p.GPIO))
		}
		if pin == nil {
			return nil, fmt.Errorf("no such pin: %d", p.GPIO)
		}
		if err := pin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			return nil, fmt.Errorf("failed to initialize pin %d: %s", p.GPIO, err)
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

func main() {
	prometheus.MustRegister(pinExp) // export the Prometheus variable

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
		os.Exit(4)
	}

	http.Handle("/metrics", promhttp.Handler())
	go log.Fatal(http.ListenAndServe(*addr, nil))

	for {
		time.Sleep(time.Second)
		for _, pin := range pins {
			glog.V(1).Infof("Pin %s is: %s\n", pin.Name(), pin.Read())
			if pin.Read() == gpio.High {
				pinExp.With(prometheus.Labels{"name": pin.Name()}).Set(1)
			} else {
				pinExp.With(prometheus.Labels{"name": pin.Name()}).Set(0)
			}
		}
	}
}
