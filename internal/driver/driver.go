package driver

import (
	"fmt"
	"sync"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gin-gonic/gin"
)

type Driver struct {
	lc      logger.LoggingClient
	asyncCh chan<- *dsModels.AsyncValues
}

type alert struct {
	camara    string
	timestamp string
}

var driver *Driver
var once sync.Once
var locker sync.Mutex
var alertList []alert

func init() {
	go pdService()
}

// NewProtocolDriver initializes the singleton Driver and
// returns it to the caller
func NewProtocolDriver() dsModels.ProtocolDriver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}

// Initialize performs protocol-specific initialization for the device service.
func (s *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) error {

	s.lc = lc
	s.asyncCh = asyncCh

	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *Driver) HandleReadCommands(deviceName string,
	protocols map[string]contract.ProtocolProperties,
	reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {

	res = make([]*dsModels.CommandValue, len(reqs))

	for i, req := range reqs {
		switch req.DeviceResourceName {
		case "Alert":
			rList := readDetectionAlert()
			timeList := []int64{}
			for i := range rList {
				t, err := time.Parse(time.RFC3339, rList[i].timestamp)
				if err != nil {
					s.lc.Error(err.Error())
					return res, err
				}
				tt := t.UnixNano()
				timeList = append(timeList, tt)
			}
			cv, err := dsModels.NewInt64ArrayValue(req.DeviceResourceName, 0, timeList)
			if err != nil {
				s.lc.Error(err.Error())
				return res, err
			}
			res[i] = cv
		}
	}
	return res, nil
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *Driver) HandleWriteCommands(deviceName string,
	protocols map[string]contract.ProtocolProperties,
	reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *Driver) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debug(fmt.Sprintf("Driver.Stop called: force=%v", force))
	}
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *Driver) AddDevice(deviceName string,
	protocols map[string]contract.ProtocolProperties,
	adminState contract.AdminState) error {
	s.lc.Debug(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *Driver) UpdateDevice(deviceName string,
	protocols map[string]contract.ProtocolProperties,
	adminState contract.AdminState) error {
	s.lc.Debug(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *Driver) RemoveDevice(deviceName string,
	protocols map[string]contract.ProtocolProperties) error {
	s.lc.Debug(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (s *Driver) Discover() {}

func pdService() {
	router := gin.Default()
	router.POST("/detection/alert", saveDetectionAlert)
	router.Run(":8888") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func saveDetectionAlert(context *gin.Context) {
	locker.Lock()
	defer locker.Unlock()
	a := alert{}
	a.camara = context.PostForm("camera")
	a.timestamp = context.PostForm("time")
	alertList = append(alertList, a)
}

func readDetectionAlert() (r []alert) {
	locker.Lock()
	defer locker.Unlock()
	t := []alert{}
	r = alertList
	alertList = t
	return
}
