package hub

import (
	"context"
	"fmt"
	"os"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/frozenkro/dirtie-srv/internal/core/topics"
	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/di"
)

var (
	totalReconnectAttempts int = 10
	client                 mqtt.Client
  deps                   *di.Deps
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

	topic := string(msg.Topic())
	utils.LogInfo(fmt.Sprintf("Received message: %s from topic %s\n", string(msg.Payload()), topic))

  var err error
	switch topic {
    case topics.Breadcrumb:
      err = deps.BrdCrmTopic.InvokeTopic(ctx, msg.Payload())
    case topics.Provision:
      err = deps.ProvisionTopic.InvokeTopic(ctx, msg.Payload())
    default:
      utils.LogInfo(fmt.Sprintf("Topic %v not recognized", topic))
	}

  if err != nil {
    utils.LogErr(fmt.Errorf("Error MessagePubHandler -> InvokeTopic: %w", err).Error())
  }
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	utils.LogInfo(fmt.Sprintf("connected to mqtt broker\n"))
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	utils.LogInfo(fmt.Sprintf("disconnected from mqtt broker\n"))
	attemptReconnect()
}

func attemptReconnect() {
	for i := 1; i <= totalReconnectAttempts; i++ {
		utils.LogInfo(fmt.Sprintf("attempting to reconnect (%d/%d)\n", i, totalReconnectAttempts))

		if token := client.Connect(); token.Wait() && token.Error() == nil {
			utils.LogInfo(fmt.Sprintf("Reconnect successful\n"))
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

  deps = di.NewDeps()

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
