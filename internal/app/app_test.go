package app_test

import (
	"github.com/kompir/golang-openweathermap/internal/app"
	storage2 "github.com/kompir/golang-openweathermap/internal/storage"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func prepareAppMin(t *testing.T, days int) *app.App {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a database connection", err)
	}
	storage := storage2.NewStorage(db)
	rows := sqlmock.NewRows([]string{"main_temp", "city_name"}).
		AddRow(14, "Rousse")
	mock.ExpectQuery(regexp.QuoteMeta("select main_temp, city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY) ORDER BY main_temp ASC limit 1")).
		WithArgs(3).
		WillReturnRows(rows)

	return &app.App{
		DB: storage,
	}
}

func prepareAppMax(t *testing.T, days int) *app.App {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a database connection", err)
	}
	storage := storage2.NewStorage(db)
	rows := sqlmock.NewRows([]string{"main_temp", "city_name"}).
		AddRow(15, "Rousse")
	mock.ExpectQuery(regexp.QuoteMeta("select main_temp, city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY) ORDER BY main_temp DESC limit 1")).
		WithArgs(3).
		WillReturnRows(rows)

	return &app.App{
		DB: storage,
	}
}

func prepareAppAverage(t *testing.T, days int) *app.App {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a database connection", err)
	}
	storage := storage2.NewStorage(db)
	rows := sqlmock.NewRows([]string{"main_temp", "city_name"}).
		AddRow(14.50, "Rousse")
	mock.ExpectQuery(regexp.QuoteMeta("select ROUND(AVG(main_temp), 2), city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY)")).
		WithArgs(3).
		WillReturnRows(rows)

	return &app.App{
		DB: storage,
	}
}

func TestApp_Min(t *testing.T) {
	appTest := prepareAppMin(t, 3)

	data, err := appTest.Min(3)
	if err != nil {
		t.Errorf("Min Method error is not nil")
	}
	if data.City != "Rousse" {
		t.Errorf("City Name Is Not Correct")
	}
	if data.Temp != 14 {
		t.Errorf("Min Temperature Is Not Correct")
	}
}

func TestApp_Max(t *testing.T) {
	appTest := prepareAppMax(t, 3)

	data, err := appTest.Max(3)
	if err != nil {
		t.Errorf("Max Method error is not nil")
	}
	if data.City != "Rousse" {
		t.Errorf("City Name Is Not Correct")
	}
	if data.Temp != 15 {
		t.Errorf("Max Temperature Is Not Correct")
	}
}

func TestApp_Average(t *testing.T) {
	appTest := prepareAppAverage(t, 3)

	data, err := appTest.Average(3)
	if err != nil {
		t.Errorf("Average Method error is not nil")
	}
	if data.City != "Rousse" {
		t.Errorf("City Name Is Not Correct")
	}
	if data.Temp != 14.50 {
		t.Errorf("Average Temperature Is Not Correct")
	}
}
