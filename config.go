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
	topic     string
}

type configButton struct {
	pin   string
	topic string
	pressDuration time.Duration
}

type config struct {
	emulate bool
	mqtt    *configMqtt
	sensor  *configSensor
	button  *configButton
}

func prepareConfig() *config {

	sensorPin := env("GODOR_SENSOR_PIN", "GPIO22")
	buttonPin := env("GODOR_BUTTON_PIN", "GPIO16")

	emulateString := env("EMULATE", "false")
	emulate, _ := strconv.ParseBool(emulateString)

	return &config{
		emulate: emulate,
		mqtt: &configMqtt{
			server: env("MQTT_SERVER", "tcp://192.168.1.194:1883"),
		},

		sensor: &configSensor{
			pin:       sensorPin,
			checkRate: time.Millisecond * 100,
			topic:     env("SENSOR_MQTT_TOPIC", "godor/door1/state"),
		},

		// not used yet
		button: &configButton{
			pin:   buttonPin,
			topic: env("BUTTON_MQTT_TOPIC", "godor/door1/trigger"),
			pressDuration: 1000 * time.Millisecond,
		},
	}
}

func env(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
