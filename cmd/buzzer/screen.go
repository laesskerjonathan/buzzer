package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"time"

	"gitlab.com/marcsauter/pitchbuzzer/pitch"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

const (
	DefaultMonospaceFontExtraSmall = "Monospace 20"
	DefaultMonospaceFontSmall      = "Monospace 40"
	DefaultMonospaceFontLarge      = "Monospace 50"
	DefaultFontExtraSmall          = "Sans 20"
	DefaultFontSmall               = "Sans 40"
	DefaultFontLarge               = "Sans 50"
	DefaultKeypadText              = "%s\nEnter a valid Pitch Code to release the Buzzer ... "
)

//
type Screen struct {
	Ticker        string
	Status        string
	PitchCode     string
	ticker        *gtk.Label
	title         *gtk.Label
	speaker       *gtk.Label
	countdown     *gtk.Label
	keypad        *gtk.Label
	stopTicker    bool
	stopCountdown bool
}

//
func NewScreen() *Screen {
	return &Screen{
		Ticker: "NEXT",
		Status: "IP: %s",
	}
}

//
func (s *Screen) Init(name, winTitle, iconName string) {
	gdk.ThreadsInit()
	gtk.Init(nil)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle(winTitle)
	window.SetIconName(iconName)
	window.Fullscreen()
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		log.Println("got destroy", ctx.Data().(string))
		gtk.MainQuit()
	}, name)

	box := gtk.NewVBox(false, 1)

	// ticker
	tickerFrame := gtk.NewFrame("")
	s.ticker = gtk.NewLabel("")
	s.ticker.ModifyFontEasy(DefaultMonospaceFontLarge)
	tickerFrame.Add(s.ticker)
	box.Add(tickerFrame)

	// pitch
	pitchFrame := gtk.NewFrame("")
	pitchBox := gtk.NewVBox(false, 1)
	pitchFrame.Add(pitchBox)

	s.speaker = gtk.NewLabel("")
	s.speaker.ModifyFontEasy(DefaultFontSmall)
	pitchBox.Add(s.speaker)

	s.title = gtk.NewLabel("")
	s.title.ModifyFontEasy(DefaultFontLarge)
	pitchBox.Add(s.title)
	box.Add(pitchFrame)

	// countdown
	countdownFrame := gtk.NewFrame("")
	s.countdown = gtk.NewLabel("")
	s.countdown.ModifyFontEasy(DefaultFontLarge)
	s.countdown.ModifyFG(0, gdk.NewColor("red"))
	countdownFrame.Add(s.countdown)
	box.Add(countdownFrame)

	// keypad
	keypadFrame := gtk.NewFrame("")
	s.keypad = gtk.NewLabel(fmt.Sprintf(DefaultKeypadText, "   "))
	s.keypad.ModifyFontEasy(DefaultMonospaceFontExtraSmall)
	keypadFrame.Add(s.keypad)
	box.Add(keypadFrame)

	// statusbar
	statusbar := gtk.NewStatusbar()
	context_id := statusbar.GetContextId("go-gtk")
	statusbar.Push(context_id, s.statusText())
	box.PackStart(statusbar, false, false, 0)

	// window
	window.Add(box)
	window.SetSizeRequest(640, 480)
	window.ShowAll()
}

//
func (s *Screen) Main() {
	go func() {
		gdk.ThreadsEnter()
		gtk.Main()
		gdk.ThreadsLeave()
	}()
}

//
func (s *Screen) Destroy() {
	gtk.MainQuit()
}

//
func (s *Screen) setLabel(label *gtk.Label, text string) {
	gdk.ThreadsEnter()
	label.SetLabel(text)
	gdk.ThreadsLeave()
}

//
func (s *Screen) StartTicker() {
	go func() {
		text := s.Ticker
		// enough text for a ticker illusion
		for i := 0; i < 25; i++ {
			text = fmt.Sprintf("%s - %s", text, s.Ticker)
		}
		count := len(s.Ticker) + 3 // three character for text separation
		i := 0
		ticker := time.NewTicker(time.Millisecond * 1000)
		for !s.stopTicker {
			if i == count {
				i = 0
			}
			s.setLabel(s.ticker, text[i:])
			i = i + 1
			<-ticker.C
		}
		s.stopTicker = false
	}()
}

//
func (s *Screen) StopTicker() {
	s.stopTicker = true
}

//
func (s *Screen) Update(data fmt.Stringer) error {
	p, _ := data.(*pitch.Pitch)
	s.setLabel(s.speaker, p.Speaker)
	s.setLabel(s.title, p.Title)
	s.stopCountdown = false
	go func() {
		ticker := time.NewTicker(time.Millisecond * 250)
		for !s.stopCountdown {
			r := time.Now().Sub(p.Date)
			sign := ""
			if r < 0 {
				sign = "-"
			}
			hrs := int(math.Abs(r.Hours()))
			min := int(math.Abs(r.Minutes())) - hrs*60
			sec := int(math.Abs(r.Seconds())) - hrs*3600 - min*60
			s.setLabel(s.countdown, fmt.Sprintf("%s %02dh %02dm %02ds", sign, hrs, min, sec))
			<-ticker.C
		}
		s.stopCountdown = false
	}()
	return nil
}

//
func (s *Screen) StopCountdown() {
	s.stopCountdown = true
}

//
func (s *Screen) Keypad(text string) {
	s.setLabel(s.keypad, text)
}

//
func (s *Screen) statusText() string {
	var ipAddrs []string
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			ip, _, _ := net.ParseCIDR(addr.String())
			if ip.To4() != nil && !ip.IsLoopback() {
				ipAddrs = append(ipAddrs, addr.String())
			}
		}
	}
	return fmt.Sprintf(s.Status, strings.Join(ipAddrs, ", "))
}
