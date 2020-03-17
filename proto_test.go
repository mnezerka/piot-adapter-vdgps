package main

import (
    "os"
    "testing"
    "github.com/mnezerka/go-piot/test"
)

func TestProcessPacket(t *testing.T) {

    logger := test.GetLogger(t)

    f, err := os.Open("test_request.txt")
    test.Ok(t, err)
    defer f.Close()

    packet, err := ProtoParse(logger, f)
    test.Ok(t, err)

    test.Equals(t, len(packet.Locs), 69)
}
