package pitch

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/marcsauter/pitchbuzzer/device"
)

//
type Updater interface {
	Update(data fmt.Stringer) error
}

// Pitch represents a pitch
type Pitch struct {
	Id           string    `json:"id"`
	Speaker      string    `json:"speaker"`
	Title        string    `json:"title"`
	Date         time.Time `json:"date"`
	RegisteredAt time.Time `json:"registeredat"`
	Released     bool      `json:"started"`
	ReleasedAt   time.Time `json:"startedat"`
	pitchUrl     *url.URL
	ticker       *time.Ticker
}

//
func (p *Pitch) FormattedDate() string {
	// get the loc of the browser - but for now it's reasonable assumption
	zrh, _ := time.LoadLocation("Europe/Zurich")
	return p.Date.In(zrh).Format("02.01.2006 15:04")
}

//
func (p *Pitch) FormattedReleasedAt() string {
	// get the loc of the browser - but for now it's reasonable assumption
	zrh, _ := time.LoadLocation("Europe/Zurich")
	if p.ReleasedAt.After(p.RegisteredAt) {
		return p.ReleasedAt.In(zrh).Format("02.01.2006 15:04")
	}
	return ""
}

// Pitches represents a slice of pitches
type Pitches []Pitch

//
func (p Pitches) Len() int {
	return len(p)
}

//
func (p Pitches) Less(i, j int) bool {
	return p[i].Date.Before(p[j].Date)
}

//
func (p Pitches) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

//
func NewPitch(u *url.URL) *Pitch {
	return &Pitch{
		pitchUrl: u,
	}
}

func (p *Pitch) String() string {
	return fmt.Sprintf("%s - %s - %s", p.Speaker, p.Title, p.Date.Format("02.01.2006 15:04"))
}

//
func (p *Pitch) StartCheckNext(interval int, out Updater) {
	//
	getNextPitch := func() Pitch {
		p.pitchUrl.Path = "next"
		resp, err := http.Get(p.pitchUrl.String())
		if err != nil {
			log.Print(err)
			return Pitch{}
		}
		if resp.StatusCode != 200 {
			log.Println(p.pitchUrl.String(), resp.Status)
			return Pitch{}
		}
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		var p Pitch
		if err := decoder.Decode(&p); err != nil {
			log.Print(err)
			return Pitch{}
		}
		return p
	}
	//
	go func() {
		p.ticker = time.NewTicker(time.Second * time.Duration(interval))
		for {
			if next := getNextPitch(); len(next.Id) > 0 && next.Id != p.Id {
				p.Id = next.Id
				p.Speaker = next.Speaker
				p.Title = next.Title
				p.Date = next.Date
				out.Update(p)
			}
			p.pitchUrl.Path = "device"
			device.Register(filepath.Base(os.Args[0]), p.pitchUrl.String())
			<-p.ticker.C
		}
	}()
}

//
func (p *Pitch) StopCheckNext() {
	p.ticker.Stop()
}
