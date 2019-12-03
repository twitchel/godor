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
	pin       string
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

	emulateString := env("EMULATE", "false")
	emulate, _ := strconv.ParseBool(emulateString)

	return &config{
		emulate: emulate,
		mqtt: &configMqtt{
			server: env("MQTT_SERVER", "tcp://192.168.1.194:1883"),
			topic:  env("MQTT_TOPIC", "godor/door1"),
		},

		sensor: &configSensor{
			pin:       gpioPin,
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
