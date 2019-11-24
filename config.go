package main

import (
	"log"
	"os"
	"strconv"
	"time"
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
