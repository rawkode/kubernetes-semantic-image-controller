package main

import (
	"log"
	"net/http"

	"gitlab.com/rawkode/kubernetes-semantic-version/pkg/controller"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", controller.HandleMutate)
	mux.HandleFunc("/health", controller.HandleHealth)
	srv := &http.Server{Addr: ":443", Handler: mux}
	log.Fatal(srv.ListenAndServeTLS("/certs/webhook.crt", "/certs/webhook-key.pem"))
}
