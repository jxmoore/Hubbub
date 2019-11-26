package main

import (
	"flag"
	"log"

	"gihutb.com/jxmoore/hubbub/controller"
)

var configPath = flag.String("c", "./config.json", "The path for the config file.")
var envOnly = flag.Bool("e", false, "Use only enviroment variables.")

func main() {

	flag.Parse()

	err := controller.BootStrap(*configPath, *envOnly)
	if err != nil {
		log.Fatal(err.Error())
	}

}
