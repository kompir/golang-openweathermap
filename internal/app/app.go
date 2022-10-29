package app

import (
	"github.com/kompir/golang-openweathermap/internal/storage"
)

type History struct {
	City string
	Temp float32
	Date string
	Days int
}

type AppStorage struct {
	DB storage.StorageI
}

func NewApp(db storage.StorageI) *AppStorage {
	return &AppStorage{DB: db}
}

func (a *AppStorage) Min(days int) (*History, error) {
	selDB, err := a.DB.Min(days)
	if err != nil {
		panic(err.Error())
	}
	history := &History{}
	for selDB.Next() {
		var main_temp float32
		var city_name string
		err = selDB.Scan(&main_temp, &city_name)
		if err != nil {
			return nil, err
		}
		history.Days = days
		history.Temp = main_temp
		history.City = city_name
	}

	return history, nil
}

func (a *AppStorage) Max(days int) (*History, error) {
	selDB, err := a.DB.Max(days)
	if err != nil {
		panic(err.Error())
	}
	history := &History{}
	for selDB.Next() {
		var main_temp float32
		var city_name string
		err = selDB.Scan(&main_temp, &city_name)
		if err != nil {
			return nil, err
		}
		history.Days = days
		history.Temp = main_temp
		history.City = city_name
	}

	return history, nil
}

func (a *AppStorage) Average(days int) (*History, error) {
	selDB, err := a.DB.Average(days)
	if err != nil {
		panic(err.Error())
	}
	history := &History{}
	for selDB.Next() {
		var main_temp float32
		var city_name string
		err = selDB.Scan(&main_temp, &city_name)
		if err != nil {
			return nil, err
		}
		history.Days = days
		history.Temp = main_temp
		history.City = city_name
	}

	return history, nil
}
