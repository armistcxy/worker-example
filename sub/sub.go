package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderConsumer struct {
	ID             int
	ch             *amqp.Channel
	logOrderChan   <-chan amqp.Delivery
	userNotiChan   <-chan amqp.Delivery
	sellerNotiChan <-chan amqp.Delivery
}

var ExchangeOrderName = "order_event_exchange"

func NewOrderConsumer(conn *amqp.Connection, id int) (*OrderConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("Failed to create a channel", "error", err.Error())
		return nil, err
	}

	err = ch.ExchangeDeclare(
		ExchangeOrderName,
		"direct",
		true, // persistent for logging is a must, I think ...
		false,
		false,
		false, // esnsure that exchange has been created in the server before any other actions.
		nil,
	)

	if err != nil {
		slog.Error("exchange not matched", "error", err.Error())
		return nil, err
	}

	logOrderDeliveryChan, err := ch.Consume(
		"order log",
		"", //  empty string will cause the library to generate a unique identity
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to create consumer fetch message from order log queue", "error", err.Error())
		return nil, err
	}

	userNotiDeliveryChan, err := ch.Consume(
		"user notify",
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to create consumer fetch message from user notify queue", "error", err.Error())
		return nil, err
	}

	sellerNotiDeliveryChan, err := ch.Consume(
		"seller notify",
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to create consumer fetch message from seller notify queue", "error", err.Error())
		return nil, err
	}

	return &OrderConsumer{
		ID:             id,
		ch:             ch,
		logOrderChan:   logOrderDeliveryChan,
		userNotiChan:   userNotiDeliveryChan,
		sellerNotiChan: sellerNotiDeliveryChan,
	}, nil
}

// what if we close 1 of 3 channel ?
func (oc *OrderConsumer) Work() {
	for {
		select {
		case d := <-oc.logOrderChan:
			var logContent LogContent
			if err := json.Unmarshal(d.Body, &logContent); err != nil {
				slog.Info("failed to deserialize log content", "error", err.Error())
				continue
			}

			slog.Info("Log for order", "consumer-id", oc.ID, "log", fmt.Sprintf("%+v\n", logContent))
		case d := <-oc.userNotiChan:
			var userNoti UserNotification
			if err := json.Unmarshal(d.Body, &userNoti); err != nil {
				slog.Info("failed to deserialize user notification", "error", err.Error())
				continue
			}
			slog.Info("Log for user notification", "consumer-id", oc.ID, "user notification", fmt.Sprintf("%+v\n", userNoti))
		case d := <-oc.sellerNotiChan:
			var sellerNoti SellerNotification
			if err := json.Unmarshal(d.Body, &sellerNoti); err != nil {
				slog.Info("failed to deserialize seller notification", "error", err.Error())
				continue
			}
			slog.Info("Log for seller notification", "consumer-id", oc.ID, "seller notification", fmt.Sprintf("%+v\n", sellerNoti))
		}
	}
}

type LogContent struct {
	OrderID   uuid.UUID `json:"order_id"`
	Buyer     string    `json:"buyer"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

type UserNotification struct {
	OrderID uuid.UUID `json:"order_id"`
	Item    string    `json:"item"`
}

type SellerNotification struct {
	OrderID uuid.UUID `json:"order_id"`
	Buyer   string    `json:"buyer"`
	Address string    `json:"address"`
}
