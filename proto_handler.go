package main

import (
    "encoding/json"
    "errors"
    "net/http"
    "github.com/op/go-logging"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/model"
)

const TOPIC_LOCATION = "location"
const MQTT_LAT = "lat"
const MQTT_LNG = "lng"
const MQTT_DATE = "ts"
const MQTT_SAT = "sat"

type ProtoHandler struct {
    log *logging.Logger
    piotDevices *piot.PiotDevices
    things *piot.Things
    mqtt piot.IMqtt
}

func NewProtoHandler(
        log *logging.Logger,
        piotDevices *piot.PiotDevices,
        things *piot.Things,
        mqtt piot.IMqtt) *ProtoHandler {

    return &ProtoHandler{log: log, piotDevices: piotDevices, things: things, mqtt: mqtt}
}

func (h *ProtoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    h.log.Debugf("Incoming packet uri: %s, method: %s, content length: %d", r.RequestURI, r.Method, r.ContentLength)

    packet, err := ProtoParse(h.log, r.Body)
    if err != nil {
        h.log.Errorf("Parsing request body failed: %s", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    h.log.Debugf("Packet successfully parsed (%d locations)", len(packet.Locs))

    err = h.processPacket(packet)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
    }
}

func (h *ProtoHandler) processPacket(packet *ProtoPacket) error {

    // name of the device cannot be empty
    if packet.Id == "" {
        return errors.New("Device name (id) cannot be empty")
    }

    // get instance of Things service and look for the device (chip),
    // register it if it doesn't exist
    thing, err := h.things.FindPiot(packet.Id)
    if err != nil {
        // register device
        thing, err = h.things.RegisterPiot(packet.Id, model.THING_TYPE_DEVICE)
        if err != nil {
            return err
        }

        // set default location values for registered device
        if err := h.things.SetLocationMqttTopic(thing.Id, TOPIC_LOCATION); err != nil {
            return err
        }
        if err := h.things.SetLocationMqttValues(thing.Id, MQTT_LAT, MQTT_LNG, MQTT_DATE); err != nil {
            return err
        }

    }

    // if thing is assigned to org
    if thing.OrgId == primitive.NilObjectID {
        h.log.Debugf("Ignoring processing of data for thing <%s> that is not assigned to any organization", thing.Name)
        return nil
    }

    if !thing.Enabled {
        h.log.Debugf("Ignoring processing of data for thing <%s> that is not enabled", thing.Name)
        return nil
    }

    // loop through all locations
    for _, loc := range packet.Locs {

        // loc to json
        locRawData, err := json.Marshal(loc)
        if err != nil {
            return err
        }

        err = h.mqtt.PushThingData(thing, TOPIC_LOCATION, string(locRawData))
        if err != nil {
            return err
        }
    }

    return nil
}

