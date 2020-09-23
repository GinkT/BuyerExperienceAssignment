package main

import (
	"fmt"
	"github.com/GinkT/BuyerExperienceAssignment/db"
	"github.com/GinkT/BuyerExperienceAssignment/subservice"
	"log"
	"net/http"
)

const (
	dbHost = "db"
	dbPort = 5432
	dbUser = "postgres"
	dbPassword = "qwerty"
	dbBase = "SubscribeService"
)

type Env struct {
	SubService *subservice.SubService
}

func main() {
	database, _ := db.NewDatabase(dbHost, dbPort, dbUser, dbPassword, dbBase)

	ss := subservice.NewSubService(database)
	env := &Env{SubService: ss}

	http.HandleFunc("/subscribe", env.SubscribeHandle)

	go ss.Run()

	log.Println("Started to listen and serve on :8080")
	log.Fatalln(http.ListenAndServe(":8080", nil))


}

func (env *Env)SubscribeHandle(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	mail := r.URL.Query().Get("mail")

	env.SubService.AddSubscriberToProduct(subservice.TrimProductLink(link), mail)

	fmt.Fprint(w, env.SubService.ProductSubs, env.SubService.ProductPrices)
}
