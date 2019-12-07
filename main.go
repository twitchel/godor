package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
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
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill)

	c := prepareConfig()
	log.Printf("configuration: %+v", c)

	client, err := connectMqtt(c)

	if err != nil {
		panic(err)
	}

	go monitorButton(c.button, client)
	go monitorSensor(c.sensor, client, c.emulate)

	<-sigChan
}

func monitorButton(config *configButton, client MQTT.Client) {
	if token := client.Subscribe(config.topic, byte(0), func (client MQTT.Client, message MQTT.Message) {
		buttonHandler(config.pin, config.pressDuration)
	}); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func monitorSensor(c *configSensor, client MQTT.Client, emulate bool) {
	pings := make(chan ping)

	go func(c *configSensor, p chan ping, emulate bool) {
		// every 100 ms check if the pin is high or low and report this to then pings channel
		ticker := time.NewTicker(c.checkRate)

		var pin gpio.PinIO
		if !emulate {
			host.Init()
			pin = getSensorPin(c.pin)
		}

		for _ = range ticker.C {
			if emulate {
				p <- emulatePing()
			} else {
				p <- gpioPing(pin)

			}
		}
	}(c, pings, emulate)

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
				go client.Publish(c.topic, byte(0), false, payload)

				log.Printf(payload)
			}

			recheck = false
		}

		previousPing = p
	}
}

func buttonHandler(buttonPin string, duration time.Duration) {
	log.Printf("activating garage door")
	pin := getButtonPin(buttonPin)
	err := pin.Out(gpio.High)
	if err != nil {
		log.Printf("unable to push to pin %s: %v", buttonPin, err)
		os.Exit(1)
	}
	time.Sleep(duration)
	pin.Out(gpio.Low)
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

func getSensorPin(pin string) gpio.PinIO {
	thePin := gpioreg.ByName(pin)

	if err := thePin.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		log.Printf("unable to open pin %s: %v", pin, err)
		os.Exit(1)
	}

	return thePin
}

func getButtonPin(pin string) gpio.PinIO {
	return gpioreg.ByName(pin)
}

func connectMqtt(config *config) (MQTT.Client, error) {
	log.Println("mqtt: preparing connection")
	server := config.mqtt.server
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
