package main

import (
	"context"
	"fmt"
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
	go ss.Run()

	// HTTP API сервиса
	srv := &http.Server{}
	http.Handle("/subscribe",  ValidSubscribeHandler(env.confirmSubscriber(http.HandlerFunc(env.SubscribeHandle))))
	log.Println("Started to listen and serve on :8181")
	ln, _ := net.Listen("tcp", ":8181")
	go func() {
		log.Fatalln(srv.Serve(ln))
	}()
	//go log.Fatalln(http.ListenAndServe(":8181", srv.Handler))

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

// Метод подписки на объявление
func (env *Env)SubscribeHandle(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	mail := r.URL.Query().Get("mail")

	env.SubService.AddSubscriberToProduct(subservice.TrimProductLink(link), mail)
}

// ----------------------------------------------------- MiddleWare for SubscribeHandle

// Валидация входных параметров link, mail
func ValidSubscribeHandler(subHandle http.Handler) http.Handler {
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

// Проверка кода подтверждения, в случае отсутствия - отправка письма с кодом
func (env *Env)confirmSubscriber(validatedHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		link, mail := r.URL.Query().Get("link"), r.URL.Query().Get("mail")
		confirmCode := r.URL.Query().Get("code")
		switch {
		case confirmCode == "":
			env.SubService.SendConfirmationEmail(link, mail)
			fmt.Fprintf(w, "Please, confirm you subscription using link from email sent to you by our service!")
			return
		case confirmCode != env.SubService.ConfirmCode:
			fmt.Fprintf(w, "Please, confirm you subscription using link from email sent to you by our service!")
			return
		}
		validatedHandler.ServeHTTP(w, r)
	})
}