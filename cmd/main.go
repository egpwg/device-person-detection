package main

import (
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
	"github.com/egpwg/device-person-detection/internal/driver"
)

const serviceName string = "device-person-detection"

func main() {
	sd := driver.Driver{}
	startup.Bootstrap(serviceName, "0.0.1", &sd)
}
