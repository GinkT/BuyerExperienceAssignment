![alt text](https://i.imgur.com/D5h2K3Q.png "Архитектура сервиса")
 
# Пару слов о работе сервиса
Для ускорения доступа, сервис хранит данные о подписках/ценах товаров в оперативной памяти.       
При остановке(программа считывает сигнал SIGTERM)/запуске сервиса происходит синхронизация с БД.    

Данные хранятся в двух структурах:

    var ProductSubs     map[ProductID][]string
    var ProductPrices   map[ProductID][]

_где ProductID - это пользовательский тип string, а []string в ProductSubs - email'ы пользователей_

В запущенном состоянии сервис раз в N секунд проходит по всей хэш таблице ProductSubs,  
применяет к её ключам(ID товаров) функцию GetCurrentPrice и получает актуальную цену товара.   
В случае несовпадения с текущей ценой - обновляет цену и рассылает всем подписчикам товара письмо.

HTTP API обеспечивает доступ к методу подписки. Цепочка Middleware обработчиков проверяет на    
валидность введенные параметры. Если параметры правильные - использует метод сервиса SendConfirmMail    
чтобы отправить письмо с ссылкой подтверждением. В ссылке встроен код(параметр ?code),  
в случае его совпадения с кодом, хранящимся в сервисе - подписывает пользователя.

Для развёртывания сервиса создана docker среда. Имеется сценарий docker-compose в котором
прописаны переменные среды для конфигурации(тестировал только на этих параметрах).

    docker-compose build
    docker-compose up

Сервис доступен по порту **8181**, также имеется панель управления бд _adminer_ на порте **8080**.

Пример использования представлен в альбоме по ссылке: https://imgur.com/a/6BhsQod

В качестве smtp сервера используется сервер Яндекса. Он частенько банит меня за "спам"  
во время тестирования сервиса), но это потому  что я отправляю письма не с доменной почты.  
Я постараюсь оплатить и настроить домен в течение нескольких дней. 


# Фрагменты кода, решающие конкретные задачи

### Подписка на изменение цены

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
    
_Здесь и далее - muPP и muPS мьютексы, обеспечивающие безопасную работу с мапами_

### Отслеживание изменений цены

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

### Отправка уведомления на почту

    // Отправляет email с обновленной стоимостью всем подписчикам из emails[]
    func (SubServ *SubService)SendMailToFollowers(id ProductID, cost string, emails []string) {
        message := []byte("From: \"Сервис подписки\" <avitobuyerexperience@yandex.ru>\r\n" +
            "Subject: Изменение цены\r\n" +
            "\r\n" +
            "Здравствуйте, цена товара, на который вы подписались (https://www.avito.ru/"+ string(id) + ") изменилась и составляет: " + cost)
    
        if err := smtp.SendMail("smtp.yandex.ru:587", SubServ.mailerAuth, "avitobuyerexperience@yandex.ru", emails, message); err != nil {
            log.Println("[mail] Error SendMail: ", err)
            return
        }
        log.Println("[mail] Send sub-change emails to:", emails)
    }

### Работа с БД

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
                log.Println(err)
                return err
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

_Дата старта выполнения задания: 22.09.2020  
Дата окончания выполнения задания: 25.09.2020_