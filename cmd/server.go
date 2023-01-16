package main

import "net/http"

func (app *application) latestScrap(w http.ResponseWriter, r *http.Request) {
	val := app.readDataFromCache("latest_scrap")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(val))
}

func (app *application) allScraps(w http.ResponseWriter, r *http.Request) {
	val := app.readDataFromCache("latest_scrap")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(val))
}

func (app *application) serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/latest", app.latestScrap)
	mux.HandleFunc("/all", app.allScraps)
	http.ListenAndServe(":3000", mux)
}
