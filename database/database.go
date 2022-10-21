package database

import (
	"database/sql"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"strings"
)

type History struct {
	City string
	Temp float32
	Date string
	Days int
}

// using viper package to read .env file
// return the value of the key
func ViperEnvVariable(key string) string {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}
	value, ok := viper.Get(key).(string)
	if !ok {
		log.Fatalf("Invalid type assertion")
	}
	return value
}

// Database Migration & Seeder From File
func LoadSQLFile(db *sql.DB, sqlFile string) error {
	file, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		tx.Rollback()
	}()
	for _, q := range strings.Split(string(file), ";") {
		q := strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Database Migration & Seeder
func Migrate(db *sql.DB) {

	dbName := ViperEnvVariable("DB_DATABASE")
	tableName := ViperEnvVariable("DB_TABLE")

	_, err := db.Exec("DROP DATABASE IF EXISTS " + dbName)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("USE " + dbName)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE " + tableName + " ( id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, city_name VARCHAR(255), main_temp DECIMAL(10,2), date DATETIME )")
	if err != nil {
		panic(err)
	}

	seed := []History{
		{
			City: "Rousse",
			Temp: 10.00,
			Date: "2022-10-06",
		},
		{
			City: "Rousse",
			Temp: 11.00,
			Date: "2022-10-07",
		},
		{
			City: "Rousse",
			Temp: 12.00,
			Date: "2022-10-08",
		},
		{
			City: "Rousse",
			Temp: 13.00,
			Date: "2022-10-09",
		},
		{
			City: "Rousse",
			Temp: 14.00,
			Date: "2022-10-10",
		},
		{
			City: "Rousse",
			Temp: 15.00,
			Date: "2022-10-11",
		},
	}
	db.Exec("SET GLOBAL sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''));")
	for _, v := range seed {
		insForm, err := db.Prepare("INSERT INTO weather.meteo_table(city_name, main_temp, date) VALUES(?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(v.City, v.Temp, v.Date)
	}

}
