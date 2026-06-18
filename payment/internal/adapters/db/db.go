package db

import (
	"fmt"

	"github.com/pimentaluan/microservices/payment/internal/application/core/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	UserID     int64
	OrderID    int64
	TotalPrice float32
}

type Bill struct {
	gorm.Model
	PaymentID uint
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(dataSourceURL string) (*Adapter, error) {
	database, err := gorm.Open(mysql.Open(dataSourceURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no banco: %v", err)
	}

	err = database.AutoMigrate(&Payment{}, &Bill{})
	if err != nil {
		return nil, fmt.Errorf("erro ao executar migration: %v", err)
	}

	return &Adapter{db: database}, nil
}

func (a Adapter) Save(payment *domain.Payment) error {
	paymentModel := Payment{
		UserID:     payment.UserID,
		OrderID:    payment.OrderID,
		TotalPrice: payment.TotalPrice,
	}

	result := a.db.Create(&paymentModel)
	if result.Error != nil {
		return result.Error
	}

	billModel := Bill{PaymentID: paymentModel.ID}
	result = a.db.Create(&billModel)
	if result.Error != nil {
		return result.Error
	}

	payment.ID = int64(paymentModel.ID)
	payment.BillID = int64(billModel.ID)

	return nil
}
