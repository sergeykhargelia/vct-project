package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"github.com/sergeykhargelia/vct-project/backend/database"
	"github.com/sergeykhargelia/vct-project/backend/server"
	"gopkg.in/gomail.v2"
)

func setupDailyRoutine(s *server.Server) {
	c := cron.New()
	c.AddFunc("0 0 * * * ", func() {
		currentDate := time.Now()
		s.DoRegularPayments(currentDate.Format(time.DateOnly))

		nextDate := currentDate.AddDate(0, 0, 1)
		s.NotifyAboutRegularPayments(nextDate.Format(time.DateOnly))
	})
}

func initEmailSender() *gomail.Dialer {
	username := os.Getenv("GMAIL_USERNAME")
	password := os.Getenv("GMAIL_PASSWORD")
	return gomail.NewDialer("smtp.gmail.com", 587, username, password)
}

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	server := &server.Server{DB: db, EmailSender: initEmailSender()}
	setupDailyRoutine(server)

	router := mux.NewRouter()
	router.HandleFunc("/regular_expenses", server.CreateRegularExpense).Methods(http.MethodPost)
	router.HandleFunc("/regular_expenses/{regular_expense_id}", server.DeleteRegularExpense).Methods(http.MethodDelete)
	router.HandleFunc("/users", server.CreateUser).Methods(http.MethodPost)

	userRouter := router.PathPrefix("/users/{user_id}").Subrouter()
	userRouter.HandleFunc("", server.UpdateUser).Methods(http.MethodPut)
	userRouter.HandleFunc("", server.DeleteUser).Methods(http.MethodDelete)
	userRouter.HandleFunc("/regular_expenses", server.GetUserRegularExpenses).Methods(http.MethodGet)
	userRouter.HandleFunc("/expenses", server.GetUserExpenses).Methods(http.MethodGet)

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", router))
}
