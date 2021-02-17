package lullaby

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type dialHandler interface {
	HandleClickUp()
	HandleTurn(clockwise bool, value int32)
}

type musicControlDialHandler struct {
	outputWriter io.Writer
}

func (h *musicControlDialHandler) HandleClickUp() {

	//togglePlayback(h.outputWriter)
	hardTogglePlayback(h.outputWriter)

}

func (h *musicControlDialHandler) HandleTurn(clockwise bool, value int32) {
	_ = clockwise
	
	// Hack?
	// Always adjust by 1
	
	if value < 0 {
		value = -1
	} else {
		value = 1
	}
	
	adjustVolume(value, h.outputWriter)
}

func runDial(handler dialHandler) {

	printMusicStatus(nil)

	f, err := os.Open("/dev/input/event0")

	if err != nil {
		log.Println("Cannot open dial file: %s\n", err.Error())
		return
	}

	defer f.Close()

	b := make([]byte, 16)

	for {
		f.Read(b)

		sec := binary.LittleEndian.Uint32(b[0:8])
		t := time.Unix(int64(sec), 0)
		_ = t

		var value int32
		typ := binary.LittleEndian.Uint16(b[8:10])
		code := binary.LittleEndian.Uint16(b[10:12])
		binary.Read(bytes.NewReader(b[12:]), binary.LittleEndian, &value)

		//fmt.Printf("type: %x\ncode: %d\nvalue: %d\n", typ, code, value)

		if typ == 0x2 && code == 7 {
			if value > 0 {

				fmt.Printf("Clockwise Turn (value %d)\n", value)
				//adjustVolume(value, nil)

				handler.HandleTurn(true, value)

			} else if value < 0 {

				fmt.Printf("Counter-clockwise Turn (value %d)\n", value)
				//adjustVolume(value, nil)
				handler.HandleTurn(false, value)

			}

		} else if typ == 0x1 && code == 256 {
			if value == 1 {

				fmt.Println("Click Down")

			} else if value == 0 {

				fmt.Println("Click Up")

				handler.HandleClickUp()

			}

		}

	}
}
