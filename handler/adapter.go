package handler

import (
    "io/ioutil"
    "log"
    "net/http"
)

func Adapter(w http.ResponseWriter, r *http.Request) {

    log.Printf("Request uri: %s,  method: %s, content length: %d", r.RequestURI, r.Method, r.ContentLength)

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("Reading request body error: %s", err)
    }
    log.Printf("Read request body passed")

    err = ioutil.WriteFile("request.txt", body, 0644)
    if err != nil {
        log.Fatalf("Writing request body to file failed: %s", err)
    }
    log.Printf("Write request body passed")
}
