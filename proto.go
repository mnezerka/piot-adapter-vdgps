package main

import (
    "errors"
    "bytes"
    "encoding/binary"
    "io"
    "io/ioutil"
    "strconv"
    "time"
    "github.com/op/go-logging"
    //"github.com/mnezerka/go-piot"
)

const PREFIX = "VD"

type ProtoPacket struct {
    Id string
    Voltage float64
    Locs []*GpsMeasurement
}

type GpsMeasurement struct {
    Timestamp int64  `json:"ts"`
    Latitude float64 `json:"lat"`
    Longitude float64 `json:"lng"`
    Satelites int32 `json:"sat"`
}

func BitCheck(b byte, nbit uint) bool {
    return  (b & (1 << nbit)) > 0
}

func ProtoParse(log *logging.Logger, data io.Reader) (*ProtoPacket, error) {

    body, err := ioutil.ReadAll(data)
    if err != nil {
        log.Errorf("Reading request body failed: %s", err)
        return nil, err
    }
    log.Debugf("Read request body passed (size: %d)", len(body))

    // Split body to blocks (separator is "|")
    // last slice should contain rest of the body
    blocks := bytes.SplitN(body, []byte("|"), 5)

    log.Debugf("Detected %d blocks", len(blocks))

    if len(blocks) < 4 {
        return nil, errors.New("Invalid number of blocks")
    }

    // we don't need key
    //key := blocks[0]

    result := ProtoPacket{}

    // parse unique ID of the sensor
    result.Id = PREFIX + string(blocks[1][:])
    log.Debugf("Sensor %s", result.Id)

    // parse string repr. of voltage
    result.Voltage, err = strconv.ParseFloat(string(blocks[2][:]), 64)
    if err != nil {
        log.Errorf("Reading request body failed on parsing voltage block: %s", err)
        return nil, err
    }
    log.Debugf("Voltage %f", result.Voltage)

    // parse flags
    flags := blocks[3][0]
    log.Debugf("Flags %d %t", flags, flags)
    log.Debugf("  bit0: %v", BitCheck(flags, 0))
    log.Debugf("  bit1: %v", BitCheck(flags, 1))
    log.Debugf("  bit2: %v", BitCheck(flags, 2))
    log.Debugf("  bit3: %v", BitCheck(flags, 3))
    log.Debugf("  bit4: %v", BitCheck(flags, 4))
    log.Debugf("  bit5: %v", BitCheck(flags, 5))
    log.Debugf("  bit6: %v", BitCheck(flags, 6))
    log.Debugf("  bit7: %v", BitCheck(flags, 7))

    // gps data are stored in 4th block
    packet := blocks[4]
    cycles := len(packet) / 13

    log.Debugf("Packet has %d cycles", cycles)

    // loop through
    for i := 0; i < cycles; i++ {

        m := GpsMeasurement{}

        position := i * 13;

        // lat - 4 bytes
        m.Latitude = float64(binary.LittleEndian.Uint32(packet[position + 9:])) / 10000000.0

        // lng  - 4 bytes
        m.Longitude = float64(binary.LittleEndian.Uint32(packet[position + 5:])) / 10000000.0

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

        //log.Printf("datetime: %s", t)
        //log.Printf("%v", m)

        result.Locs = append(result.Locs, &m)
    }
    return &result, nil

    /* write request to text file
    err = ioutil.WriteFile("request.txt", body, 0644)
    if err != nil {
        log.Fatalf("Writing request body to file failed: %s", err)
    }
    log.Printf("Write request body passed")
    */
}
