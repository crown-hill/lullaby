package main

import (
	"flag"
	"log"
	"os"

	"github.com/crown-hill/lullaby"
)

var (
	configFile string
)

func init() {

	flag.StringVar(&configFile, "config", "lul.config", "Path to Lullaby config file.")
}

func main() {

	flag.Parse()

	config := lullaby.DefaltConfig()

	configRdr, rerr := os.Open(configFile)

	if rerr != nil {
		log.Printf("Could not open %s, using default config. Error: %s\n", configFile, rerr.Error())
	} else {

		config, rerr = lullaby.ReadConfig(configRdr)

		configRdr.Close()

	}

	lulabyMachine := new(lullaby.Machine)

	lulabyMachine.Start(config)

}
