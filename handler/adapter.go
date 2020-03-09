package handler

import (
    "bytes"
    "encoding/binary"
    "io/ioutil"
    "log"
    "net/http"
    "time"
)

type TGpsMeasurement struct {
    Timestamp int64  `json:"timestamp"`
    Latitude float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Satelites int32 `json:"satelites"`
}

func Adapter(w http.ResponseWriter, r *http.Request) {

    log.Printf("Request uri: %s,  method: %s, content length: %d", r.RequestURI, r.Method, r.ContentLength)

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("Reading request body error: %s", err)
    }
    log.Printf("Read request body passed (size: %d)", len(body))

    // Split body to blocks (separator is "|")
    // last slice should contain rest of the body
    blocks := bytes.SplitN(body, []byte("|"), 5)

    log.Printf("Detected %d blocks", len(blocks))

    if len(blocks) < 4 {
        log.Printf("Invalid number of blocks")
        return
    }

    // we don't need key
    //key := blocks[0]

    // parse unique ID of the sensor
    sensor := string(blocks[1][:])

    // parse string repr. of voltage
    voltage_str := string(blocks[2][:])

    // parse string repr. of battery level
    battery_str := string(blocks[3][:])

    log.Printf("Sensor %s", sensor)
    log.Printf("Voltage %s", voltage_str)
    log.Printf("Battery %s", battery_str)

    // gps data are stored in 4th block
    packet := blocks[4]
    cycles := len(packet) / 13

    // loop through
    for i := 0; i < cycles; i++ {

        m := TGpsMeasurement{}

        position := i * 13;

        // lat - 4 bytes
        m.Latitude = float64(binary.LittleEndian.Uint32(packet[position + 5:])) / float64(10000000)

        // lng  - 4 bytes
        m.Longitude = float64(binary.LittleEndian.Uint32(packet[position + 9:])) / 10000000.0

        // satelites - 1 byte
        m.Satelites = int32(packet[position + 4])

        // proprietary timestamp - 4 bytes
        ttime  := binary.LittleEndian.Uint32(packet[position:])
        date := ttime / 100000
        ttime = ttime - date * 100000;

        year := date / 372
        date -= year * 372
        month := date / 31
        date -= month * 31
        day := date

        hour := ttime / 3600
        ttime -= hour * 3600
        min := ttime / 60
        ttime -= min * 60
        sec := ttime;

        // conversion to unix timestamp
        t := time.Date(int(year + 2000), time.Month(month), int(day), int(hour), int(min), int(sec), 0, time.UTC);
        m.Timestamp = t.Unix()

        log.Printf("datetime: %s", t)
        log.Printf("%v", m)
    }

    /* write request to text file
    err = ioutil.WriteFile("request.txt", body, 0644)
    if err != nil {
        log.Fatalf("Writing request body to file failed: %s", err)
    }
    log.Printf("Write request body passed")
    */
}
