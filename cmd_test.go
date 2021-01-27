package lullaby

import (
	"encoding/json"
	"testing"
)

func TestCommands(t *testing.T) {

	err := runCommandAndPrintOutput("pwd", []string{}, nil)

	if err != nil {
		t.Error(err)
	}

	err = togglePlayback(nil)

	if err != nil {
		t.Error(err)
	}

}

func TestStatus(t *testing.T) {

	l := `KEXP.ORG 90.3FM - where the music matters
[paused] #1/13   0:00/0:00 (0%)
volume: 52%   repeat: off   random: off   single: off   consume: off`

	s, err := parseMusicStatus([]byte(l))
	if err != nil {
		t.Error(err)
	}

	_ = s

	j, err := json.MarshalIndent(s, "", "   ")
	if err != nil {
		t.Error(err)
	}

	t.Log(string(j))

}
