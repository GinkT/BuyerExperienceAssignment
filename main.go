package main

import (
	"context"
	"encoding/json"
	"github.com/GinkT/BuyerExperienceAssignment/db"
	"github.com/GinkT/BuyerExperienceAssignment/subservice"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
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
	//Соединение с БД
	database, err := db.NewDatabase(dbHost, dbPort, dbUser, dbPassword, dbBase)
	if err != nil {
		log.Fatalln("Database was not connected!")
	}
	defer database.Close()

	// Инициализация сервиса
	ss := subservice.NewSubService(database)
	env := &Env{SubService: ss}

	// Запуск сервиса
	ss.Run()

	// HTTP API сервиса
	srv := &http.Server{}
	http.Handle("/subscribe",  ValidSubscribeHandler(env.confirmSubscriber(http.HandlerFunc(env.SubscribeHandle))))
	log.Println("Started to listen and serve on :8181")
	ln, _ := net.Listen("tcp", ":8181")
	go func () {
		log.Fatalln(srv.Serve(ln))
	} ()

	// Graceful shutdown
	log.Println("Ready for graceful")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	ss.LoadSubMapToDB()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Gracefully shut down!")
}

type OkJson struct {
	ProductID 	subservice.ProductID		`json:"productid"`
	Email		string						`json:"email"`
	Message		string						`json:"message"`
}

// Метод подписки на объявление
func (env *Env)SubscribeHandle(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	mail := r.URL.Query().Get("mail")

	prodID := subservice.TrimProductLink(link)
	env.SubService.AddSubscriberToProduct(prodID, mail)

	okMsg := &OkJson{
		ProductID: prodID,
		Email:     mail,
		Message:   "Subscribe Confirmed! Congrats!",
	}
	json.NewEncoder(w).Encode(&okMsg)
	w.WriteHeader(http.StatusOK)
}

// ----------------------------------------------------- MiddleWare for SubscribeHandle

type ErrorJson struct {
	Error map[string]interface{} 		`json:"error"`
}

// Валидация входных параметров link, mail
func ValidSubscribeHandler(subHandle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		link, mail := r.URL.Query().Get("link"), r.URL.Query().Get("mail")
		re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
		switch {
		case link == "":
			error := &ErrorJson{
				Error: map[string]interface{}{
					"code": "404",
					"message": "Empty link!",
				},
			}
			json.NewEncoder(w).Encode(&error)
			w.WriteHeader(http.StatusNotFound)
			return
		case mail == "":
			error := &ErrorJson{
				Error: map[string]interface{}{
					"code": "404",
					"message": "Empty mail!",
				},
			}
			json.NewEncoder(w).Encode(&error)
			w.WriteHeader(http.StatusNotFound)
			return
		case !re.MatchString(mail):
			error := &ErrorJson{
				Error: map[string]interface{}{
					"code": "404",
					"message": "Invalid email!",
				},
			}
			json.NewEncoder(w).Encode(&error)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		subHandle.ServeHTTP(w, r)
	})
}

// Проверка кода подтверждения, в случае отсутствия - отправка письма с кодом
func (env *Env)confirmSubscriber(validatedHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		link, mail := r.URL.Query().Get("link"), r.URL.Query().Get("mail")
		confirmCode := r.URL.Query().Get("code")
		switch {
		case confirmCode == "":
			env.SubService.SendConfirmationEmail(link, mail)
			error := &ErrorJson{
				Error: map[string]interface{}{
					"code": "409",
					"message": "Пожалуйста, подтвердите вашу подписку, воспользовавшись ссылкой из письма!",
				},
			}
			json.NewEncoder(w).Encode(&error)
			w.WriteHeader(http.StatusConflict)
			return
		case confirmCode != env.SubService.ConfirmCode:
			error := &ErrorJson{
				Error: map[string]interface{}{
					"code": "404",
					"message": "Ошибка подтверждения! Воспользуйтесь ссылкой из письма!",
				},
			}
			json.NewEncoder(w).Encode(&error)
			w.WriteHeader(http.StatusOK)
			return
		}
		validatedHandler.ServeHTTP(w, r)
	})
}