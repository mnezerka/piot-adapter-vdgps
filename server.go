package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "github.com/urfave/cli"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/config"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

const LOG_FORMAT = "%{color}%{time:2006/01/02 15:04:05 -07:00 MST} [%{level:.6s}] %{shortfile} : %{color:reset}%{message}"

func runServer(c *cli.Context) {

    cfg := config.NewParameters()
    cfg.DbUri = c.GlobalString("mongodb-uri")
    cfg.DbName = "piot"
    cfg.LogLevel = c.GlobalString("log-level")

    ///////////////// LOGGER instance
    logger, err := piot.NewLogger(LOG_FORMAT, cfg.LogLevel)
    if err != nil {
        log.Fatalf("Cannot create logger for level %s (%v)", cfg.LogLevel, err)
        os.Exit(1)
    }

    /////////////// DB (mongo)
    dbUri := c.GlobalString("mongodb-uri")

    // try to open database
    dbClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dbUri))
    if err != nil {
        logger.Fatalf("Failed to open database on %s (%v)", dbUri, err)
        os.Exit(1)
    }

    // Check the connection
    err = dbClient.Ping(context.TODO(), nil)
    if err != nil {
        logger.Fatalf("Cannot ping database on %s (%v)", dbUri, err)
        os.Exit(1)
    }

    db := dbClient.Database("piot")

    /////////////// ORGS service
    orgs := piot.NewOrgs(logger, db)

    /*

    /////////////// HTTP CLIENT service
    var httpClient piot.IHttpClient
    httpClient = piot.NewHttpClient(logger)

    /////////////// PIOT INFLUXDB SERVICE
    influxDbUri := c.GlobalString("influxdb-uri")
    influxDbUsername := c.GlobalString("influxdb-user")
    influxDbPassword := c.GlobalString("influxdb-password")
    influxDb := piot.NewInfluxDb(logger, orgs, httpClient, influxDbUri, influxDbUsername, influxDbPassword)

    /////////////// PIOT MYSQLDB SERVICE
    mysqlDbHost := c.GlobalString("mysqldb-host")
    mysqlDbUsername := c.GlobalString("mysqldb-user")
    mysqlDbPassword := c.GlobalString("mysqldb-password")
    mysqlDbName := c.GlobalString("mysqldb-name")
    mysqlDb := piot.NewMysqlDb(logger, orgs, mysqlDbHost, mysqlDbUsername, mysqlDbPassword, mysqlDbName)
    err = mysqlDb.Open()
    if err != nil {
        logger.Fatalf("Connect to mysql server failed %v", err)
        os.Exit(1)
    }
    */

    //////////////// THINGS service instance
    things := piot.NewThings(db, logger)

    /////////////// PIOT MQTT service instance
    mqttUri := c.GlobalString("mqtt-uri")
    mqttUsername := c.GlobalString("mqtt-user")
    mqttPassword := c.GlobalString("mqtt-password")
    mqttClient := c.GlobalString("mqtt-client")
    mqtt := piot.NewMqtt(mqttUri, logger, things, orgs, nil, nil)
    mqtt.SetUsername(mqttUsername)
    mqtt.SetPassword(mqttPassword)
    mqtt.SetClient(mqttClient)
    err = mqtt.Connect()
    if err != nil {
        logger.Fatalf("Connect to mqtt server failed %v", err)
        os.Exit(1)
    }

    /////////////// PIOT DEVICES service instance
    piotDevices := piot.NewPiotDevices(logger, things, mqtt, cfg)

    protoHandler := NewProtoHandler(logger, piotDevices, things, mqtt)
    http.Handle("/", protoHandler)

    logger.Infof("Listening on %s...", c.GlobalString("bind-address"))
    err = http.ListenAndServe(c.GlobalString("bind-address"), nil)
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
        cli.StringFlag{
            Name:   "mqtt-uri,q",
            Usage:  "Endpoint for the Mosquitto message broker",
            EnvVar: "MQTT_URI",
            Value:  "tcp://localhost:1883",
        },
        cli.StringFlag{
            Name:   "log-level,l",
            Usage:  "Logging level",
            Value:  "INFO",
            EnvVar: "LOG_LEVEL",
        },
        cli.StringFlag{
            Name:   "mqtt-user",
            Usage:  "Username for mqtt authentication",
            EnvVar: "MQTT_USER",
        },
        cli.StringFlag{
            Name:   "mqtt-password",
            Usage:  "Password for mqtt authentication",
            EnvVar: "MQTT_PASSWORD",
        },
        cli.StringFlag{
            Name:   "mqtt-client",
            Usage:  "Id used for identification of this mqtt client",
            Value:  "piot-adapter-vdgps",
            EnvVar: "MQTT_CLIENT",
        },
    }

    app.Run(os.Args)
}
