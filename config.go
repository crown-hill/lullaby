package lullaby

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
)

type Config struct {
	DialFile         string
	RadioStream      string
	BedTime          string
	SleepTime        string
	WakeTime         string
	StartVolume      uint
	VolumeDownAmount uint
	TickerInterval   string
}

func DefaltConfig() *Config {

	c := new(Config)

	c.DialFile = "/dev/input/event0"
	c.RadioStream = "http://live-aacplus-64.kexp.org/kexp64.aac"
	c.BedTime = "9:00PM"
	c.SleepTime = "10:00PM"
	c.WakeTime = "6:00AM"
	c.VolumeDownAmount = 2
	c.TickerInterval = "30s"
	c.StartVolume = 100

	return c
}

type timeConfig struct {
	bedtime   time.Time
	sleepTime time.Time
	wakeTime  time.Time
}

func setTimeForDate(targetDate time.Time, targetTime time.Time) time.Time {

	t := time.Date(targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		targetTime.Hour(),
		targetTime.Minute(),
		targetTime.Second(),
		targetTime.Nanosecond(),
		targetDate.Location())

	return t
}

func (c *Config) TimeConfig() (tc *timeConfig, err error) {

	bedTime, err := time.Parse(time.Kitchen, c.BedTime)
	if err != nil {
		return
	}

	sleepTime, err := time.Parse(time.Kitchen, c.SleepTime)
	if err != nil {
		return
	}

	wakeTime, err := time.Parse(time.Kitchen, c.WakeTime)
	if err != nil {
		return
	}

	if sleepTime.Before(bedTime) {
		err = fmt.Errorf("Sleep time %s cannot be before bed time %s", c.SleepTime, c.BedTime)
		return
	}

	if wakeTime.After(bedTime) {
		err = fmt.Errorf("Wake time %s cannot be after bed time %s", c.WakeTime, c.BedTime)
		return
	}

	today := time.Now()

	bedTimeToday := setTimeForDate(today, bedTime)
	sleepTimeToday := setTimeForDate(today, sleepTime)
	wakeTimeToday := setTimeForDate(today, wakeTime)

	log.Printf("bed: %s\n", bedTimeToday)
	log.Printf("sleep: %s\n", sleepTimeToday)

	// if it's already past wake time, move it to tomorrow
	//if today.After(wakeTimeToday) {
	//	wakeTimeToday = wakeTimeToday.Add(time.Hour * 24)
	//	}

	log.Printf("wake: %s\n", wakeTimeToday)

	tc = new(timeConfig)

	tc.bedtime = bedTimeToday
	tc.wakeTime = wakeTimeToday
	tc.sleepTime = sleepTimeToday

	log.Printf("curent %s\n", today)

	log.Printf("past bedtime: %v\n", tc.PastBedtime(today))
	log.Printf("before bedtime: %v\n", tc.BeforeBedtime(today))

	log.Printf("past sleeptime: %v\n", tc.PastSleepTime(today))
	log.Printf("past waketime: %v\n", tc.PastWakeTime(today))

	return
}

func (tc *timeConfig) PastBedtime(refTime time.Time) bool {
	return refTime.After(tc.bedtime)
}

func (tc *timeConfig) BeforeBedtime(refTime time.Time) bool {
	return refTime.Before(tc.bedtime)
}

func (tc *timeConfig) PastSleepTime(refTime time.Time) bool {
	return refTime.After(tc.sleepTime)
}

func (tc *timeConfig) PastWakeTime(refTime time.Time) bool {
	return refTime.After(tc.wakeTime)
}

func ReadConfig(rdr io.Reader) (c *Config, err error) {

	dec := json.NewDecoder(rdr)

	c = DefaltConfig()

	err = dec.Decode(c)

	if err != nil {
		c = nil
	}

	return

}

func (config *Config) tickerInterval() (i time.Duration, err error) {

	return time.ParseDuration(config.TickerInterval)

}
