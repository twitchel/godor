package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
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

		for _ = range ticker.C {
			if c.emulate {

				_, err := ioutil.ReadFile("./open")

				var isOpen bool
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
				// REAL PIN!!!!!
				pings <- ping{
					isOpen: true,
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

				log.Printf("Fire: %v", p.isOpen)
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

	log.Println(mqttTopic)

	return &config{
		emulate: true,
		mqtt: &configMqtt{
			server: mqttServer,
			topic:  mqttTopic,
		},

		sensor: &configSensor{
			pin:       1,
			checkRate: time.Second,
		},

		button: &configButton{
			pin: 1,
		},
	}
}

func connectMqtt(server string) (MQTT.Client, error) {
	log.Println("MQTT: preparing connection")
	connOpts := MQTT.NewClientOptions().AddBroker(server).SetCleanSession(true)
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	connOpts.SetTLSConfig(tlsConfig)

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return client, token.Error()
	}
	log.Printf("MQTT: connected to %s\n", server)
	return client, nil
}
