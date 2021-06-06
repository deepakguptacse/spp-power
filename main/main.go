package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/distatus/battery"
	"go.uber.org/zap"
)

type PowerStatus string

const (
	PowerGenerator PowerStatus = "GEN"
	PowerMain      PowerStatus = "MAIN"
)

var powerStatus PowerStatus

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	defer logger.Sugar()

	initialStatusStr := os.Args[1]
	powerStatus = PowerMain
	if initialStatusStr == "gen" {
		powerStatus = PowerGenerator
	}

	go syncPowerStatus()

	// Setup HTTP server
	http.HandleFunc(
		"/",
		func(w http.ResponseWriter, req *http.Request) {
			greenColorPage := "<!DOCTYPE html>\n<html>\n   <head>\n      <title>HTML Backgorund Color</title>\n   </head>\n   <body style=\"background-color:green;\">\n      <h1>Power Status: </h1>\n      <p>Main</p>\n   </body>\n</html>"
			redColorPage := "<!DOCTYPE html>\n<html>\n   <head>\n      <title>HTML Backgorund Color</title>\n   </head>\n   <body style=\"background-color:red;\">\n      <h1>Power Status: </h1>\n      <p>Generator</p>\n   </body>\n</html>"
			response := greenColorPage
			if powerStatus == PowerGenerator {
				response = redColorPage
			}
			_, err := w.Write([]byte(response))
			if err != nil {
				zap.S().Errorf("Cannot send power status response: %v", err)
			}
		},
	)
	err := http.ListenAndServe(":80", nil)
	zap.S().Errorf("Server existed with error: %v", err)
}

func syncPowerStatus() {
	dischargingStartedTimestamp := time.Now()
	previousChargingStatus := chargingStatus()
	for {
		currentChargingStatus := chargingStatus()
		currentStatusTimestamp := time.Now()
		if previousChargingStatus != currentChargingStatus {
			zap.
				S().
				Infof(
					"Battery charging status changed - (%s, %s)",
					previousChargingStatus,
					currentChargingStatus,
				)
			if currentChargingStatus == battery.Discharging {
				dischargingStartedTimestamp = time.Now()
				zap.S().Infof("Discharging started at: %v", dischargingStartedTimestamp)
			} else if currentChargingStatus == battery.Charging {
				secondsElapsed := currentStatusTimestamp.Sub(dischargingStartedTimestamp).Seconds()
				zap.S().Infof("Found discharging for %f seconds", secondsElapsed)
				if secondsElapsed < 15 {
					powerStatus = PowerMain
				} else {
					powerStatus = PowerGenerator
				}
				updateStatus(string(powerStatus))
			}
		}
		previousChargingStatus = currentChargingStatus
		time.Sleep(1 * time.Second)
	}
}

func chargingStatus() battery.State {
	batteries, err := battery.GetAll()
	if err != nil {
		fmt.Println("Could not get battery info!")
		return battery.Charging
	}
	state := batteries[0].State
	if state == battery.Unknown {
		state = battery.Charging
	}
	return state
}

func updateStatus(newStatus string) {
	json := fmt.Sprintf("{\"source\": \"%s\"}", newStatus)
	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodPut,
		"https://spp-power-source-default-rtdb.asia-southeast1.firebasedatabase.app/power-source.json",
		bytes.NewBufferString(json),
	)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		zap.S().Errorf("Couldn't create request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Errorf("Couldn't update status: %v", err)
	}
	zap.S().Infof("Status update response: %v", resp)
}
