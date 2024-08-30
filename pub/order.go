package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/gofrs/uuid/v5"
)

var localRand *rand.Rand

func init() {
	localRand = rand.New(rand.NewSource(time.Now().Unix()))
}

type Order struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     int       `json:"price"`
	Shop      string    `json:"shop"`
	Buyer     string    `json:"buyer"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

func (o Order) String() string {
	return fmt.Sprintf(
		"Order ID: %s\nName: %s\nPrice: %d\nShop: %s\nBuyer: %s\nAddress: %s\nCreated At: %s",
		o.ID.String(),
		o.Name,
		o.Price,
		o.Shop,
		o.Buyer,
		o.Address,
		o.CreatedAt.Format(time.RFC3339),
	)
}

func newOrder() Order {
	return Order{
		ID:        randomID(),
		Name:      randomItemName(),
		Price:     randomPrice(),
		Shop:      randomShop(),
		Buyer:     randomBuyer(),
		Address:   randomAddress(),
		CreatedAt: time.Now(),
	}
}

// simulation of user clicking buy item (create order)
type OrderSource struct {
	generateRate int // number of orders is created in 1 second
	OrderChan    chan Order
	done         chan struct{}
}

func NewOrderSource(rate int) *OrderSource {
	return &OrderSource{
		generateRate: rate,
		OrderChan:    make(chan Order),
		done:         make(chan struct{}),
	}
}

func (os *OrderSource) StopGenerate() {
	close(os.done)
}

func (os *OrderSource) CreateOrder() {
	gapInterval := time.Second * time.Duration(1/os.generateRate)
	for {
		select {
		case <-os.done:
			return
		default:
			order := newOrder()
			os.OrderChan <- order
			slog.Info("new order has been created", "order", order)
			time.Sleep(gapInterval)
		}
	}
}

func randomID() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		slog.Info("Failed to create uuid", "error", err.Error())
	}
	return id
}

func randomItemName() string {
	items := []string{"Laptop", "Pillow", "Headphones", "CoffeeMug", "Backpack", "Notebook", "Smartphone", "DeskLamp", "WaterBottle", "MousePad"}
	randomItemName := items[localRand.Intn(len(items))]
	return randomItemName
}

func randomPrice() int {
	return localRand.Intn(100) + 20 // 20 -> 119
}

func randomShop() string {
	shopNames := []string{"Tech Haven", "Cozy Corner", "Gadget Galaxy", "Elegant Emporium", "Urban Outfitters", "The Book Nook", "Fashion Forward", "Gourmet Delights", "Trendy Treasures", "Chic Boutique"}
	randomShopName := shopNames[localRand.Intn(len(shopNames))]
	return randomShopName
}

func randomBuyer() string {
	buyerNames := []string{"Alice Johnson", "Bob Smith", "Carol Davis", "David Wilson", "Emma Brown", "Frank Harris", "Grace Lee", "Henry Martin", "Ivy Clark", "Jack Turner"}
	randomBuyerName := buyerNames[localRand.Intn(len(buyerNames))]
	return randomBuyerName
}

func randomAddress() string {
	addresses := []string{"Springfield", "Shelbyville", "Capital City", "Rivertown", "Lakewood", "Metropolis", "Gotham", "Star City", "Central City", "Sunnydale"}
	randomAddress := addresses[localRand.Intn(len(addresses))]
	return randomAddress
}
