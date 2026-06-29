package db

import (
	"fmt"

	"github.com/pimentaluan/microservices/shipping/internal/application/core/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Shipping struct {
	gorm.Model
	OrderID              int64
	DeliveryForecastDays int32
	Items                []ShippingItem
}

type ShippingItem struct {
	gorm.Model
	ProductCode string
	Quantity    int32
	ShippingID  uint
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(dataSourceURL string) (*Adapter, error) {
	database, err := gorm.Open(mysql.Open(dataSourceURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no banco: %v", err)
	}

	err = database.AutoMigrate(&Shipping{}, &ShippingItem{})
	if err != nil {
		return nil, fmt.Errorf("erro ao executar migration: %v", err)
	}

	return &Adapter{db: database}, nil
}

func (a Adapter) Save(shipping *domain.Shipping) error {
	var shippingItems []ShippingItem

	for _, item := range shipping.Items {
		shippingItems = append(shippingItems, ShippingItem{
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
		})
	}

	shippingModel := Shipping{
		OrderID:              shipping.OrderID,
		DeliveryForecastDays: shipping.DeliveryForecastDays,
		Items:                shippingItems,
	}

	result := a.db.Create(&shippingModel)
	if result.Error != nil {
		return result.Error
	}

	shipping.ID = int64(shippingModel.ID)

	return nil
}
