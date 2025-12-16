package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sergeykhargelia/vct-project/model"
	"github.com/sergeykhargelia/vct-project/templates"
)

func (s *Server) CreateRegularExpense(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		templates.ErrorMessage("Failed to parse user id from request context").Render(r.Context(), w)
		return
	}

	nextDate := r.PostFormValue("nextDate")
	amount, err := strconv.ParseUint(r.PostFormValue("amount"), 10, 32)

	if err != nil {
		templates.ErrorMessage("Failed to parse amount")
		return
	}

	regularExpense := model.RegularExpense{
		UserID:      userID,
		Name:        r.PostFormValue("name"),
		Description: r.PostFormValue("description"),
		NextDate:    &nextDate,
		Frequency:   r.PostFormValue("frequency"),
		Amount:      uint(amount),
	}

	if err := s.DB.Create(&regularExpense).Error; err != nil {
		templates.ErrorMessage("Error while creating regular expense record")
		return
	}

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) DeleteRegularExpense(w http.ResponseWriter, r *http.Request) {
	regular_expense_id, err := strconv.ParseUint(mux.Vars(r)["regular_expense_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid regular expense id", http.StatusBadRequest)
		return
	}

	result := s.DB.Model(&model.RegularExpense{}).Where("id = ?", regular_expense_id).Update("next_date", nil)
	if result.Error != nil {
		http.Error(w, "Failed to delete regular expense", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "Regular expense not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetUserRegularExpenses(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		templates.ErrorMessage("Failed to parse user id from request context").Render(r.Context(), w)
		return
	}

	var regularExpenses []model.RegularExpense
	if s.DB.Where("user_id = ? AND next_date IS NOT NULL", userID).Find(&regularExpenses).Error != nil {
		templates.ErrorMessage("Error while finding regular expenses")
		return
	}

	templates.ExpensesList(regularExpenses).Render(r.Context(), w)
}

func (s *Server) GetUserExpenses(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uint64)
	if !ok {
		http.Error(w, "Failed to parse user id from request context", http.StatusUnauthorized)
		return
	}

	startDate, err := time.Parse(time.DateOnly, r.URL.Query().Get("start_date"))
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(time.DateOnly, r.URL.Query().Get("end_date"))
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	if endDate.Before(startDate) {
		http.Error(w, "End date must be after start date", http.StatusBadRequest)
		return
	}

	var expenses []model.Expense
	if s.DB.Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).Find(&expenses).Error != nil {
		http.Error(w, "Error while finding expenses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expenses)
}
