package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/marcsauter/buzzer/pkg/pitch"
	"github.com/mholt/binding"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

var (
	address, port, cache string
	mutex                = new(sync.Mutex)
	nextPitch            = pitch.Pitch{}
)

func init() {
	const (
		defaultAddress = "127.0.0.1"
		defaultPort    = "8080"
	)
	flag.StringVar(&address, "address", defaultAddress, "address")
	flag.StringVar(&port, "port", defaultPort, "port")
	flag.StringVar(&cache, "cache", fmt.Sprintf("/tmp/%s.cache", filepath.Base(os.Args[0])), "cache file")
}

func main() {
	flag.Parse()

	// setup basic authentication
	credentials := make(map[string][]string)
	username := os.Getenv("BUZZER_USERNAME")
	password := os.Getenv("BUZZER_PASSWORD")
	if len(username) != 0 {
		credentials[username] = []string{password}
	}

	// read next pitch from cache
	if _, err := os.Stat(cache); err == nil {
		if c, err := ioutil.ReadFile(cache); err == nil {
			if err := json.Unmarshal(c, &nextPitch); err != nil {
				log.Println("ERROR:", err)
			}
		} else {
			log.Println("ERROR:", err)
		}
	}

	api := chi.NewRouter()
	api.Use(basicAuth("buzzer", credentials))
	api.Post("/next", func(w http.ResponseWriter, r *http.Request) {
		p := pitch.Pitch{}
		if errs := binding.Bind(r, &p); errs.Handle(w) {
			return
		}
		// update next pitch if the pitch id changed
		if p.ID != nextPitch.ID {
			mutex.Lock()
			if p.ID != nextPitch.ID {
				nextPitch = p
				// write cache
				if data, err := json.Marshal(nextPitch); err == nil {
					if err := ioutil.WriteFile(cache, data, 0600); err != nil {
						log.Println("ERROR:", err)
					}
				} else {
					log.Println("ERROR:", err)
				}
			}
			mutex.Unlock()
		}
		//log.Printf("next pitch: \"%s\" talks about \"%s\" on \"%s\"", nextPitch.Speaker, nextPitch.Title, nextPitch.Date)
	})
	api.Get("/next", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		p := nextPitch
		mutex.Unlock()
		render.JSON(w, r, p)
	})

	// migration endpoints
	// have to exist but do nothing
	api.Get("/", func(w http.ResponseWriter, r *http.Request) {})

	log.Printf("server is listening on %s:%s", address, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), api))
}

func basicAuth(realm string, credentials map[string][]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				unauthorized(w, realm)
				return
			}

			validPasswords, userFound := credentials[username]
			if !userFound {
				unauthorized(w, realm)
				return
			}

			for _, validPassword := range validPasswords {
				if password == validPassword {
					next.ServeHTTP(w, r)
					return
				}
			}

			unauthorized(w, realm)
		})
	}
}

func unauthorized(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
