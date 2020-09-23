package subservice

import (
	"database/sql"
	"strings"
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
	db           *sql.DB
	ProductSubs  map[productID][]string
	productPrice map[productID]string
}

func NewSubService(db *sql.DB) *SubService {
	return &SubService{
		db:           db,
		ProductSubs:  make(map[productID][]string),
		productPrice: make(map[productID]string),
	}
}

// -------------------------------------------------- в functions

func (ss *SubService)AddSubscriber(id productID, mail string) {
	ss.ProductSubs[id] = append(ss.ProductSubs[id], mail)
}

func TrimProductLink(link string) productID {
	idx := strings.LastIndex(link, "_")
	return productID(link[idx+1:])
}

