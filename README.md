# Golang OpenWeatherMap
****

[Github Repository](https://github.com/kompir/golang-openweathermap)


## Features

### Save The Weather Temperature For The Last Day, Of The Given City In .env file, In History. It Checks Over A Period Of Time, Set In .env (Minute, Hour, Day) The [Openweathermap API](https://openweathermap.org/api) For Data. 
### Provide Access To History Through Rest API On Given Endpoints For Minimum, Maximum And Average Temperature For The Last “N” Days

- http://localhost:8008/min/1
- http://localhost:8008/max/1
- http://localhost:8008/average/1

## Standalone Installation
### go mod tidy
Adds any missing module requirements necessary to build the current module’s packages and dependencies.
### .env
Copy .env.example as .env with corresponding database credentials and set DB_MIGRATE=FALSE after first run.
### go run main.go
Runs the application
### go build
Builds the application
### ./golang-openweathermap 
Executes builded application

## Containerization 
### docker compose build --no-cache
Building the image without cache
### docker compose up
Builds, (re)creates, starts, and attaches to containers for a service.

## Tests
go test ./...