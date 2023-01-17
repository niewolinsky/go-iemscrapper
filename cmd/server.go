package main

import "net/http"

func (app *application) latestScrap(w http.ResponseWriter, r *http.Request) {
	latest_scrap_data := app.getLatestScrapCache("latest_scrap")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(latest_scrap_data))
}

func (app *application) allScraps(w http.ResponseWriter, r *http.Request) {
	all_scrap_data := app.getAllScraps("iems")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(all_scrap_data))
}

func (app *application) serve() {
	mux := http.NewServeMux()

	mux.HandleFunc("/latest", app.latestScrap)
	mux.HandleFunc("/all", app.allScraps)

	http.ListenAndServe(":3000", mux)
}
