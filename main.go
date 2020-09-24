package main

import (
	"fmt"
	"github.com/GinkT/BuyerExperienceAssignment/db"
	"github.com/GinkT/BuyerExperienceAssignment/subservice"
	"log"
	"net/http"
	"regexp"
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
	database, err := db.NewDatabase(dbHost, dbPort, dbUser, dbPassword, dbBase)
	if err != nil {
		log.Println("Database connected!")
	}

	ss := subservice.NewSubService(database)
	env := &Env{SubService: ss}

	go ss.Run()

	http.Handle("/subscribe",  validSubscribeHandler(http.HandlerFunc(env.SubscribeHandle)))

	log.Println("Started to listen and serve on :8181")
	log.Fatalln(http.ListenAndServe(":8181", nil))
}

func (env *Env)SubscribeHandle(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	mail := r.URL.Query().Get("mail")

	env.SubService.AddSubscriberToProduct(subservice.TrimProductLink(link), mail)
	env.SubService.LoadSubMapToDB()

	fmt.Fprint(w, env.SubService.ProductSubs, env.SubService.ProductPrices)
}

// Валидация входных параметров link, mail
func validSubscribeHandler(subHandle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		link, mail := r.URL.Query().Get("link"), r.URL.Query().Get("mail")
		re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
		switch {
		case link == "":
			fmt.Fprintf(w, "Empty link!")
			return
		case mail == "":
			fmt.Fprintf(w, "Empty mail!")
			return
		case !re.MatchString(mail):
			fmt.Fprintf(w, "Invalid mail!")
			return
		}
		subHandle.ServeHTTP(w, r)
	})
}