package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stianeikeland/go-rpio"
)

type configMqtt struct {
	server string
	topic  string
}

type configSensor struct {
	pin       int
	checkRate time.Duration
}

type configButton struct {
	pin int
}

type config struct {
	emulate bool
	mqtt    *configMqtt
	sensor  *configSensor
	button  *configButton
}

type ping struct {
	isOpen bool
	time   time.Time
}

func main() {
	c := prepareConfig()
	log.Println(c.mqtt.topic)
	client, error := connectMqtt(c.mqtt.server)

	if error != nil {
		panic(error)
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
			var isOpen bool

			if c.emulate {

				_, err := ioutil.ReadFile("./open")

				if err != nil {
					isOpen = false
				} else {
					isOpen = true
				}

				pings <- ping{
					isOpen: isOpen,
					time:   time.Now(),
				}
			} else {
				pinState := pin.Read()

				if pinState == rpio.Low {
					isOpen = false
				} else {
					isOpen = true
				}

				pings <- ping{
					isOpen: isOpen,
					time:   time.Now(),
				}
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

func prepareConfig() *config {

	mqttServer, ok := os.LookupEnv("MQTT_SERVER")

	if !ok {
		mqttServer = "tcp://192.168.1.194:1883"
	}

	mqttTopic, ok := os.LookupEnv("MQTT_TOPIC")

	if !ok {
		mqttTopic = "godor/door1"
	}

	lePin, ok := os.LookupEnv("GODOR_PIN")
	if !ok {
		lePin = "1"
	}

	leIntPin, err := strconv.Atoi(lePin)
	if err != nil {
		panic("Wrong pin")
	}

	log.Println(mqttTopic)
	log.Println(lePin)

	return &config{
		emulate: true,
		mqtt: &configMqtt{
			server: mqttServer,
			topic:  mqttTopic,
		},

		sensor: &configSensor{
			pin:       leIntPin,
			checkRate: time.Second,
		},

		button: &configButton{
			pin: 1,
		},
	}
}

func getPin(pin int) rpio.Pin {
	thePin := rpio.Pin(pin)

	if err := rpio.Open(); err != nil {
		log.Printf("unable to open pin %i: %v", pin, err)
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
