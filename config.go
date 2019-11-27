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

	sensorPin := env("GODOR_SENSOR_PIN", "1")

	sensorPinInt, err := strconv.Atoi(sensorPin)
	if err != nil {
		panic("sensor: unable to convert pin number")
	}

	buttonPin := env("GODOR_BUTTON_PIN", "2")
	buttonPinInt, err := strconv.Atoi(buttonPin)
	if err != nil {
		panic("button: unable to convert pin number")
	}

	return &config{
		emulate: true,
		mqtt: &configMqtt{
			server: env("GODOR_MQTT_SERVER", "tcp://192.168.1.194:1883"),
			topic:  env("GODOR_SENSOR_MQTT_TOPIC", "godor/door1/state"),
		},

		sensor: &configSensor{
			pin:       sensorPinInt,
			checkRate: time.Second,
		},

		// not used yet
		button: &configButton{
			pin: buttonPinInt,
		},
	}
}

func env(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	// read from .env file here as second option

	return defaultValue
}
