package pitch

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mholt/binding"
)

// Updater interface
type Updater interface {
	Stop() error
	Update(data fmt.Stringer) error
}

// Pitch represents a pitch
type Pitch struct {
	ID           string    `json:"id"`
	Speaker      string    `json:"speaker"`
	Title        string    `json:"title"`
	Date         time.Time `json:"date"`
	RegisteredAt time.Time `json:"registeredat"`
	Released     bool      `json:"started"`
	ReleasedAt   time.Time `json:"startedat"`
	pitchURL     *url.URL
	ticker       *time.Ticker
}

// FieldMap implements the FieldMapper interface for github.com/mholt/binding
// these are the only vital fields
func (p *Pitch) FieldMap(r *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&p.ID:      "id",
		&p.Speaker: "speaker",
		&p.Title:   "title",
		&p.Date:    "date",
	}
}

// FormattedDate returns the formatted Date
// TODO: remove if buzzer-ws is no longer in use
func (p *Pitch) FormattedDate() string {
	// get the loc of the browser - but for now it's reasonable assumption
	zrh, _ := time.LoadLocation("Europe/Zurich")
	return p.Date.In(zrh).Format("02.01.2006 15:04")
}

// FormattedReleasedAt returns the formatted ReleasedAt
// TODO: remove if buzzer-ws is no longer in use
func (p *Pitch) FormattedReleasedAt() string {
	// get the loc of the browser - but for now it's reasonable assumption
	zrh, _ := time.LoadLocation("Europe/Zurich")
	if p.ReleasedAt.After(p.RegisteredAt) {
		return p.ReleasedAt.In(zrh).Format("02.01.2006 15:04")
	}
	return ""
}

// NewPitch returns a new Pitch instance
func NewPitch(u *url.URL) *Pitch {
	return &Pitch{
		pitchURL: u,
	}
}

// Pitch has to fullfill the Stringer interface - see also Updater interface
func (p *Pitch) String() string {

	text := fmt.Sprintf("In %d Minuten Pitch im PFLab: %s von %s", int(time.Until(p.Date).Minutes()), p.Title, p.Speaker)
	return text
}

// StartCheckNext start the checker for the next pitch and updates Pitch if something changes
func (p *Pitch) StartCheckNext(interval int, out Updater) {
	//
	getNextPitch := func() Pitch {
		p.pitchURL.Path = "next"
		resp, err := http.Get(p.pitchURL.String())
		if err != nil {
			log.Print(err)
			return Pitch{}
		}
		if resp.StatusCode != 200 {
			log.Println(p.pitchURL.String(), resp.Status)
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
			next := getNextPitch()
			
			zrh, _ := time.LoadLocation("Europe/Zurich")
			pitchDate := next.Date.In(zrh)
			minutesUntilNextPitch := int(time.Until(pitchDate).Minutes());
			
			
			logText := fmt.Sprintf("ID: %s, Speaker: %s, Title: %s, Date: %s", next.ID, next.Speaker, next.Title, pitchDate)
			
			if len(next.ID) > 0 {
				p.ID = next.ID
				p.Speaker = next.Speaker
				p.Title = next.Title
				p.Date = next.Date
				out.Update(p)
			}
			
			if len(next.ID) <= 0 || minutesUntilNextPitch > 30 || minutesUntilNextPitch <=0 {
				out.Stop();
			}
			<-p.ticker.C
		}
	}()
}

// StopCheckNext stop the checker for the next pitch
func (p *Pitch) StopCheckNext() {
	p.ticker.Stop()
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
