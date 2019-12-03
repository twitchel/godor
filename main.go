package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
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

		var pin gpio.PinIO
		if !c.emulate {
			host.Init()
			pin = getPin(c.sensor.pin)
			//pin.WaitForEdge(-1)
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
				go client.Publish(c.sensor.topic, byte(0), false, payload)

				log.Printf(payload)
			}

			recheck = false
		}

		previousPing = p
	}

}

func gpioPing(pin gpio.PinIO) ping {
	pinState := pin.Read()

	var isOpen bool
	if pinState == gpio.Low {
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

func getPin(pin string) gpio.PinIO {
	thePin := gpioreg.ByName(pin)

	if err := thePin.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		log.Printf("unable to open pin %s: %v", pin, err)
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
