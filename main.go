package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"github.com/sergeykhargelia/vct-project/database"
	"github.com/sergeykhargelia/vct-project/server"
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

	s := &server.Server{DB: db, EmailSender: initEmailSender()}
	setupDailyRoutine(s)

	router := mux.NewRouter()
	router.HandleFunc("/register", s.RegisterHandler).Methods(http.MethodPost)
	router.HandleFunc("/login", s.LoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/register", s.RegisterPage).Methods(http.MethodGet)
	router.HandleFunc("/login", s.LoginPage).Methods(http.MethodGet)
	router.HandleFunc("/health", s.Health).Methods(http.MethodGet)

	router.HandleFunc("/", server.AuthMiddleware(s.MainPage)).Methods(http.MethodGet)
	router.HandleFunc("/regular_expenses", server.AuthMiddleware(s.CreateRegularExpense)).Methods(http.MethodPost)
	router.HandleFunc("/regular_expenses", server.AuthMiddleware(s.GetUserRegularExpenses)).Methods(http.MethodGet)
	router.HandleFunc("/regular_expenses/{regular_expense_id}", server.AuthMiddleware(s.DeleteRegularExpense)).Methods(http.MethodDelete)
	router.HandleFunc("/expenses", server.AuthMiddleware(s.GetUserExpenses)).Methods(http.MethodGet)

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", router))
}
