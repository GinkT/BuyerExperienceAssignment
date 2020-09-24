package subservice

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (SubServ *SubService)InitFromDB(db *sql.DB) error {
	return nil
}

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
		var key productID
		var value []string
		err := rows.Scan(&key, pq.Array(&value))
		if err != nil {
			log.Fatal(err)
		}
		SubServ.ProductSubs[key] = value
	}
	return nil
}

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

func (SubServ *SubService)AddSubscriberToProduct(id productID, mail string) {
	if _, ok := SubServ.ProductPrices[id]; !ok {
		SubServ.ProductPrices[id], _ = GetProductPrice(id)
	}
	SubServ.ProductSubs[id] = append(SubServ.ProductSubs[id], mail)
}

func (SubServ *SubService)Run() {
	SubServ.LoadSubMapFromDB()
	fmt.Println(SubServ.ProductSubs)

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

// Обрезает ссылку и получает ID объявления
func TrimProductLink(link string) productID {
	idx := strings.LastIndex(link, "_")
	return productID(link[idx+1:])
}


// Получает цену объявления используя API авито. Входной параметр - ID объявления
func GetProductPrice(id productID) (string, error) {
	log.Println("Requested price for product id:", id)

	resp, err := http.Get("https://m.avito.ru/api/14/items/" + string(id) + "?key=af0deccbgcgidddjgnvljitntccdduijhdinfgjgfjir")
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(err)
		return "", err
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

func (ss *SubService)SendMailToFollowers(cost string, emails []string) {
	message := `From: "Other User" <buyerjobassignment@yandex.ru>
cc: 
Subject: The price of your subscribed product changed! Its costs now:` + cost

	if err := smtp.SendMail("smtp.yandex.ru:587", ss.mailerAuth, "buyerjobassignment@yandex.ru", emails, []byte(message)); err != nil {
		fmt.Println("Error SendMail: ", err)
	}
	fmt.Println("Send emails to:", emails)
}