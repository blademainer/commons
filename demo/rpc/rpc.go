package main

import (
	"github.com/blademainer/commons/pkg/mqtt"
	"net/url"
	"time"
)

func main() {
	u, e := url.Parse("tcp://127.0.0.1:8001")
	if e != nil{
		panic(e)
	}
	client, e := mqtt.CreateClientByUri("test", u, 5*time.Second)

}



