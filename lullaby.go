package lullaby

import (
	"fmt"
	"log"
	"os"
	"time"
)

type State struct {
	BedtimeStarted    bool
	BedtimeOverride   bool
	SleeptimeOverride bool
	WaketimeStarted   bool
	WorktimeOverride  bool
}

func (state *State) clear() {

	// Do not start playing music even though it's bedtime.
	state.BedtimeOverride = false

}

func resetState(state *State, timeConfig *timeConfig, now time.Time) {

	// Reset any state irellevant to the current time period

	if timeConfig.PastSleeptime(now) {

		//state.SleeptimeOverride = false
		state.BedtimeStarted = false
		state.BedtimeOverride = false
		state.WaketimeStarted = false
		state.WorktimeOverride = false

	} else if timeConfig.PastBedtime(now) {

		state.SleeptimeOverride = false
		//state.BedtimeStarted = false
		//state.BedtimeOverride   = false
		state.WaketimeStarted = false
		state.WorktimeOverride = false

	} else if timeConfig.PastWorktime(now) {

		state.SleeptimeOverride = false
		state.BedtimeStarted = false
		state.BedtimeOverride = false
		state.WaketimeStarted = false
		//state.WorktimeOverride = false

	} else if timeConfig.PastWaketime(now) {

		state.SleeptimeOverride = false
		state.BedtimeOverride = false
		state.BedtimeStarted = false
		//state.WaketimeStarted = false
		state.WorktimeOverride = false

	}
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

	resetState(state, timeConfig, now)

	// Check times from latest to earliest

	if timeConfig.PastSleeptime(now) {

		log.Printf("It's time to sleep. If the music's playing, turn down the volume.\n")

		// if the volume is already zero, stop
		if playerStatus.Volume == 0 {

			log.Println("Volume is already zero, so we'll stop the music.")

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

			}
		}

	} else if timeConfig.PastBedtime(now) {

		log.Printf("It's bed time. Time to get to bed and enjoy some tunes. (override %v)\n", state.BedtimeOverride)

		// Time to get to bed and enjoy some tunes.

		if state.BedtimeStarted == false {

			err = stopAndResetPlaylistWithKexp(os.Stdout)

			if err == nil {

				err = setVolume(int32(config.BaseVolume), os.Stdout)

				if err == nil {

					err = playMusic(os.Stdout)

				}
			}

			if err != nil {
				return
			}

			state.BedtimeStarted = true
		}

		if playerStatus.IsPaused &&
			state.BedtimeOverride == false {

			err = togglePlayback(os.Stdout)
			if err != nil {
				return
			}

		}

	} else if timeConfig.PastWorktime(now) {

		if playerStatus.IsPlaying && state.WorktimeOverride == false {

			err = stopMusic(os.Stdout)

			if err != nil {
				return
			}
		}

	} else if timeConfig.PastWaketime(now) {

		if state.WaketimeStarted == false {

			err = stopAndResetPlaylistWithKexp(os.Stdout)

			if err == nil {

				err = setVolume(int32(config.WakeVolume), os.Stdout)

				if err == nil {

					err = playMusic(os.Stdout)

					state.WaketimeStarted = true

				}
			}

			if err != nil {
				return
			}

			state.WaketimeStarted = true

		} else {

			if uint64(playerStatus.Volume) <= uint64(config.BaseVolume) {

				err = adjustVolume(int32(config.VolumeUpAmount), os.Stdout)
			}

		}

	} else {
		state.BedtimeOverride = false
		state.BedtimeStarted = false
		state.WaketimeStarted = false
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

		log.Println("Lullaby HandleClickUp")

		timeConfig, err := m.config.TimeConfig()
		if err != nil {
			log.Printf("Error: failed to get timeConfig: %s\n", err.Error())
			return
		}

		playerStatus, err := getMusicStatus()
		if err != nil {
			log.Printf("Error: failed to get music status: %s\n", err.Error())
			return
		}

		now := time.Now()

		// Check times from latest to earliest

		if timeConfig.PastSleeptime(now) {

			log.Println("Clicking past sleep time...")

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

			log.Println("Clicking past bed time...")

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

		} else if timeConfig.PastWaketime(now) {

			m.state.WorktimeOverride = !m.state.WorktimeOverride

			m.basicDialHandler.HandleClickUp()

		} else {

			log.Println("Lullaby calling basic dial handler...")

			m.basicDialHandler.HandleClickUp()

		}
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
		err := adjustMusicPlayer(m.config, m.state)

		if err != nil {
			log.Printf("Error adjusting music player: %s\n", err.Error())
		}
	}

	e := &event{h}

	m.events <- e

}

func (m *Machine) Start(config *Config) {

	m.basicDialHandler = new(musicControlDialHandler)

	m.state = new(State)
	m.config = config

	m.events = make(chan *event)

	go func() {
		log.Println("Starting Event Handler")
		for e := range m.events {

			log.Println("Handling Event")
			e.handler()

			log.Println("-- POST HANDLER ----------------------------------------------------------------")
			printMusicStatus(os.Stdout)
			log.Println("--------------------------------------------------------------------------------")
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
