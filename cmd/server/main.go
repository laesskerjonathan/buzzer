package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/marcsauter/buzzer/pkg/pitch"
	"github.com/mholt/binding"
	"github.com/unrolled/render"
)

var (
	address, port string
	mutex         = new(sync.Mutex)
	nextPitch     = pitch.Pitch{}
)

func init() {
	const (
		defaultAddress = "127.0.0.1"
		defaultPort    = "8080"
	)
	flag.StringVar(&address, "address", defaultAddress, "address")
	flag.StringVar(&port, "port", defaultPort, "port")
}

func main() {
	flag.Parse()
	api := mux.NewRouter().StrictSlash(true)
	api.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		p := &pitch.Pitch{}
		if errs := binding.Bind(r, p); errs.Handle(w) {
			return
		}
		mutex.Lock()
		nextPitch = *p
		mutex.Unlock()
		//log.Printf("next pitch: \"%s\" talks about \"%s\" on \"%s\"", nextPitch.Speaker, nextPitch.Title, nextPitch.Date)
		render.New().JSON(w, http.StatusOK, nil)
	}).Methods("POST")
	api.HandleFunc("/next", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		p := nextPitch
		mutex.Unlock()
		render.New().JSON(w, http.StatusOK, p)
	}).Methods("GET")

	// migration endpoints
	// have to exist but do nothing
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	log.Printf("server is listening on %s:%s", address, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), api))
}
