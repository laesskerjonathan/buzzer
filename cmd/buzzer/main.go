package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/marcsauter/buzzer/pkg/pitch"

	"github.com/luismesas/goPi/piface"
	"github.com/luismesas/goPi/spi"
)

func main() {
	device := os.Getenv("PBUZZER_KEYPAD_DEVICE")
	if len(device) == 0 {
		log.Fatal("PBUZZER_KEYPAD_DEVICE missing or not valid")
	}
	url, err := url.Parse(os.Getenv("PBUZZER_PITCH_URL"))
	if err != nil || !url.IsAbs() {
		log.Fatal("PBUZZER_PITCH_URL missing or not valid")
	}
	interval, err := strconv.Atoi(os.Getenv("PBUZZER_PITCH_CHECK_INTERVAL"))
	if err != nil {
		log.Fatal("PBUZZER_PITCH_CHECK_INTERVAL missing or not valid")
	}

	// creates a new pifacedigital instance
	pfd := piface.NewPiFaceDigital(spi.DEFAULT_HARDWARE_ADDR, spi.DEFAULT_BUS, spi.DEFAULT_CHIP)

	// initializes pifacedigital board
	err = pfd.InitBoard()
	if err != nil {
		fmt.Printf("Error on init board: %s", err)
		return
	}
	b := NewBuzzer(pfd)
	h := NewHorn(pfd)
	h.WatchButton()
	l := NewLight(pfd)
	l.WatchButton()
	k, err := NewKeypad(device)
	if err != nil {
		log.Fatal(err)
	}
	//
	s := NewScreen()
	s.Init("buzzer", "Pitch Info", "Pitch Info")
	s.Main()
	s.StartTicker()
	//
	p := pitch.NewPitch(url)
	p.StartCheckNext(interval, s)
	//
	startPitch := func(pid string) error {
		url.Path = fmt.Sprintf("start/%s", pid)
		resp, err := http.Get(url.String())
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		var p struct {
			Message string
		}
		if err := decoder.Decode(&p); err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New(p.Message)
		}
		return nil
	}
	//
	code := k.Start()
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for {
		select {
		case c := <-code:
			if err := startPitch(c); err != nil {
				s.Keypad(fmt.Sprintf(DefaultKeypadText, fmt.Sprintf("ERROR: %s", err.Error())))
			} else {
				s.Keypad(fmt.Sprintf("Pitch Code valid - Please press the Buzzer to release the Pitch ...\n"))
				b.Watch()
				s.Keypad(fmt.Sprintf(DefaultKeypadText, fmt.Sprintf("%s - Pitch %s released", time.Now().Format("02.01.2006 15:04:05"), c)))
				l.On() // light on
				h.On() // horn on
			}
		case <-cancel:
			p.StopCheckNext()
			b.Unwatch()
			l.UnwatchButton()
			l.Off()
			h.UnwatchButton()
			h.Off()
			k.Stop()
			s.StopCountdown()
			s.Destroy()
			log.Fatalln("signal received - exiting")
		}
	}
}
