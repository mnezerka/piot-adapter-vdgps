package main

import (
    "log"
    "net/http"
    "os"
    "github.com/urfave/cli"
    "piot-adapter-vdgps/handler"
)

func runServer(c *cli.Context) {

    http.HandleFunc("/", handler.Adapter)

    log.Printf("Listening on %s...", c.GlobalString("bind-address"))
    err := http.ListenAndServe(c.GlobalString("bind-address"), nil)
    FatalOnError(err, "Failed to bind on %s: ", c.GlobalString("bind-address"))
}

func FatalOnError(err error, msg string, args ...interface{}) {
    if err != nil {
        log.Fatalf(msg, args...)
        os.Exit(1)
    }
}

func main() {
    app := cli.NewApp()

    app.Name = "PIOT Adapter VDGps"
    app.Version = "1.0"
    app.Authors = []cli.Author{
        {
            Name:  "Michal Nezerka",
            Email: "michal.nezerka@gmail.com",
        },
    }
    app.Usage = "PIOT Adapter for VD Gps sensors"
    app.Action = runServer
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:   "bind-address,b",
            Usage:  "Listen address for API HTTP endpoint",
            Value:  "0.0.0.0:8888",
            EnvVar: "BIND_ADDRESS",
        },
    }

    app.Run(os.Args)
}
