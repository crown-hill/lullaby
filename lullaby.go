package lullaby

import (
	"fmt"
	"log"
	"os"
	"time"
)

type State struct {
	BedtimeOverride   bool
	SleeptimeOverride bool
}

func (state *State) clear() {

	// Do not start playing music even though it's bedtime.
	state.BedtimeOverride = false

}

func adjustMusicPlayer(config *Config, state *State) (err error) {

	timeConfig, err := config.TimeConfig()
	if err != nil {
		return
	}

	playerStatus, err := getMusicStatus()
	if err != nil {
		return
	}

	now := time.Now()

	if timeConfig.PastSleepTime(now) {

		log.Println("Starting Event Handler")

		state.BedtimeOverride = false

		// It's time to sleep.
		// If the music's playing, turn down the volume.

		// if the volume is already zero, stop
		if playerStatus.Volume == 0 {

			err = stopMusic(os.Stdout)
			if err != nil {
				return
			}

		} else {

			// If music is playing, turn it down

			if playerStatus.IsPlaying {

				err = adjustVolume(0-int32(config.VolumeDownAmount), os.Stdout)
				if err != nil {
					return
				}

			} else {
				// if the music is already off, don't change anything
			}
		}

	} else if timeConfig.PastBedtime(now) {

		log.Println("It's bed time")

		// Time to get to bed and enjoy some tunes.

		if playerStatus.IsPaused &&
			state.BedtimeOverride == false {

			err = setVolume(int32(config.StartVolume), os.Stdout)
			if err != nil {
				return
			}

			err = togglePlayback(os.Stdout)
			if err != nil {
				return
			}

		} else if playerStatus.IsPlaying == false &&
			state.BedtimeOverride == false {

			err = setVolume(int32(config.StartVolume), os.Stdout)
			if err != nil {
				return
			}

			err = playMusic(os.Stdout)
			if err != nil {
				return
			}

		}

	} else if timeConfig.PastWakeTime(now) {
		// TODO
		state.BedtimeOverride = false
	} else {
		state.BedtimeOverride = false
	}

	return
}

type event struct {
	handler eventHandler
}

type eventHandler func()

type Machine struct {
	state            *State
	config           *Config
	events           chan *event
	basicDialHandler *musicControlDialHandler
}

func (m *Machine) HandleClickUp() {

	h := func() {

		timeConfig, err := m.config.TimeConfig()
		if err != nil {
			return
		}

		playerStatus, err := getMusicStatus()
		if err != nil {
			return
		}

		now := time.Now()

		if timeConfig.PastSleepTime(now) {

			if playerStatus.IsPlaying {

				err = stopMusic(os.Stdout)
				if err != nil {
					log.Printf("Error: %s\n", err.Error())
				}

			} else {

				err = playMusic(os.Stdout)
				if err != nil {
					log.Printf("Error: %s\n", err.Error())
				}
			}

		} else if timeConfig.PastBedtime(now) {

			if playerStatus.IsPlaying {

				m.state.BedtimeOverride = true

				err = stopMusic(os.Stdout)
				if err != nil {
					log.Printf("Error: %s\n", err.Error())
				}

			} else {

				m.state.BedtimeOverride = false

				err = playMusic(os.Stdout)
				if err != nil {
					log.Printf("Error: %s\n", err.Error())
				}
			}
		}

		m.basicDialHandler.HandleClickUp()
	}

	e := &event{h}

	m.events <- e

}

func (m *Machine) HandleTurn(clockwise bool, value int32) {

	h := func() {
		m.basicDialHandler.HandleTurn(clockwise, value)
	}

	e := &event{h}

	m.events <- e

}

func (m *Machine) HandleTimer() {

	log.Println("Handling Timer Tick")

	h := func() {
		adjustMusicPlayer(m.config, m.state)
	}

	e := &event{h}

	m.events <- e

}

func (m *Machine) Start(config *Config) {

	m.state = new(State)
	m.config = config

	m.events = make(chan *event)

	go func() {
		log.Println("Starting Event Handler")
		for e := range m.events {
			log.Println("Handling Event")
			e.handler()
		}

	}()

	go func() {
		log.Println("Starting Dial Handler")
		runDial(m)
	}()

	ti, err := m.config.tickerInterval()
	if err != nil {
		panic(err)
	}

	log.Println("Starting Ticker")
	ticker := time.NewTicker(ti)
	done := make(chan bool)

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println("Tick at", t)
			m.HandleTimer()
		}
	}
}
