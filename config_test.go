package lullaby

import (
	"testing"
)

func TestTiming(t *testing.T) {

	/*timeStr := "4:00PM"

	pt, err := time.Parse(time.Kitchen, timeStr)

	if err != nil {
		t.Error(err)
	}

	t.Logf("parsed: %s\n", pt)

	c := time.Now()

	alramTime := time.Date(c.Year(),
		c.Month(),
		c.Day(),
		pt.Hour(),
		pt.Minute(),
		pt.Second(),
		pt.Nanosecond(),
		c.Location())

	t.Logf("current time: %s\n", c)
	t.Logf("alarm time: %s\n", alramTime)*/

	c := DefaltConfig()

	_, err := c.TimeConfig()

	if err != nil {
		t.Error(err)
	}

}
