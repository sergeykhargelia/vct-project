package server_test

/*import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/sergeykhargelia/vct-project/backend/model"
	"github.com/sergeykhargelia/vct-project/backend/server"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initTestServer() (*server.Server, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return nil, nil, err
	}

	s := &server.Server{DB: gormDB, Metrics: server.InitMetrics(false)}
	return s, mock, nil
}

func TestCreate(t *testing.T) {
	s, mock, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	user := model.User{Email: "test@gmail.com", Name: "X", PasswordHash: "111"}
	userJson, _ := json.Marshal(user)

	nextDate := "2026-01-01"
	regularExpense := model.RegularExpense{
		UserID:    1,
		Name:      "yandex-plus",
		NextDate:  &nextDate,
		Frequency: "1 month",
		Amount:    700,
	}
	regularExpenseJson, _ := json.Marshal(regularExpense)

	type testParameters struct {
		target         string
		queryArguments []driver.Value
		payload        []byte
		f              func(w http.ResponseWriter, r *http.Request)
	}

	table := []testParameters{
		{
			"/users",
			[]driver.Value{user.Email, user.Name, user.PasswordHash},
			userJson, s.CreateUser,
		},
		{
			"/regular_expenses",
			[]driver.Value{
				regularExpense.UserID,
				regularExpense.Name,
				regularExpense.Description,
				regularExpense.NextDate,
				regularExpense.Frequency,
				regularExpense.Amount,
			},
			regularExpenseJson,
			s.CreateRegularExpense,
		},
	}

	for _, params := range table {
		req := httptest.NewRequest("POST", params.target, bytes.NewBuffer(params.payload))
		w := httptest.NewRecorder()

		mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%s" .* RETURNING "id"`, params.target[1:])).
			WithArgs(params.queryArguments...).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		params.f(w, req)
		assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
	}
}

func TestDeleteUser(t *testing.T) {
	s, mock, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	type responseParameters struct {
		expectedResult     driver.Result
		expectedStatusCode int
	}

	table := []responseParameters{
		{sqlmock.NewResult(0, 1), http.StatusNoContent},
		{sqlmock.NewResult(0, 0), http.StatusNotFound},
	}

	for _, params := range table {
		req := httptest.NewRequest("DELETE", "/users/1", nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
		w := httptest.NewRecorder()

		mock.ExpectExec(`DELETE FROM "users" WHERE id = \$1`).
			WithArgs(uint64(1)).
			WillReturnResult(params.expectedResult)

		s.DeleteUser(w, req)
		assert.Equal(t, params.expectedStatusCode, w.Result().StatusCode)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

func TestUpdateUser(t *testing.T) {
	s, mock, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	user := model.User{ID: 1, Email: "test@gmail.com", Name: "X", PasswordHash: "111"}
	userJson, _ := json.Marshal(user)
	req := httptest.NewRequest("PUT", "/users/1", bytes.NewBuffer(userJson))
	req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
	w := httptest.NewRecorder()

	mock.ExpectExec(`UPDATE "users" SET .* "id" = \$\d+`).
		WithArgs(user.Email, user.Name, user.PasswordHash, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.UpdateUser(w, req)

	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDeleteRegularExpense(t *testing.T) {
	s, mock, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("DELETE", "/regular_expenses/1", nil)
	req = mux.SetURLVars(req, map[string]string{"regular_expense_id": "1"})
	w := httptest.NewRecorder()

	mock.ExpectExec(`UPDATE "regular_expenses" SET "next_date"=\$1 WHERE id = \$2`).
		WithArgs(nil, uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.DeleteRegularExpense(w, req)

	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetUserRegularExpenses(t *testing.T) {
	s, mock, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/users/1/regular_expenses", nil)
	req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
	w := httptest.NewRecorder()

	mock.ExpectQuery(`SELECT \* FROM "regular_expenses" WHERE user_id = \$1 AND next_date IS NOT NULL`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(nil))

	s.GetUserRegularExpenses(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetUserExpenses(t *testing.T) {
	s, mock, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/users/1/expenses?start_date=2026-01-01&end_date=2026-01-01", nil)
	req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
	w := httptest.NewRecorder()

	mock.ExpectQuery(`SELECT \* FROM "expenses" WHERE user_id = \$1 AND date >= \$2 AND date <= \$3`).
		WithArgs(1, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(nil))

	s.GetUserExpenses(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetUserExpensesInvalidArguments(t *testing.T) {
	s, _, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	targets := []string{
		"/users/abc/expenses",
		"/users/1/expenses?start_date=1",
		"/users/1/expenses?start_date=2026-01-01&end_date=1",
		"/users/1/expenses?start_date=2026-02-01&end_date=2026-01-01",
	}

	for _, target := range targets {
		req := httptest.NewRequest("GET", target, nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": strings.Split(target, "/")[1]})
		w := httptest.NewRecorder()
		s.GetUserExpenses(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	}
}

type queryParameters struct {
	method string
	target string
	f      func(w http.ResponseWriter, r *http.Request)
}

func TestQueryInvalidUser(t *testing.T) {
	s, _, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	table := []queryParameters{
		{"DELETE", "", s.DeleteUser},
		{"PUT", "", s.UpdateUser},
		{"GET", "/regular_expenses", s.GetUserRegularExpenses},
		{"GET", "/expenses", s.GetUserExpenses},
	}

	for _, params := range table {
		req := httptest.NewRequest(params.method, "/users/abc"+params.target, nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": "abc"})
		w := httptest.NewRecorder()

		params.f(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestInvalidJSON(t *testing.T) {
	s, _, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	table := []queryParameters{
		{"POST", "/users", s.DeleteUser},
		{"PUT", "/users/1", s.UpdateUser},
		{"POST", "/regular_expenses", s.CreateRegularExpense},
	}

	badJson := `{"id": "1"}`

	for _, params := range table {
		req := httptest.NewRequest(params.method, params.target, bytes.NewBuffer([]byte(badJson)))
		if params.method == "PUT" {
			req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
		}
		w := httptest.NewRecorder()

		params.f(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestDBError(t *testing.T) {
	s, _, err := initTestServer()
	if err != nil {
		t.Fatal(err)
	}

	s.DB.Logger = logger.Default.LogMode(logger.Silent)

	table := []queryParameters{
		{"POST", "/users", s.CreateUser},
		{"DELETE", "/users/1", s.DeleteUser},
		{"PUT", "/users/1", s.UpdateUser},
		{"POST", "/regular_expenses", s.CreateRegularExpense},
		{"DELETE", "/regular_expenses/1", s.DeleteRegularExpense},
		{"GET", "/users/1/regular_expenses", s.GetUserRegularExpenses},
		{"GET", "/users/1/expenses?start_date=2026-01-01&end_date=2026-01-01", s.GetUserExpenses},
	}

	json := []byte(`{"id": 1}`)

	for _, params := range table {
		req := httptest.NewRequest(params.method, params.target, bytes.NewBuffer(json))
		req = mux.SetURLVars(req, map[string]string{
			"user_id":            "1",
			"regular_expense_id": "1",
		})
		w := httptest.NewRecorder()

		params.f(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	}
}*/
