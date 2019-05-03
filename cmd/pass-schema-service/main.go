package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "schemas"
	app.Usage = "PASS schema service"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		serve,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
