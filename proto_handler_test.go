package main

import (
    "context"
    "strings"
    "testing"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "github.com/op/go-logging"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/model"
    "github.com/mnezerka/go-piot/test"
)

type services struct {
    log *logging.Logger
    db *mongo.Database
    things *piot.Things
    mqtt *test.MqttMock
    pdevices *piot.PiotDevices
    phandler *ProtoHandler
}

func getServices(t *testing.T) *services {
    services := services{}
    services.log = test.GetLogger(t)
    services.db = test.GetDb(t)
    services.things = test.GetThings(t, services.log, services.db)
    services.mqtt = test.GetMqtt(t, services.log)
    services.pdevices = test.GetPiotDevices(t , services.log, services.things, services.mqtt)
    services.phandler = NewProtoHandler(services.log, services.pdevices, services.things, services.mqtt)

    test.CleanDb(t, services.db)

    return &services
}

func contains(t *testing.T, str, pattern string) {
    test.Assert(t, strings.Contains(str, pattern), "String <" + str + "> doesn't contain <" + pattern + ">")
}

func TestPacketDeviceReg(t *testing.T) {

    const DEVICE = "device01"

    s := getServices(t)

    packet := ProtoPacket{
        Id: DEVICE,
        Voltage: 23.45,
    }
    packet.Locs = append(packet.Locs, &GpsMeasurement{
        Latitude: 34.56,
        Longitude: 12.34,
    })

    err := s.phandler.processPacket(&packet)
    test.Ok(t, err)

    // Check if device is registered
    var thing model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": DEVICE}).Decode(&thing)
    test.Ok(t, err)
    test.Equals(t, DEVICE, thing.Name)
    test.Equals(t, model.THING_TYPE_DEVICE, thing.Type)
    test.Equals(t, TOPIC_LOCATION, thing.LocationTopic)
    test.Equals(t, MQTT_LAT, thing.LocationLatValue)
    test.Equals(t, MQTT_LNG, thing.LocationLngValue)
    test.Equals(t, MQTT_DATE, thing.LocationDateValue)

    // check that no mqtt msgs were sent out
    test.Equals(t, 0, len(s.mqtt.Calls))
}

func TestPacketDeviceAssigned(t *testing.T) {

    const DEVICE = "device01"

    s := getServices(t)

    // create and assign thing to org
    test.CreateDevice(t, s.db, DEVICE)
    orgId := test.CreateOrg(t, s.db, "org1")
    test.AddOrgThing(t, s.db, orgId, DEVICE)

    packet := ProtoPacket{
        Id: DEVICE,
        Voltage: 23.45,
    }
    packet.Locs = append(packet.Locs, &GpsMeasurement{
        Latitude: 34.56,
        Longitude: 12.34,
        Satelites: 3,
    })

    err := s.phandler.processPacket(&packet)
    test.Ok(t, err)

    // check that no mqtt msgs were sent out
    test.Equals(t, 1, len(s.mqtt.Calls))
    test.Equals(t, "location", s.mqtt.Calls[0].Topic)
    contains(t, s.mqtt.Calls[0].Value, "\"ts\":0")
    contains(t, s.mqtt.Calls[0].Value, "\"lat\":34.56")
    contains(t, s.mqtt.Calls[0].Value, "\"lng\":12.34")
    contains(t, s.mqtt.Calls[0].Value, "\"sat\":3")
}
