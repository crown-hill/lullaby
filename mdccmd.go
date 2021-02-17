package lullaby

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func runCommandAndPrintOutput(name string, args []string, output io.Writer) (err error) {

	log.Printf("command: %s %s\n", name, strings.Join(args, " "))

	if output == nil {
		output = os.Stdout
	}

	cmd := exec.Command(name, args...)

	var cmdOutput []byte

	cmdOutput, err = cmd.CombinedOutput()

	if err != nil {
		fmt.Fprintf(output, "Error running commands: %s\n", err.Error())
	} else {
		fmt.Fprintf(output, "%s\n", string(cmdOutput))
	}

	return err

}

func printMusicStatus(output io.Writer) (err error) {

	err = runCommandAndPrintOutput("mpc", []string{"status"}, output)

	return
}

type musicPlayerStatus struct {
	Title     string
	IsPlaying bool
	IsPaused  bool
	Volume    uint64
}

func parseMusicStatus(buf []byte) (s *musicPlayerStatus, err error) {

	//log.Println("- MUSIC STATUS RAW -------------------------------------------------------------")
	//log.Println(string(buf))
	//log.Println("--------------------------------------------------------------------------------")

	lines := strings.Split(string(buf), "\n")

	if len(lines) < 1 {
		err = fmt.Errorf("failed to parse mdc status output: %s", string(buf))
		return
	}

	parseVolume := func(line string) (v int64, perr error) {

		vpat := regexp.MustCompile(`volume:\s*(\d*)`)

		m := vpat.FindStringSubmatch(line)

		if len(m) == 2 {
			v, perr = strconv.ParseInt(m[1], 10, 64)
		}

		return
	}

	parseStatus := func(line string) (isPlaying bool, isPaused bool) {
		isPlaying = strings.Contains(line, "[playing]")
		isPaused = strings.Contains(line, "[paused]")
		return
	}

	s = new(musicPlayerStatus)

	var v int64

	switch len(lines) {
	case 3, 4, 5:
		s.Title = lines[0]

		isPlaying, isPaused := parseStatus(lines[1])

		s.IsPaused = isPaused
		s.IsPlaying = isPlaying

		v, err = parseVolume(lines[2])

		s.Volume = uint64(v)

	case 2:
		err = fmt.Errorf("failed to parse mdc status output, only 2 lines: %s", strings.Join(lines, "\n"))
		return
	case 1:
		v, err = parseVolume(lines[0])
		s.Volume = uint64(v)
	default:
		err = fmt.Errorf("failed to parse mdc status output, too many lines: %d", len(lines))
		return
	}

	return
}

func getMusicStatus() (s *musicPlayerStatus, err error) {

	buf := new(bytes.Buffer)

	err = printMusicStatus(buf)

	if err != nil {
		return
	}

	return parseMusicStatus(buf.Bytes())

}

func setVolume(value int32, output io.Writer) (err error) {

	err = runCommandAndPrintOutput("mpc", []string{"volume", fmt.Sprintf("%d", value)}, output)

	return
}

func adjustVolume(value int32, output io.Writer) (err error) {

	if value > 0 {

		err = runCommandAndPrintOutput("mpc", []string{"volume", fmt.Sprintf("+%d", value)}, output)

	} else if value < 0 {

		err = runCommandAndPrintOutput("mpc", []string{"volume", fmt.Sprintf("%d", value)}, output)
	}

	return
}

func togglePlayback(output io.Writer) (err error) {

	return runCommandAndPrintOutput("mpc", []string{"toggle"}, output)
}

func hardTogglePlayback(output io.Writer) (err error) {
	
	status, err := getMusicStatus()
	
	if err != nil {
		return err
	}
	
	if status.IsPlaying {
		err = stopMusic(output)
	} else {
		err = playMusic(output)
	}

	return err
}

func playMusic(output io.Writer) (err error) {

	return runCommandAndPrintOutput("mpc", []string{"play"}, output)
}

func stopMusic(output io.Writer) (err error) {

	return runCommandAndPrintOutput("mpc", []string{"stop"}, output)
}

func clearPlaylist(output io.Writer) (err error) {
	return runCommandAndPrintOutput("mpc", []string{"clear"}, output)
}

func addToPlaylist(music string, output io.Writer) (err error) {
	return runCommandAndPrintOutput("mpc", []string{"add", music}, output)
}

func addKexp(output io.Writer) (err error) {
	return addToPlaylist("http://live-aacplus-64.kexp.org/kexp64.aac", output)

}

func stopAndResetPlaylistWithKexp(output io.Writer) (err error) {

	err = stopMusic(output)

	if err == nil {

		err = clearPlaylist(output)

		if err == nil {
			err = addKexp(output)
		}
	}

	return
}
