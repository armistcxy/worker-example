package main

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderProducer struct {
	ch *amqp.Channel
}

var ExchangeOrderName = "order_event_exchange"

func NewOrderProducer(conn *amqp.Connection) (*OrderProducer, error) {
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("Failed to create a channel", "error", err.Error())
		return nil, err
	}

	// I want to make this simple, so just use "direct exchange" here, "topic exchange" will be used in case like this
	// `order.*` -> send to all queue that will process all type of order
	// `order.*.urgent` -> send to queue that will process urgent order

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
		slog.Error("failed to declare exchange", "error", err.Error())
		return nil, err
	}

	return &OrderProducer{
		ch: ch,
	}, nil
}

// if queue already created, this will not affect.
func (op *OrderProducer) CreateQueue(name string, opts ...Option) error {
	queueCfg := defaultConfig

	for _, opt := range opts {
		opt.apply(&queueCfg)
	}

	_, err := op.ch.QueueDeclare(
		name,
		queueCfg.durable,
		queueCfg.autoDelete,
		queueCfg.exclusive,
		queueCfg.noWait,
		nil,
	)

	return err
}

func (op *OrderProducer) QueueBind(queueName string, routingKey string) error {
	return op.ch.QueueBind(queueName, routingKey, ExchangeOrderName, false, nil)
}

func (op *OrderProducer) publish(routingKey string, content []byte, contentType string) error {
	return op.ch.Publish(
		ExchangeOrderName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        content,
		},
	)
}

func (op *OrderProducer) PublishLog(order *Order) error {
	return op.publish("log.orders", createLogContent(order), "application/json")
}

type LogContentForm struct {
	OrderID   uuid.UUID `json:"order_id"`
	Buyer     string    `json:"buyer"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

func createLogContent(order *Order) []byte {
	logContent := LogContentForm{
		OrderID:   order.ID,
		Buyer:     order.Buyer,
		Price:     order.Price,
		CreatedAt: order.CreatedAt,
	}

	jsonData, err := json.Marshal(logContent)
	if err != nil {
		slog.Info("Failed to serialize log content", "error", err.Error())
		return []byte("")
	}
	return jsonData
}

func (op *OrderProducer) PublishUserNotification(order *Order) error {
	return op.publish("notify.users", createUserNotification(order), "application/json")
}

type UserNotification struct {
	OrderID uuid.UUID `json:"order_id"`
	Item    string    `json:"item"`
}

func createUserNotification(order *Order) []byte {
	userNoti := UserNotification{
		OrderID: order.ID,
		Item:    order.Name,
	}

	jsonData, err := json.Marshal(userNoti)
	if err != nil {
		slog.Info("Failed to serialize user notification", "error", err.Error())
		return []byte("")
	}
	return jsonData
}

func (op *OrderProducer) PublishSellerNotification(order *Order) error {
	return op.publish("notify.sellers", createSellerNotification(order), "application/json")
}

type SellerNotification struct {
	OrderID uuid.UUID `json:"order_id"`
	Buyer   string    `json:"buyer"`
	Address string    `json:"address"`
}

func createSellerNotification(order *Order) []byte {
	sellerNoti := SellerNotification{
		OrderID: order.ID,
		Buyer:   order.Buyer,
		Address: order.Address,
	}

	jsonData, err := json.Marshal(sellerNoti)
	if err != nil {
		slog.Info("Failed to serialize seller notification")
		return []byte("")
	}
	return jsonData
}

type queueConfig struct {
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
}

var defaultConfig = queueConfig{
	durable:    false,
	autoDelete: false,
	exclusive:  false,
	noWait:     false,
}

type Option interface {
	apply(cfg *queueConfig)
}

type DurableOpt struct {
	isDurable bool
}

func (do DurableOpt) apply(cfg *queueConfig) {
	cfg.durable = do.isDurable
}

type AutoDeleteOpt struct {
	isAutoDelete bool
}

func (ado AutoDeleteOpt) apply(cfg *queueConfig) {
	cfg.autoDelete = ado.isAutoDelete
}

type ExclusiveOpt struct {
	isExclusive bool
}

func (eo ExclusiveOpt) apply(cfg *queueConfig) {
	cfg.exclusive = eo.isExclusive
}

type NoWaitOpt struct {
	isNoWait bool
}

func (nwo NoWaitOpt) apply(cfg *queueConfig) {
	cfg.noWait = nwo.isNoWait
}
