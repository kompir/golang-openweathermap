package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	env "github.com/kompir/golang-openweathermap/database"
	app2 "github.com/kompir/golang-openweathermap/internal/app"
	http2 "github.com/kompir/golang-openweathermap/internal/http"
	storage2 "github.com/kompir/golang-openweathermap/internal/storage"
	"github.com/kompir/golang-openweathermap/internal/vault"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {

	conf, err := vault.New()

	//Connection To Database
	db, _ := dbConnection(conf)
	defer db.Close()

	storage := storage2.NewStorage(db)
	app := app2.NewApp(storage)
	httpHandler := http2.NewHttpHandler(app)

	//Migrations If True
	boolValue, err := strconv.ParseBool(env.ViperEnvVariable("DB_MIGRATE"))
	if err != nil {
		log.Fatal(err)
	}
	if boolValue == true {
		env.LoadSQLFile(db, "database/migration.sql")
	}

	//APi
	r := mux.NewRouter()
	fileServer := http.FileServer(http.Dir("./web/templates"))
	r.Handle("/", fileServer).Methods("GET")
	r.HandleFunc("/index", httpHandler.IndexHandler).Methods("GET")
	r.HandleFunc("/min/{days:[0-9]+}", httpHandler.Min).Methods("GET")
	r.HandleFunc("/max/{days:[0-9]+}", httpHandler.Max).Methods("GET")
	r.HandleFunc("/average/{days:[0-9]+}", httpHandler.Average).Methods("GET")

	//Meteo Server
	go cron(db)

	fmt.Printf("Starting server at port 8008\n")
	if err := http.ListenAndServe(":8008", r); err != nil {
		log.Fatal(err)
	}
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

func cron(db *sql.DB) {

	var interval int
	switch env.ViperEnvVariable("TIME_INTERVAL") {
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

func getWeather() (*OpenWheatherMap, error) {
	w := &OpenWheatherMap{
		Settings: NewSettings(),
		Unit:     env.ViperEnvVariable("UNITS"),
		Lang:     env.ViperEnvVariable("LANG"),
		Key:      env.ViperEnvVariable("OWM_KEY"),
	}
	resp := getWeatherAPI(w)
	err := json.Unmarshal(resp, w)
	if err != nil {
		panic(err)
	}
	return w, nil
}

func getWeatherAPI(w *OpenWheatherMap) []byte {
	response, err := w.client.Get(fmt.Sprintf(fmt.Sprintf(fmt.Sprint(env.ViperEnvVariable("BASE_OWM_URL")), "appid=%s&q=%s&units=%s&lang=%s"), w.Key, url.QueryEscape(env.ViperEnvVariable("LOCATION")), w.Unit, w.Lang))
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	resp, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return resp
}

func dbConnection(conf *vault.Provider) (db *sql.DB, err error) {
	get := func(v string) string {
		res, err := conf.Get(v)
		if err != nil {
			log.Fatalf("Couldn't get configuration value for %s: %s", v, err)
		}

		return res
	}
	dbHost := env.ViperEnvVariable("DB_HOST")
	dbPort := env.ViperEnvVariable("DB_PORT")

	dbUsername := get(env.ViperEnvVariable("DB_USERNAME"))
	dbPassword := get(env.ViperEnvVariable("DB_PASSWORD"))
	dbDriver := env.ViperEnvVariable("DB_DRIVER")

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return sql.Open(dbDriver, dbUsername+":"+dbPassword+"@tcp(db:"+dbPort+")/")
	}
	return sql.Open(dbDriver, dbUsername+":"+dbPassword+"@tcp("+dbHost+":"+dbPort+")/")
}
