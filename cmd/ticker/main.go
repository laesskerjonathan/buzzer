package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"gitlab.com/marcsauter/pitchbuzzer/pitch"
	"gitlab.com/marcsauter/ticker"
)

func main() {

	device := os.Getenv("PTICKER_DEVICE")
	if _, err := os.Stat(device); os.IsNotExist(err) {
		log.Fatal("PTICKER_DEVICE missing or not valid")
	}
	url, err := url.Parse(os.Getenv("PTICKER_PITCH_URL"))
	if err != nil || !url.IsAbs() {
		log.Fatal("PTICKER_PITCH_URL missing or not valid")
	}
	interval, err := strconv.Atoi(os.Getenv("PTICKER_PITCH_CHECK_INTERVAL"))
	if err != nil {
		log.Fatal("PTICKER_PITCH_CHECK_INTERVAL missing or not valid")
	}

	t, err := ticker.NewTicker(device)
	if err != nil {
		log.Fatal(err)
	}

	p := pitch.NewPitch(url)
	p.StartCheckNext(interval, t)

	//
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for {
		select {
		case <-cancel:
			log.Fatalln("signal received - exiting")
		}
	}
}
