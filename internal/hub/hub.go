package hub

import (
	"fmt"
	"os"

	"github.com/eclipse/paho.mqtt.golang"
)

var (
	totalReconnectAttempts int = 10
	client                 mqtt.Client
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic %s\n", msg.Payload, msg.Topic)
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
	// TODO confirm this is necessary
	select {}
}
