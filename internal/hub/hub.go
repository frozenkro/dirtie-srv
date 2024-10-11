package hub

import (
	"fmt"
	"os"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/frozenkro/dirtie-srv/internal/core/topics"
)

var (
	totalReconnectAttempts int = 10
	client                 mqtt.Client
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	topic := string(msg.Topic())
	fmt.Printf("Received message: %s from topic %s\n", string(msg.Payload()), topic)

	switch topic {
	case topics.Breadcrumb:
		// need deps injected
	case topics.Provision:
		// yep still need deps
	default:
		fmt.Printf("Topic %v not recognized", topic)
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Printf("connected to mqtt broker\n")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("disconnected from mqtt broker\n")
	attemptReconnect()
}

func attemptReconnect() {
	for i := 1; i <= totalReconnectAttempts; i++ {
		fmt.Printf("attempting to reconnect (%d/%d)\n", i, totalReconnectAttempts)

		if token := client.Connect(); token.Wait() && token.Error() == nil {
			fmt.Printf("Reconnect successful\n")
			return
		}
	}
	panic("Failed to reconnect to mqtt broker")
}

func Init() {
	uri, ok := os.LookupEnv("MOSQUITTO_URI")
	if !ok {
		uri = "localhost:1883"
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri))
	opts.SetClientID("dirtie_hub")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectionLostHandler
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// this just keeps the goroutine alive
	select {}
}
