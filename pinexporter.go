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
	helpText = "Status of the GPIO pin identified by the label."
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
	GPIO     string
	Name     string
	DetectAC bool
	Labels   map[string]string
}

// configuredPin matches a pinConfig structs to the configured Pin to be monitored
type configuredPin struct {
	config pinConfig
	pin    Pin
	metric prometheus.Gauge
}

// parseConfig parses a TOML string into a Config struct
func parseConfig(tomlString string) (Config, error) {
	var conf Config
	if _, err := toml.Decode(tomlString, &conf); err != nil {
		return Config{}, err
	}
	return conf, nil
}

// configurePins accepts a Config struct, initalizes a set of Pins, and stores the pin objects in the
func configurePins(conf Config) ([]configuredPin, error) {
	if len(conf.Pin) == 0 {
		return nil, fmt.Errorf("no pins specified")
	}
	pins := make([]configuredPin, 0, len(conf.Pin))
	for _, p := range conf.Pin {
		// Initialize the GPIO pin
		var pin Pin
		if p.DetectAC {
			pin = acpin.ByName(p.GPIO)
		} else {
			pin = gpioreg.ByName(p.GPIO)
		}
		if pin == nil {
			return nil, fmt.Errorf("no GPIO pin named: %s", p.GPIO)
		}
		if err := pin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			return nil, fmt.Errorf("failed to initialize pin %s: %s", p.GPIO, err)
		}

		// Configure the Prometheus Gauge metric
		labelNames := []string{"gpio"}
		labels := map[string]string{"gpio": p.GPIO}
		for name, value := range p.Labels {
			labelNames = append(labelNames, name)
			labels[name] = value
		}
		g := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: p.Name, Help: helpText},
			labelNames,
		)
		if err := prometheus.Register(g); err != nil {
			return nil, fmt.Errorf("failed to register metric for pin %s: %s", p.GPIO, err)
		}
		metric, err := g.GetMetricWith(labels)
		if err != nil {
			return nil, fmt.Errorf("failed to set metrics for pin %s: %s", p.GPIO, err)
		}
		pins = append(pins, configuredPin{config: p, pin: pin, metric: metric})
	}
	return pins, nil
}

func main() {
	flag.Parse()
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
	go func() { log.Fatal(http.ListenAndServe(*addr, nil)) }()

	for {
		time.Sleep(time.Second)
		for _, p := range pins {
			value := p.pin.Read()
			glog.V(1).Infof("Pin %s [%s] is: %s\n", p.config.Name, p.config.GPIO, value)
			if value == gpio.High {
				p.metric.Set(1)
			} else {
				p.metric.Set(0)
			}
		}
	}
}
