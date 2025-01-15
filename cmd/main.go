package main

import (
	"fmt"
	"os"

	"github.com/kahsengphoon/Tron-Votes/cmd/handler"
	"github.com/urfave/cli/v2"
)

func main() {
	svr := handler.NewHttpsServer()
	app := cli.NewApp()
	app.Name = "mpc-backend"
	app.Usage = "mpc CLI"
	app.Version = "1.0.0"

	// find TARGET
	target := os.Getenv("TARGET")
	if target != "" {
		arg := fmt.Sprintf("start-%s-app", target)
		os.Args = append(os.Args, arg)
		fmt.Printf("Target: %s \n", arg)
	}
	app.Commands = []*cli.Command{
		{
			Name:   "start-api-app",
			Usage:  "Start API Server",
			Action: svr.StartApiServer,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
