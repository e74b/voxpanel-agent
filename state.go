package main

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type AppState struct {
	amqpConnection *amqp.Connection
	podmanConnection *context.Context
	// Channel for short events
	controlChannel *amqp.Channel
	controlChannelLock *sync.Mutex
	waitGroup *sync.WaitGroup
	logger *zap.SugaredLogger
}


