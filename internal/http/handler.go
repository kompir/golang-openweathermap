package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kompir/golang-openweathermap/internal/app"
	"html/template"
	"net/http"
	"strconv"
)

type HttpHandler struct {
	app *app.App
}

func NewHttpHandler(app *app.App) *HttpHandler {
	return &HttpHandler{
		app: app,
	}
}

func (h *HttpHandler) Min(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	days, _ := strconv.Atoi(vars["days"])

	history, err := h.app.Min(days)

	t, err := template.ParseFiles("web/templates/min.html")
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
		return
	}
	t.Execute(w, history)
}

func (h *HttpHandler) Max(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	days, _ := strconv.Atoi(vars["days"])

	history, err := h.app.Max(days)
	t, err := template.ParseFiles("web/templates/min.html")
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
		return
	}
	t.Execute(w, history)
}

func (h *HttpHandler) Average(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	days, _ := strconv.Atoi(vars["days"])

	history, err := h.app.Average(days)
	t, err := template.ParseFiles("web/templates/min.html")
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
		return
	}
	t.Execute(w, history)
}

func (h *HttpHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/index" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Index!")
}
