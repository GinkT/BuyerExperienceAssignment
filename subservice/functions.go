package subservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Загружает в память список подписок на товары из БД
func (SubServ *SubService) LoadSubMapFromDB() error {
	sqlStatement := `
		SELECT * FROM public."products"
	`
	rows, err := SubServ.db.Query(sqlStatement)
	if err != nil {
		log.Println("Error happened during load from DB!", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key ProductID
		var value []string
		err := rows.Scan(&key, pq.Array(&value))
		if err != nil {
			log.Fatal(err)
		}
		SubServ.ProductSubs[key] = value
	}

	// Заполняем таблицу цен продуктов актуальной информацией
	for key, _ := range SubServ.ProductSubs {
		SubServ.ProductPrices[key], _ = GetProductPrice(key)
	}
	return nil
}

// Загружает в БД список подписок на товары из памяти
func (SubServ *SubService) LoadSubMapToDB() error {
	for key, value := range SubServ.ProductSubs {
		sqlStatement := `
		INSERT INTO public."products"
		VALUES ($1, $2)
		ON CONFLICT ("productid")
		DO UPDATE 
		SET "subscribedusers" = $2
		`
		_, err := SubServ.db.Exec(sqlStatement, key, pq.Array(value))
		if err != nil {
			log.Println("Error happened during load to DB!", err)
			return err
		}
	}
	return nil
}

// Добавляет в базу подписку на продукт
func (SubServ *SubService)AddSubscriberToProduct(id ProductID, mail string) error {
	// Если продукта ещё нет в списке отслеживаемых товаров - добавляем
	if _, ok := SubServ.ProductPrices[id]; !ok {
		productPrice, err := GetProductPrice(id)
		if err != nil {
			return err
		}
		SubServ.ProductPrices[id] = productPrice
	}
	SubServ.ProductSubs[id] = append(SubServ.ProductSubs[id], mail)

	return nil
}

// Загружает подписки из БД, запускает цикл обновления цен.
// Раз в 40 секунд проходит по подпискаем и актуализирует значения.
func (SubServ *SubService)Run() {
	SubServ.LoadSubMapFromDB()

	for {
		time.Sleep(40 * time.Second)

		for productID, productPrice := range SubServ.ProductPrices {
			currentPrice, _ := GetProductPrice(productID)
			if currentPrice != productPrice {
				SubServ.SendMailToFollowers(currentPrice, SubServ.ProductSubs[productID])
			}
		}
	}
}

// Отправляет email с обновленной стоимостью всем подписчикам из emails[]
func (SubServ *SubService)SendMailToFollowers(cost string, emails []string) {
	message := `From: "Other User" <buyerjobassignment@yandex.ru>
cc: 
Subject: The price of your subscribed product changed! Its costs now:` + cost

	if err := smtp.SendMail("smtp.yandex.ru:587", SubServ.mailerAuth, "buyerjobassignment@yandex.ru", emails, []byte(message)); err != nil {
		fmt.Println("Error SendMail: ", err)
	}
	fmt.Println("Send emails to:", emails)
}

// Отправляет письмо. Внутри письма ссылка с подтверждением
func (SubServ *SubService)SendConfirmationEmail(link, email string) {
	message := `From: "Сервис подписки" <buyerjobassignment@yandex.ru>
cc: 
Subject: Please enter this link to a browser! http://localhost:8181/subscribe?link=` + link + `&mail=` + email + `&code=` + SubServ.ConfirmCode
	if err := smtp.SendMail("smtp.yandex.ru:587", SubServ.mailerAuth, "buyerjobassignment@yandex.ru", []string{email}, []byte(message)); err != nil {
		fmt.Println("Error SendMail: ", err)
	}
	fmt.Println("Send emails to:", email)
}

// Обрезает ссылку и получает ID объявления
func TrimProductLink(link string) ProductID {
	idx := strings.LastIndex(link, "_")
	return ProductID(link[idx+1:])
}

// Получает цену объявления используя API авито. Входной параметр - ID объявления
func GetProductPrice(id ProductID) (string, error) {
	log.Println("Requested price for product id:", id)

	resp, err := http.Get("https://m.avito.ru/api/14/items/" + string(id) + "?key=af0deccbgcgidddjgnvljitntccdduijhdinfgjgfjir")
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Invalid link")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if !json.Valid(body) {
		log.Println(err)
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	price := result["price"].(map[string]interface{})["value"]

	log.Println("Unmarshalled JSON for price, got:", price)
	return price.(string), nil
}