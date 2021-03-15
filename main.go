package main

import (
	"log"
	"net/http"

	"example.com/hungry-server/controller"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/register", controller.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", controller.LoginHandler).Methods("POST")
	router.HandleFunc("/profile", controller.ProfileHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
