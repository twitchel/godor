package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stianeikeland/go-rpio"
)

type ping struct {
	isOpen bool
	time   time.Time
}

func main() {
	c := prepareConfig()
	log.Printf("configuration: %+v", c)

	client, err := connectMqtt(c.mqtt.server)

	if err != nil {
		panic(err)
	}

	pings := make(chan ping)

	go func(c *config, p chan ping) {
		// every 100 ms check if the pin is high or low and report this to then pings channel
		ticker := time.NewTicker(c.sensor.checkRate)

		var pin rpio.Pin
		if !c.emulate {
			pin = getPin(c.sensor.pin)
			pin.Input()
		}

		for _ = range ticker.C {
			if c.emulate {
				p <- emulatePing()
			} else {
				p <- gpioPing(pin)

			}
		}
	}(c, pings)

	recheck := false
	recheckValue := true
	var previousPing ping
	for p := range pings {
		if p.isOpen != previousPing.isOpen {
			recheckValue = p.isOpen
			recheck = true
		} else if recheck == true {
			if p.isOpen == recheckValue {

				payload := fmt.Sprintf("{\"isOpen\":%v}", p.isOpen)
				go client.Publish(c.mqtt.topic, byte(0), false, payload)

				log.Printf(payload)
			}

			recheck = false
		}

		previousPing = p
	}

}

func gpioPing(pin rpio.Pin) ping {
	pinState := pin.Read()

	var isOpen bool
	if pinState == rpio.Low {
		isOpen = false
	} else {
		isOpen = true
	}

	return ping{
		isOpen: isOpen,
		time:   time.Now(),
	}
}

func emulatePing() ping {
	_, err := ioutil.ReadFile("./open")

	var isOpen bool
	if err != nil {
		isOpen = false
	} else {
		isOpen = true
	}

	return ping{
		isOpen: isOpen,
		time:   time.Now(),
	}
}

func getPin(pin int) rpio.Pin {
	thePin := rpio.Pin(pin)

	if err := rpio.Open(); err != nil {
		log.Printf("unable to open pin %d: %v", pin, err)
		os.Exit(1)
	}

	return thePin
}

func connectMqtt(server string) (MQTT.Client, error) {
	log.Println("mqtt: preparing connection")
	connOpts := MQTT.NewClientOptions().AddBroker(server).SetCleanSession(true)
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	connOpts.SetTLSConfig(tlsConfig)

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return client, token.Error()
	}
	log.Printf("mqtt: connected to %s\n", server)
	return client, nil
}
