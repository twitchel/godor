package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_prepareConfig(t *testing.T) {
	c := prepareConfig()
	assert.EqualValues(t, "tcp://192.168.1.194:1883", c.mqtt.server)
	assert.EqualValues(t, 1, c.sensor.pin)

}
