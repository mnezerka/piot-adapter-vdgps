package handler_test

import (
    "net/http"
    "net/http/httptest"
    "os"
    //"strings"
    "testing"
    "piot-adapter-vdgps/handler"
    "piot-adapter-vdgps/test"
)

/* GET method is not supported */
func TestForbiddenGet(t *testing.T) {

    f, err := os.Open("test_request.txt")
    test.Ok(t, err)
    defer f.Close()

    //req, err := http.NewRequest("GET", "/", strings.NewReader(""))
    req, err := http.NewRequest("POST", "/", f)
    test.Ok(t, err)
    //req = req.WithContext(test.CreateTestContext())

    rr := httptest.NewRecorder()
    handler.Adapter(rr, req)

    test.CheckStatusCode(t, rr, 200)
}


/* Post data for device that is not registered  */
/*
func TestPacketForUnknownThing(t *testing.T) {
    ctx := test.CreateTestContext()

    test.CleanDb(t, ctx)

    deviceData := `
    {
        "device": "Device123",
        "readings": [
            {
                "address": "SensorXYZ",
                "t": 23
            }
        ]
    }`

    req, err := http.NewRequest("POST", "/", strings.NewReader(deviceData))
    test.Ok(t, err)

    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()

    handler := handler.Adapter{}

    handler.ServeHTTP(rr, req)

    test.CheckStatusCode(t, rr, 200)

    // TODO: Check if device is registered
}
*/
