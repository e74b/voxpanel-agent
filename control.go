package main

import (
	"encoding/json"
	"os"
	amqp "github.com/rabbitmq/amqp091-go"
)

const AMQP_URL = "amqp://default:password@127.0.0.1/"
const AMQP_CONTROL_EXCHANGE = "control-exchange"


type Routing struct {
	correlationId string
	replyQueue string
}

func rpcReply (app *AppState, routing Routing, response any) {
	app.controlChannelLock.Lock()
	defer app.controlChannelLock.Unlock()
	encodedBody, err := json.Marshal(response)
	app.warnOnError(err, "Failed JSON marshall")

	app.controlChannel.Publish(AMQP_CONTROL_EXCHANGE, routing.replyQueue, false, false, amqp.Publishing{
		CorrelationId: routing.correlationId,
		Body: encodedBody,
	})
}

func startControlReceiver (app *AppState) {
	app.logger.Info("RPC Listener started.")
	connection := app.amqpConnection;

	channel, err := connection.Channel()
	app.panicOnError(err, "Creating channel to control socket")

	delivery, err :=  channel.Consume(
		os.Getenv("VP_CONTROL_QUEUE"), 
		os.Getenv("VP_AGENT_NAME"),
		true, false, false, false, nil,
	)
	app.panicOnError(err, "Consuming control queue")
	for message := range delivery {
		data := ControlMessage{}
		err = json.Unmarshal(message.Body, &data)
		app.warnOnError(err, "Decoding JSON")
		// message.Ack(false)
		app.logger.Infof("Received rpc command `%s`", data.Cmd)
		routing := Routing{
			correlationId: message.CorrelationId,
			replyQueue: message.ReplyTo,
		}
		switch (data.Cmd) {
		case "ping":
		 	go handlePing(data, routing, app);
		 case "start_server":
		 	go handleStartServer(data, routing, app);
		 case "stop_server":
		 	go handleStopServer(data, routing, app)
		 case "get_running":
		 	go handleGetRunning(data, routing, app)
		}
	}
}

