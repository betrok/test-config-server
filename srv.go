package main

import (
	"net/http"
)

func configHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`"wat?"`))
}
