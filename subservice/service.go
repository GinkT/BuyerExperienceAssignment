package subservice

import (
	"database/sql"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"sync"
)

/*
	Что умеет делать сервис:
	Запускаясь, загружает таблицу с подписками на товары из БД в свою структуру (   map[productID][]string   ).
	Хранит цены на товары в отдельной структуре вида (    map[productID]string     )
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
	SyncTime			int
	ConfirmCode   		string
	mailerAuth    		smtp.Auth
	db            		*sql.DB

	ProductSubs   		map[ProductID][]string
	muPS		  		*sync.Mutex

	ProductPrices 		map[ProductID]string
	muPP		  		*sync.Mutex
}

func NewSubService(db *sql.DB) (*SubService, error) {
	syncTime, err := strconv.Atoi(os.Getenv("SUB_SERVICE_SYNC_TIME"))
	if err != nil {
		log.Println("[SubService] Invalid SUB_SERVICE_SYNC_TIME config!")
		return nil, err
	}
	if syncTime < 3 {
		log.Println("[SubService] Please, choose at least a 3 sec SUB_SERVICE_SYNC_TIME!")
		return nil, err
	}
	return &SubService{
		SyncTime:		syncTime,
		ConfirmCode:    "0000",
		mailerAuth: 	smtp.PlainAuth("", "avitobuyerexperience@yandex.ru", "192837465", "smtp.yandex.ru"),
		db:            	db,
		ProductSubs:   	make(map[ProductID][]string),
		ProductPrices: 	make(map[ProductID]string),
	}, nil
}