package main

import (
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

	gpioPin := env("GODOR_PIN", "1")

	leIntPin, err := strconv.Atoi(gpioPin)
	if err != nil {
		panic("unable to convert pin number")
	}

	return &config{
		emulate: true,
		mqtt: &configMqtt{
			server: env("MQTT_SERVER", "tcp://192.168.1.194:1883"),
			topic:  env("MQTT_TOPIC", "godor/door1"),
		},

		sensor: &configSensor{
			pin:       leIntPin,
			checkRate: time.Second,
		},

		// not used yet
		button: &configButton{
			pin: 1,
		},
	}
}

func env(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
