package db

import (
	"fmt"

	"github.com/pimentaluan/microservices/order/internal/application/core/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	CustomerID int64
	Status     string
	OrderItems []OrderItem
}

type OrderItem struct {
	gorm.Model
	ProductCode string
	UnitPrice   float32
	Quantity    int32
	OrderID     uint
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(dataSourceURL string) (*Adapter, error) {
	database, err := gorm.Open(mysql.Open(dataSourceURL), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no banco: %v", err)
	}

	err = database.AutoMigrate(&Order{}, &OrderItem{})

	if err != nil {
		return nil, fmt.Errorf("erro ao executar migration: %v", err)
	}

	return &Adapter{
		db: database,
	}, nil
}

func (a Adapter) Get(id string) (domain.Order, error) {
	var orderEntity Order

	result := a.db.Preload("OrderItems").First(&orderEntity, id)

	var orderItems []domain.OrderItem

	for _, item := range orderEntity.OrderItems {
		orderItems = append(orderItems, domain.OrderItem{
			ProductCode: item.ProductCode,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
		})
	}

	order := domain.Order{
		ID:         int64(orderEntity.ID),
		CustomerID: orderEntity.CustomerID,
		Status:     orderEntity.Status,
		OrderItems: orderItems,
		CreatedAt:  orderEntity.CreatedAt.Unix(),
	}

	return order, result.Error
}

func (a Adapter) Save(order *domain.Order) error {
	var orderItems []OrderItem

	for _, item := range order.OrderItems {
		orderItems = append(orderItems, OrderItem{
			ProductCode: item.ProductCode,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
		})
	}

	orderModel := Order{
		CustomerID: order.CustomerID,
		Status:     order.Status,
		OrderItems: orderItems,
	}

	result := a.db.Create(&orderModel)

	if result.Error == nil {
		order.ID = int64(orderModel.ID)
	}

	return result.Error
}

func (a Adapter) UpdateStatus(id int64, status string) error {
	return a.db.Model(&Order{}).Where("id = ?", id).Update("status", status).Error
}
