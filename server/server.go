package server

import (
	"fmt"
	"net/http"

	"github.com/sergeykhargelia/vct-project/model"
	"github.com/sergeykhargelia/vct-project/templates"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type Metrics struct {
}

func InitMetrics(registerInPrometheus bool) *Metrics {
	//
	return &Metrics{}
}

type Server struct {
	DB          *gorm.DB
	Metrics     *Metrics
	EmailSender *gomail.Dialer
}

func (s *Server) MainPage(w http.ResponseWriter, r *http.Request) {
	templates.Dashboard().Render(r.Context(), w)
}

func (s *Server) DoRegularPayments(date string) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var updatedExpenses []model.RegularExpense
		err := tx.Raw(
			"UPDATE regular_expenses SET next_date = next_date + frequency WHERE next_date = ? RETURNING *",
			date,
		).Scan(&updatedExpenses).Error

		if err != nil {
			return err
		}

		var expenses []model.Expense
		for _, regularExpense := range updatedExpenses {
			expenses = append(expenses, model.Expense{
				UserID:           regularExpense.UserID,
				RegularExpenseID: regularExpense.ID,
				Date:             date,
			})
		}

		return s.DB.Create(&expenses).Error
	})
}

func (s *Server) NotifyAboutRegularPayments(date string) error {
	var regularExpenses []model.RegularExpense
	err := s.DB.Preload("User").Where("next_date = ?", date).Find(&regularExpenses).Error

	if err != nil {
		return err
	}

	for _, e := range regularExpenses {
		msg := gomail.NewMessage()
		msg.SetHeader("From", s.EmailSender.Username)
		msg.SetHeader("To", e.User.Email)
		msg.SetHeader("Subject", "Regular expense is coming")
		msg.SetBody("text/html", fmt.Sprintf(
			"Dear %s! Please, don't forget about your %s payment of %d rubles, it will be tomorrow.\n",
			e.User.Name,
			e.Name,
			e.Amount,
		))

		if err := s.EmailSender.DialAndSend(msg); err != nil {
			return err
		}
	}

	return nil
}
