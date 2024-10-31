package hub

import (
	"context"
	"fmt"
	"os"

	"github.com/eclipse/paho.mqtt.golang"
	core_topics "github.com/frozenkro/dirtie-srv/internal/core/topics"
	"github.com/frozenkro/dirtie-srv/internal/core/utils"
	"github.com/frozenkro/dirtie-srv/internal/di"
)

type TopicInvoker interface {
	InvokeTopic(ctx context.Context, payload []byte) error
}

var (
	totalReconnectAttempts int = 10
	client                 mqtt.Client
	deps                   *di.Deps
	ErrTopicNotFound       error = fmt.Errorf("MQTT Topic Not Found")
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := string(msg.Topic())
	utils.LogInfo(fmt.Sprintf("Received message: %s from topic %s\n", string(msg.Payload()), topic))

	ivk, err := getTopicInvoker(msg.Topic())
	if ivk != nil {
		err = ivk.InvokeTopic(ctx, msg.Payload())
	}

	if err != nil {
		utils.LogErr(fmt.Errorf("Error MessagePubHandler -> InvokeTopic: %w", err).Error())
	}
}

func getTopicInvoker(topic string) (TopicInvoker, error) {
	switch topic {
	case core_topics.Breadcrumb:
		return deps.BrdCrmTopic, nil
	case core_topics.Provision:
		return deps.ProvisionTopic, nil
	default:
		return nil, ErrTopicNotFound
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

func Init(deps *di.Deps) {
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
