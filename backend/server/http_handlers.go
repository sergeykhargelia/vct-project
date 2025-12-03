package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sergeykhargelia/vct-project/backend/model"
)

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.DB.Create(&user).Error; err != nil {
		http.Error(w, "Error while inserting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user_id, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	result := s.DB.Where("id = ?", user_id).Delete(&model.User{})
	if result.Error != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user_id, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user.ID = user_id

	err = s.DB.Save(&user).Error
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) CreateRegularExpense(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var regularExpense model.RegularExpense
	if err := json.NewDecoder(r.Body).Decode(&regularExpense); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.DB.Create(&regularExpense).Error; err != nil {
		http.Error(w, "Error while inserting regular expense", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(regularExpense)
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
	user_id, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	var regularExpenses []model.RegularExpense
	if s.DB.Where("user_id = ? AND next_date IS NOT NULL", user_id).Find(&regularExpenses).Error != nil {
		http.Error(w, "Error while finding regular expenses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(regularExpenses)
}

func (s *Server) GetUserExpenses(w http.ResponseWriter, r *http.Request) {
	user_id, err := strconv.ParseUint(mux.Vars(r)["user_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
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
	if s.DB.Where("user_id = ? AND date >= ? AND date <= ?", user_id, startDate, endDate).Find(&expenses).Error != nil {
		http.Error(w, "Error while finding expenses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expenses)
}
