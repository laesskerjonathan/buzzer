package buzzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"text/template"
	"time"

	"golang.org/x/net/context"

	"gitlab.com/marcsauter/pitchbuzzer/device"
	"gitlab.com/marcsauter/pitchbuzzer/pitch"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/unrolled/render"
)

const (
	Kind     = "buzzer"
	StringId = "pitches"
)

// Error
type Error struct {
	Message string
}

// Entity
type Entity struct {
	Pitches pitch.Pitches
}

var (
	devices *device.Devices
	rd      *render.Render
)

func init() {
	devices = device.NewDevices()
	rd = render.New()
	http.HandleFunc("/", view)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/save", save)
	http.HandleFunc("/del", delHandler)
	http.HandleFunc("/remove", remove)
	http.HandleFunc("/start/", start)
	http.HandleFunc("/next", next)
	http.HandleFunc("/device", addDevice)
	http.HandleFunc("/devices", viewDevices)
}

// retrieve returns the records for a given device
func view(w http.ResponseWriter, r *http.Request) {
	p, err := getAllPitches(r)
	et, _ := template.ParseFiles("error.tmpl")
	if err != nil {
		et.Execute(w, err)
		return
	}
	t, _ := template.ParseFiles("view.tmpl")
	t.Execute(w, p)
	return
}

//
func addHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("add.tmpl")
	t.Execute(w, nil)
	return
}

// save a pitch
func save(w http.ResponseWriter, r *http.Request) {
	speaker := r.FormValue("speaker")
	title := r.FormValue("title")
	// validate
	et, _ := template.ParseFiles("error.tmpl")
	if len(speaker) == 0 || len(title) == 0 {
		et.Execute(w, errors.New("Speaker and Title are required"))
		return
	}
	// get the loc of the browser - but for now it's reasonable assumption
	loc, _ := time.LoadLocation("Europe/Zurich")
	date, err := time.ParseInLocation("02.01.2006 15:04", r.FormValue("date"), loc)
	if err != nil {
		et.Execute(w, err)
		return
	}
	if date.Before(time.Now()) {
		et.Execute(w, errors.New("Pitch is the past"))
		return
	}
	// save the pitch
	p := pitch.Pitch{
		Speaker: speaker,
		Title:   title,
		Date:    date,
	}
	if err := addPitch(r, &p); err != nil {
		et.Execute(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

//
func delHandler(w http.ResponseWriter, r *http.Request) {
	p, err := getAllPitches(r)
	if err != nil {
		et, _ := template.ParseFiles("error.tmpl")
		et.Execute(w, err)
		return
	}
	t, _ := template.ParseFiles("del.tmpl")
	t.Execute(w, p)
	return
}

// remove a pitch
func remove(w http.ResponseWriter, r *http.Request) {
	pid := r.FormValue("pitchid")
	if err := delPitch(r, pid); err != nil {
		et, _ := template.ParseFiles("error.tmpl")
		et.Execute(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

//
func start(w http.ResponseWriter, r *http.Request) {
	pid := r.URL.Path[len("/start/"):]
	if pid == "1982" {
		rd.JSON(w, http.StatusOK, nil)
		return
	}
	if err := startPitch(r, pid); err != nil {
		rd.JSON(w, http.StatusNotFound, Error{Message: err.Error()})
		return
	}
	// initiate ... whatever ...
	rd.JSON(w, http.StatusOK, nil)
	return
}

//
func next(w http.ResponseWriter, r *http.Request) {
	p, err := getNextPitch(r)
	if err != nil {
		rd.JSON(w, http.StatusNotFound, Error{Message: err.Error()})
		return
	}
	rd.JSON(w, http.StatusOK, p)
	return
}

func addDevice(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var d device.Device
	if err := decoder.Decode(&d); err != nil {
		rd.JSON(w, http.StatusInternalServerError, Error{Message: err.Error()})
		return
	}
	devices.Lock()
	devices.Items[d.Name] = d
	devices.Unlock()
	rd.JSON(w, http.StatusOK, d)
	return
}

func viewDevices(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("devices.tmpl")
	items := []device.Device{}
	for _, d := range devices.Items {
		items = append(items, d)
	}
	log.Println(items)
	t.Execute(w, items)
	return
}

//
func addPitch(r *http.Request, p *pitch.Pitch) error {
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, Kind, StringId, 0, nil)
	if err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var e Entity
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		err := datastore.Get(ctx, key, &e)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		// assign a unique pitch id
		pitchid := make(map[string]bool)
		for _, p := range e.Pitches {
			pitchid[p.Id] = true
		}
		id := time.Now().Year()*100 + 1
		for pitchid[fmt.Sprintf("%d", id)] {
			id += 1
		}
		p.Id = fmt.Sprintf("%d", id)
		// set date
		p.RegisteredAt = time.Now()
		// add pitch to pitches and save the list
		e.Pitches = append(e.Pitches, *p)
		sort.Sort(e.Pitches)
		if _, err = datastore.Put(ctx, key, &e); err != nil {
			return err
		}
		return nil
	}, nil); err != nil {
		return err
	}
	return nil
}

//
func getPitch(r *http.Request, pid string) (*pitch.Pitch, error) {
	var found *pitch.Pitch
	if len(pid) == 0 {
		return nil, errors.New("no pitch id given")
	}
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, Kind, StringId, 0, nil)
	if err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var e Entity
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		if err := datastore.Get(ctx, key, &e); err != nil {
			return err
		}
		for _, p := range e.Pitches {
			if p.Id == pid {
				found = &p
				return nil
			}
		}
		return errors.New(fmt.Sprintf("no pitch with id %s found", pid))
	}, nil); err != nil {
		return nil, err
	}
	return found, nil
}

