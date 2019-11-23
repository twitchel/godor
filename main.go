package main

import (
	"crypto/tls"
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	server, topic := prepareConfig()
	log.Println(topic)
	_, error := connectMqtt(server)

	if error != nil {
		panic(error)
	}
}

func prepareConfig() (string, string) {
	mqttServer, ok := os.LookupEnv("MQTT_SERVER")

	if !ok {
		mqttServer = "tcp://192.168.1.194:1883"
	}

	mqttTopic, ok := os.LookupEnv("MQTT_TOPIC")

	if !ok {
		mqttTopic = "godor/door1"
	}

	log.Println(mqttTopic)

	return mqttServer, mqttTopic
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
