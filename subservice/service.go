package subservice

import (
	"database/sql"
	"net/smtp"
	"sync"
)

/*
	Что умеет делать сервис:
	Запускаясь, загружает таблицу с подписками на товары в свою структуру (   map[product][]string   )
	Раз в n секунд проходит по данной структуре и сверяется с ценами на товары(сверяется с помощью запросов к
	Avito API. Если цена не актуальна, отправляет уведомления на почту подписанным на товар юзерам и обновляет
	цену в структуре и базе данных.
 */

type productID string
type productPrice string

type SubService struct {
	mailerAuth    smtp.Auth
	mu            *sync.Mutex
	db            *sql.DB
	ProductSubs   map[productID][]string
	ProductPrices map[productID]string
}

func NewSubService(db *sql.DB) *SubService {
	return &SubService{
		mailerAuth: 	smtp.PlainAuth("", "buyerjobassignment@yandex.ru", "192837465", "smtp.yandex.ru"),
		db:            	db,
		ProductSubs:   	make(map[productID][]string),
		ProductPrices: 	make(map[productID]string),
	}
}