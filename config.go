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
	WorkTime         string
	BaseVolume       uint
	WakeVolume       uint
	VolumeDownAmount uint
	VolumeUpAmount   uint
	TickerInterval   string
	Location         string
}

func DefaltConfig() *Config {

	c := new(Config)

	c.DialFile = "/dev/input/event0"
	c.RadioStream = "http://live-aacplus-64.kexp.org/kexp64.aac"
	c.BedTime = "8:00PM"
	c.SleepTime = "11:00PM"
	c.WakeTime = "6:00AM"
	c.WorkTime = "7:00AM"
	c.VolumeDownAmount = 2
	c.VolumeUpAmount = 2
	c.TickerInterval = "30s"
	c.BaseVolume = 50
	c.WakeVolume = 2
	c.Location = "America/Los_Angeles"

	return c
}

type timeConfig struct {
	bedtime   time.Time
	sleepTime time.Time
	wakeTime  time.Time
	workTime  time.Time
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

	workTime, err := time.Parse(time.Kitchen, c.WorkTime)
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

	if workTime.Before(wakeTime) {
		err = fmt.Errorf("Work time %s cannot be before wake time %s", c.WorkTime, c.WakeTime)
	}

	location, err := time.LoadLocation(c.Location)
	if err != nil {
		return
	}

	now := time.Now().In(location)

	log.Printf("curent %s\n", now)

	bedTimenow := setTimeForDate(now, bedTime)
	sleepTimenow := setTimeForDate(now, sleepTime)
	wakeTimenow := setTimeForDate(now, wakeTime)

	log.Printf("bed: %s\n", bedTimenow)
	log.Printf("sleep: %s\n", sleepTimenow)

	log.Printf("wake: %s\n", wakeTimenow)

	tc = new(timeConfig)

	tc.bedtime = bedTimenow
	tc.wakeTime = wakeTimenow
	tc.sleepTime = sleepTimenow

	log.Printf("past bedtime: %v\n", tc.PastBedtime(now))
	log.Printf("before bedtime: %v\n", tc.BeforeBedtime(now))

	log.Printf("past sleeptime: %v\n", tc.PastSleepTime(now))
	log.Printf("past waketime: %v\n", tc.PastWakeTime(now))

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

func (tc *timeConfig) PastWorkTime(refTime time.Time) bool {
	return refTime.After(tc.workTime)
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
