package main

import (
	"context"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	amqp "github.com/rabbitmq/amqp091-go"
	podman "github.com/containers/podman/v5/pkg/bindings"
)

func setupEnvVars (app *AppState) {
	// TODO: fix env var setups
	err := godotenv.Load()
	app.panicOnError(err, "Loading dotenv")

	_, defined := os.LookupEnv("VP_PODMAN_SOCK")
	if (!defined) {
		os.Setenv("VP_PODMAN_SOCK", "unix:///run/user/1000/podman/podman.sock")
	}

	_, defined = os.LookupEnv("VP_AGENT_NAME")
	if (!defined) {
		os.Setenv("VP_AGENT_NAME", "agent-1")
	}

	_, defined = os.LookupEnv("VP_CONTROL_QUEUE")
	if (!defined) {
		os.Setenv("VP_CONTROL_QUEUE", "control")
	}
}

func main () {
	rootLogger, _ := zap.NewDevelopment()
	logger := rootLogger.Sugar()

	app := AppState{};
	app.waitGroup = &sync.WaitGroup{}
	app.logger = logger

	logger.Info("Reading env vars")
	setupEnvVars(&app)

	logger.Info("Connecting to podman socket")
	podmanConnection, err := podman.NewConnection(context.Background(), os.Getenv("VP_PODMAN_SOCK"))
	app.panicOnError(err, "Connecting to podman socket")
	app.podmanConnection = &podmanConnection

	logger.Info("Connecting to rabbitmq")
	amqpConnection, err := amqp.Dial(AMQP_URL)
	app.panicOnError(err, "Connecting to control socket")
	app.amqpConnection = amqpConnection

	controlChannel, err := amqpConnection.Channel()
	app.panicOnError(err, "Creating control response channel")
	app.controlChannel = controlChannel
	err = app.controlChannel.QueueBind(
		os.Getenv("VP_CONTROL_QUEUE"),
		os.Getenv("VP_AGENT_NAME"),
		AMQP_CONTROL_EXCHANGE,
		false, nil,
	)
	// panicOnError(err, "Binding control queue to agent name")
	app.controlChannelLock = &sync.Mutex{}

	go startControlReceiver(&app)
	logger.Info("Startup complete.")
	app.waitGroup.Add(1)
	app.waitGroup.Wait()
}

