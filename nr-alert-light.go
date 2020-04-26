// Listens for New Relic webhooks and turns on a status light when one is received.
// Written by AI of New Relic 4/16/2020
//
package main

import (
	"encoding/json"
	"os"
	"log"
	"flag"
	"net/http"
	"fmt"
	"os/signal"
	"syscall"

	"lights"
)

var appVersion = "2.0"

//New Relic webhooks look like this.
type nrWebhookStruct struct {
	AccountID                     int         `json:"account_id"`
	AccountName                   string      `json:"account_name"`
	ClosedViolationsCountCritical int         `json:"closed_violations_count_critical"`
	ClosedViolationsCountWarning  int         `json:"closed_violations_count_warning"`
	ConditionFamilyID             int         `json:"condition_family_id"`
	ConditionID                   int         `json:"condition_id"`
	ConditionName                 string      `json:"condition_name"`
	CurrentState                  string      `json:"current_state"`
	Details                       string      `json:"details"`
	Duration                      int         `json:"duration"`
	EventType                     string      `json:"event_type"`
	IncidentAcknowledgeURL        string      `json:"incident_acknowledge_url"`
	IncidentID                    int         `json:"incident_id"`
	IncidentURL                   string      `json:"incident_url"`
	OpenViolationsCountCritical   int         `json:"open_violations_count_critical"`
	OpenViolationsCountWarning    int         `json:"open_violations_count_warning"`
	Owner                         string      `json:"owner"`
	PolicyName                    string      `json:"policy_name"`
	PolicyURL                     string      `json:"policy_url"`
	RunbookURL                    interface{} `json:"runbook_url"`
	Severity                      string      `json:"severity"`
	Targets                       []struct {
		ID     string `json:"id"`
		Labels struct {
			Account           string `json:"account"`
			Environment       string `json:"environment"`
			ExtMyKey          string `json:"extMyKey"`
			FullHostname      string `json:"fullHostname"`
			Hostname          string `json:"hostname"`
			InstanceType      string `json:"instanceType"`
			LinuxDistribution string `json:"linuxDistribution"`
			Service           string `json:"service"`
		} `json:"labels"`
		Link    string `json:"link"`
		Name    string `json:"name"`
		Product string `json:"product"`
		Type    string `json:"type"`
	} `json:"targets"`
	Timestamp            int64       `json:"timestamp"`
	ViolationCallbackURL string      `json:"violation_callback_url"`
	ViolationChartURL    interface{} `json:"violation_chart_url"`
}

// Map to store open alerts. Outer map holds severity (i.e. critical, warning), inner holds
// incident ID and status (i.e. open)
var alerts = make(map[string]map[int]string)

// Which light is on what pin on the Pi
var critPin int = 18
var warnPin int = 14

func main() {
	//Initialze logging
	logFile, err := os.OpenFile("nr-alert-light.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
	    log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//Setup startup flags.
	log.Println("")
	log.Println("New Relic Alert Light v",appVersion)
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging.")
	serverPort := flag.String("listen", "9000", "Port to listen on.")
	flag.Parse()

	if *logVerbose {
		log.Println("Verbose logging enabled.")
	}

	//Setup our shutdown handler
	shutdownHandler()

	//Initialize the inner maps for the open alert tracker.
	alerts["CRITICAL"] = make(map[int]string)
	alerts["WARNING"] = make(map[int]string)

	//Launch HTTP listener.  Blocks until program end.
	http.HandleFunc("/info/", infoHandler)
	http.HandleFunc("/hook", hookHandler)
	log.Fatal(http.ListenAndServe(":"+*serverPort, nil))
}

//Handle incoming webhooks from NR.
func hookHandler (resp http.ResponseWriter, req *http.Request) {
	log.Println("Attempting to process webhook")
	var nrAlertInfo nrWebhookStruct
	if err := json.NewDecoder(req.Body,).Decode(&nrAlertInfo); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		log.Println("Malformed request received:", err.Error())
		return
	}
	log.Println("New Incident:",nrAlertInfo.Severity, nrAlertInfo.IncidentID, nrAlertInfo.CurrentState)
	alertCount := alertTracker(nrAlertInfo.Severity, nrAlertInfo.IncidentID, nrAlertInfo.CurrentState)
	if err := lights.LightDriver(alertCount); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		log.Println("GPIO error, unable to set lights.")
	}
}

//Take valid hooks, extract the info we need, then account for and return the number of open crit and warn alerts.
//We keep an in-memory struct of open critical and warn alerts by alert ID.  When an alert is closed, we remove it from the struct.
func alertTracker (nrSeverity string, nrIncidentID int, nrCurrentState string) (map[string]int) {
	log.Println ("Processing Incident: ", nrSeverity, nrIncidentID, nrCurrentState)
	switch nrCurrentState {
	case "open":
		log.Println("Received alert open message")
		if _, alertExists := alerts[nrSeverity][nrIncidentID]; alertExists {
			log.Println("Alert exists, nothing to do.")
		} else {
			log.Println("Alert does not exist, adding to map.")
			alerts[nrSeverity][nrIncidentID] = nrCurrentState
		}
	case "closed":
		log.Println("Received alert close message")
		if _, alertExists := alerts[nrSeverity][nrIncidentID]; alertExists {
			log.Println("Alert exists, removing.")
			delete(alerts[nrSeverity], nrIncidentID)
		} else {
			log.Println("Trying to close an non-existent alert, this shouldn't happen often.")
		}
	default:
		log.Println("Received alert with unexpected current state")
	}
	log.Println("Open Critical:", len(alerts["CRITICAL"]), "Open Warning:", len(alerts["WARNING"]))
	alertCount := map[string]int{"CRITICAL": len(alerts["CRITICAL"]), "WARNING": len(alerts["WARNING"])}
	return alertCount
}

//
//If someone calls the /info URL, send back some info.
func infoHandler (resp http.ResponseWriter, req *http.Request) {
	log.Println("Info request received")
	title := "Info Page"
	fmt.Fprintf(resp, "<h1>%s</h1><div>Version: %v</div><div>Critical Alerts %v</div><div>Warning Alerts %v</div>", title, appVersion, alerts["CRITICAL"], alerts["WARNING"])
}

//If any lights are on when the app ends, they'll stay on. This fixes that.
func shutdownHandler() {
	shutdownChannel := make(chan os.Signal)
	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownChannel
		log.Println("Shutdown signal received, turning off lights")
		if err := rpio.Open(); err != nil {
			log.Println("Failed to open RPI for IO", err)
			os.Exit(1)
		}
		pin := rpio.Pin(warnPin)
		pin.Output()
		pin.Low()

		pin = rpio.Pin(critPin)
		pin.Output()
		pin.Low()

		os.Exit(0)
	}()
}
