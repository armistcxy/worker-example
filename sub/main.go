package main

import (
	"log"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitmqURL = "amqp://guest:guest@localhost:5672/"

func main() {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("failed when connection to rabbitmq server")
	}
	defer conn.Close()

	numberOfConsumer := 3

	forever := make(chan struct{})

	for i := range numberOfConsumer {
		orderConsumer, err := NewOrderConsumer(conn, i+1)
		if err != nil {
			slog.Error("", "error", err.Error())
			log.Fatal("failed when create order consumer")
		}
		defer orderConsumer.ch.Close()

		go orderConsumer.Work()
	}

	<-forever

}
