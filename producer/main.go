package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

var rabbitmqURL = "amqp://guest:guest@localhost:5672/"

func main() {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to create to rabbitmq server")
	}
	defer conn.Close()

	orderPub, err := NewOrderProducer(conn)
	if err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to create order publisher")
	}
	defer orderPub.ch.Close()

	if err = orderPub.CreateQueue("order log", DurableOpt{true}); err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to create order log queue")
	}

	if err = orderPub.CreateQueue("user notify", DurableOpt{true}); err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to create user notify queue")
	}

	if err = orderPub.CreateQueue("seller notify", DurableOpt{true}); err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to create seller notify queue")
	}

	if err := orderPub.QueueBind("order log", "log.orders"); err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to bind order log queue to exchange")
	}

	if err := orderPub.QueueBind("user notify", "notify.users"); err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to bind user notify queue to exchange")
	}

	if err := orderPub.QueueBind("seller notify", "notify.sellers"); err != nil {
		slog.Error("", "error", err.Error())
		log.Fatal("Failed to bind seller notify queue to exchange")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // Ctrl + C

	rate := 1
	orderSource := NewOrderSource(rate)

	go orderSource.CreateOrder()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sigChan:
				slog.Info("Publisher server is shutting down")
				orderSource.StopGenerate()
				return
			case order := <-orderSource.OrderChan:
				if err := orderPub.PublishLog(&order); err != nil {
					slog.Info("Failed to publish log", "error", err.Error())
				}

				if err := orderPub.PublishUserNotification(&order); err != nil {
					slog.Info("Failed to publish user notification", "error", err.Error())
				}

				if err := orderPub.PublishSellerNotification(&order); err != nil {
					slog.Info("Failed to publish seller notification", "error", err.Error())
				}
			}
		}
	}()

	wg.Wait()
	slog.Info("Shutdown complete")
}
