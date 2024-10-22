package services

import (
	"net/http"
	"time"

	"github.com/atinyakov/go_final_project/nextdate"
)

type NextDate struct {
}

func NewNextDateService() *NextDate {
	return &NextDate{}
}

func (n *NextDate) HandleNextDate(w http.ResponseWriter, req *http.Request) {
	repeat := req.FormValue("repeat")
	reqDate := req.FormValue("date")
	now, err := time.Parse("20060102", req.FormValue("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	date, err := nextdate.Get(now, reqDate, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(date))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
