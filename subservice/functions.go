package subservice

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
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
		log.Println("[Database] Error happened getting rows from DB!", err)
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
	SubServ.muPP.Lock()
	if _, ok := SubServ.ProductPrices[id]; !ok {
		productPrice, err := GetProductPrice(id)
		if err != nil {
			return err
		}
		SubServ.ProductPrices[id] = productPrice
	}
	SubServ.muPP.Unlock()

	SubServ.muPS.Lock()
	SubServ.ProductSubs[id] = append(SubServ.ProductSubs[id], mail)
	SubServ.muPS.Unlock()

	return nil
}

// Загружает подписки из БД, запускает цикл обновления цен.
// Раз в n секунд(SYNC_TIME config) проходит по подпискаем и актуализирует значения.
func (SubServ *SubService)Run() {
	SubServ.LoadSubMapFromDB()

	go func () {
		for {
			time.Sleep(time.Duration(SubServ.SyncTime) * time.Second)
			log.Println("[SubService] Starting an update loop")

			for productID, productPrice := range SubServ.ProductPrices {
				currentPrice, err := GetProductPrice(productID)
				if err != nil {
					log.Println("[SubService] Error in Run loop:", err)
					continue
				}
				if currentPrice != productPrice {
					log.Printf("[SubService] Found changed price for product (%s): %s\n", productID, currentPrice)

					SubServ.muPS.Lock()
					SubServ.SendMailToFollowers(productID, currentPrice, SubServ.ProductSubs[productID])
					SubServ.muPS.Unlock()

					SubServ.muPP.Lock()
					SubServ.ProductPrices[productID] = currentPrice
					SubServ.muPP.Unlock()
				}
			}
			log.Println("[SubService] Ending an update loop")
		}
	} ()
}

// Обрезает ссылку и получает ID объявления
func TrimProductLink(link string) ProductID {
	idx := strings.LastIndex(link, "_")
	return ProductID(link[idx+1:])
}

// Получает цену объявления используя API авито. Входной параметр - ID объявления
func GetProductPrice(id ProductID) (string, error) {
	log.Println("[SubService] Requested price for product id:", id)

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

	log.Printf("[SubService] Got price for product(%s): %s", id, price)
	return price.(string), nil
}