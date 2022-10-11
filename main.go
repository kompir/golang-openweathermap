package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {

	//Connection To Database
	db, _ := dbConnection()
	defer db.Close()

	//Migrations If True
	boolValue, err := strconv.ParseBool(viperEnvVariable("DB_MIGRATE"))
	if err != nil {
		log.Fatal(err)
	}
	if boolValue == true {
		migrate(db)
	}

	//APi
	r := mux.NewRouter()
	fileServer := http.FileServer(http.Dir("./web/templates"))
	r.Handle("/", fileServer).Methods("GET")
	r.HandleFunc("/hello", indexHandler).Methods("GET")
	//r.HandleFunc("/insert", insert).Methods("POST")
	r.HandleFunc("/min/{days:[0-9]+}", min).Methods("GET")
	r.HandleFunc("/max/{days:[0-9]+}", max).Methods("GET")
	r.HandleFunc("/average/{days:[0-9]+}", average).Methods("GET")

	//Meteo Server
	go cron(db)

	fmt.Printf("Starting server at port 8008\n")
	if err := http.ListenAndServe(":8008", r); err != nil {
		log.Fatal(err)
	}

}

// use viper package to read .env file
// return the value of the key
func viperEnvVariable(key string) string {
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

type Main struct {
	Temp    float64 `json:"temp"`
	TempMin float64 `json:"temp_min"`
	TempMax float64 `json:"temp_max"`
}

type OpenWheatherMap struct {
	Main Main   `json:"main"`
	Dt   int64  `json:"dt"`
	Name string `json:"name"`
	Unit string
	Lang string
	Key  string
	*Settings
}

type Settings struct {
	client *http.Client
}

func NewSettings() *Settings {
	return &Settings{
		client: http.DefaultClient,
	}
}

func getWeather() (*OpenWheatherMap, error) {
	w := &OpenWheatherMap{
		Settings: NewSettings(),
		Unit:     viperEnvVariable("UNITS"),
		Lang:     viperEnvVariable("LANG"),
		Key:      viperEnvVariable("OWM_KEY"),
	}
	response, err := w.client.Get(fmt.Sprintf(fmt.Sprintf(fmt.Sprint(viperEnvVariable("BASE_OWM_URL")), "appid=%s&q=%s&units=%s&lang=%s"), w.Key, url.QueryEscape(viperEnvVariable("LOCATION")), w.Unit, w.Lang))
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&w); err != nil {
		panic(err)
	}
	return w, nil
}

func cron(db *sql.DB) {

	var interval int
	switch viperEnvVariable("TIME_INTERVAL") {
	case "Minute":
		interval = 1
	case "Hour":
		interval = 60
	case "Day":
		interval = 1440
	}

	for {
		select {
		case <-time.After(time.Duration(interval) * time.Minute):
			fmt.Println("Cron is working !")
			wd, err := getWeather()
			if err != nil {
				fmt.Println("Cron Error")
				return
			}
			mainTemp := wd.Main.Temp
			date := time.Unix(wd.Dt, 0)
			cityName := wd.Name
			insForm, err := db.Prepare("INSERT INTO weather.meteo_table(city_name, main_temp, date) VALUES(?,?,?)")
			if err != nil {
				panic(err.Error())
			}
			insForm.Exec(cityName, mainTemp, date)
			fmt.Println("Data Inserted: ", mainTemp, date, cityName)
		}
	}
}

func dbConnection() (db *sql.DB, err error) {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return sql.Open(viperEnvVariable("DB_DRIVER"), viperEnvVariable("DB_USERNAME")+":"+viperEnvVariable("DB_PASSWORD")+"@tcp(db:"+viperEnvVariable("DB_PORT")+")/")
	}
	return sql.Open(viperEnvVariable("DB_DRIVER"), viperEnvVariable("DB_USERNAME")+":"+viperEnvVariable("DB_PASSWORD")+"@tcp("+viperEnvVariable("DB_HOST")+":"+viperEnvVariable("DB_PORT")+")/")
}

func migrate(db *sql.DB) {

	dbName := viperEnvVariable("DB_DATABASE")
	tableName := viperEnvVariable("DB_TABLE")

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
			Date: "2022-10-04",
		},
		{
			City: "Rousse",
			Temp: 11.00,
			Date: "2022-10-05",
		},
		{
			City: "Rousse",
			Temp: 12.00,
			Date: "2022-10-06",
		},
		{
			City: "Rousse",
			Temp: 13.00,
			Date: "2022-10-07",
		},
		{
			City: "Rousse",
			Temp: 14.00,
			Date: "2022-10-08",
		},
		{
			City: "Rousse",
			Temp: 15.00,
			Date: "2022-10-09",
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "GET method is not supported", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "hello!")
}

type History struct {
	City string
	Temp float32
	Date string
	Days int
}

func min(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	days, _ := strconv.Atoi(vars["days"])
	db, err := dbConnection()
	selDB, err := db.Query("select main_temp, city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY) ORDER BY `meteo_table`.`main_temp` ASC limit 1", days)
	if err != nil {
		panic(err.Error())
	}
	history := History{}
	for selDB.Next() {
		var main_temp float32
		var city_name string
		err = selDB.Scan(&main_temp, &city_name)
		if err != nil {
			panic(err.Error())
		}
		history.Days = days
		history.Temp = main_temp
		history.City = city_name
	}
	t, err := template.ParseFiles("web/templates/min.html")
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
		return
	}
	t.Execute(w, history)

}

func max(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	days, _ := strconv.Atoi(vars["days"])
	db, err := dbConnection()
	selDB, err := db.Query("select main_temp, city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY) ORDER BY `meteo_table`.`main_temp` DESC limit 1", days)
	if err != nil {
		panic(err.Error())
	}
	history := History{}
	for selDB.Next() {
		var main_temp float32
		var city_name string
		err = selDB.Scan(&main_temp, &city_name)
		if err != nil {
			panic(err.Error())
		}
		history.Days = days
		history.Temp = main_temp
		history.City = city_name
	}
	t, err := template.ParseFiles("web/templates/max.html")
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
		return
	}
	t.Execute(w, history)
}

func average(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	days, _ := strconv.Atoi(vars["days"])
	db, err := dbConnection()
	selDB, err := db.Query("select ROUND(AVG(main_temp), 2), city_name from weather.meteo_table WHERE date > DATE_ADD(CURDATE(), INTERVAL -? DAY)", days)
	if err != nil {
		panic(err.Error())
	}
	history := History{}
	for selDB.Next() {
		var main_temp float32
		var city_name string
		err = selDB.Scan(&main_temp, &city_name)
		if err != nil {
			panic(err.Error())
		}
		history.Days = days
		history.Temp = main_temp
		history.City = city_name
	}
	t, err := template.ParseFiles("web/templates/average.html")
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
		return
	}
	t.Execute(w, history)
}
