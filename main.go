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
	envir := &Env{SubService: ss}

	http.HandleFunc("/subscribe", envir.SubscribeHandle)

	log.Println("Started to listen and serve on :8080")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func (check *Env)SubscribeHandle(w http.ResponseWriter, r *http.Request) {


	fmt.Fprint(w, check.SubService.ProductSubs)
}
