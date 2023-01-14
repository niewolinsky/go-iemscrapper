package main

import "net/http"

func retrieveData(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello!"))
}

func (app *application) serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/data", retrieveData)
	http.ListenAndServe(":3000", mux)
}
