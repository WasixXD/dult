package main

import (
	"dult/internals"
	"flag"
	"fmt"
)

func main() {

	configPath := flag.String("config", "foo", "-config=\"path_to_your_config.json\"")

	flag.Parse()

	if *configPath == "foo" {
		fmt.Println("See: dult -h")
		return
	}

	c := internals.Configuration{}

	err := c.ReadConfig(*configPath)

	if err != nil {
		fmt.Printf("ERROR: path %s does not exist", *configPath)
		return
	}

	d := internals.Dult{}
	d.FromConfig(c)

	d.Start()
}
