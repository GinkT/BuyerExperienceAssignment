package subservice

import (
	"database/sql"
	"net/smtp"
)

/*
	Что умеет делать сервис:
	Запускаясь, загружает таблицу с подписками на товары из БД в свою структуру (   map[product][]string   )
	После запуска(метод Run), раз в n секунд проходит по данной структуре и сверяется с ценами на товары
	(сверяется с помощью запросов к Avito API. Если цена не актуальна, отправляет уведомления на почту подписанным
	на товар юзерам и обновляет цену в структуре и базе данных. Отправляет письмо с кодом подтверждения на почту.
 */

type SubscribeServiceInterface interface {
	LoadSubMapFromDB() error
	LoadSubMapToDB() error
	AddSubscriberToProduct(ProductID, string) error
	Run()
	SendMailToFollowers(string, []string)
	SendConfirmationEmail(string, string)
}

type ProductID string

type SubService struct {
	ConfirmCode   string
	mailerAuth    smtp.Auth
	db            *sql.DB
	ProductSubs   map[ProductID][]string
	ProductPrices map[ProductID]string
}

func NewSubService(db *sql.DB) *SubService {
	return &SubService{
		ConfirmCode:    "0000",
		mailerAuth: 	smtp.PlainAuth("", "avitobuyerexperience@yandex.ru", "192837465", "smtp.yandex.ru"),
		db:            	db,
		ProductSubs:   	make(map[ProductID][]string),
		ProductPrices: 	make(map[ProductID]string),
	}
}