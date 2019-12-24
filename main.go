package main

import (
	"flag"
	"log"

	"gihutb.com/jxmoore/hubbub/bootstrap"
)

var configPath = flag.String("c", "./config.json", "The path for the config file.")
var envOnly = flag.Bool("e", false, "Use only enviroment variables.")

func main() {

	flag.Parse()

	if err := bootstrap.BootStrap(*configPath, *envOnly); err != nil {
		log.Fatal(err.Error())
	}

}