//
func getAllPitches(r *http.Request) (*pitch.Pitches, error) {
	var pitches *pitch.Pitches
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, Kind, StringId, 0, nil)
	if err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var e Entity
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		if err := datastore.Get(ctx, key, &e); err != nil {
			return err
		}
		pitches = &e.Pitches
		return nil
	}, nil); err != nil {
		return nil, err
	}
	return pitches, nil
}

//
func getNextPitch(r *http.Request) (*pitch.Pitch, error) {
	var next *pitch.Pitch
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, Kind, StringId, 0, nil)
	if err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var e Entity
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		if err := datastore.Get(ctx, key, &e); err != nil {
			return err
		}
		next = &pitch.Pitch{}
		for _, p := range e.Pitches {
			if p.Date.After(time.Now()) {
				//if p.Released && p.Date.After(time.Now()) {
				next = &p
				return nil
			}
		}
		return nil
	}, nil); err != nil {
		return nil, err
	}
	return next, nil
}

//
func delPitch(r *http.Request, pid string) error {
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, Kind, StringId, 0, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var e Entity
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		if err := datastore.Get(ctx, key, &e); err != nil {
			return err
		}
		for i, p := range e.Pitches {
			if p.Id == pid {
				e.Pitches = append(e.Pitches[:i], e.Pitches[i+1:]...)
				if _, err := datastore.Put(ctx, key, &e); err != nil {
					return err
				}
				return nil
			}
		}
		return errors.New(fmt.Sprintf("no pitch with id %s found", pid))
	}, nil)
	return err
}

//
func startPitch(r *http.Request, pid string) error {
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, Kind, StringId, 0, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var e Entity
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		if err := datastore.Get(ctx, key, &e); err != nil {
			return err
		}
		for i, p := range e.Pitches {
			if p.Id == pid {
				if p.Released {
					return errors.New(fmt.Sprintf("pitch with id %s already released", p.Id))
				}
				// release pitch
				e.Pitches[i].Released = true
				e.Pitches[i].ReleasedAt = time.Now()
				// save data
				if _, err := datastore.Put(ctx, key, &e); err != nil {
					return err
				}
				return nil
			}
		}
		return errors.New(fmt.Sprintf("no pitch with id %s found", pid))
	}, nil)
	return err
}
