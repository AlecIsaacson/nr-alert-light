// Listens for New Relic webhooks and turns on a status light when one is received.
//
package main

import (
	"encoding/json"
	"os"
	"log"
	"flag"
	"net/http"
	//"fmt"
)

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

func main() {
	//Initialze logging
	logFile, err := os.OpenFile("nr-alert-light.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
	    log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("")
	log.Println("New Relic Alert Light v1.0")
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
	flag.Parse()

	if *logVerbose {
		log.Println("Verbose logging enabled.")
	}

	//Initialize the inner maps.
	alerts["CRITICAL"] = make(map[int]string)
	alerts["WARNING"] = make(map[int]string)

	//Launch HTTP listener
	http.HandleFunc("/", hookHandler)
	log.Fatal(http.ListenAndServe(":9000", nil))
}

//Handle incoming hooks.
func hookHandler (resp http.ResponseWriter, req *http.Request) {
	var nrAlertInfo nrWebhookStruct
	if err := json.NewDecoder(req.Body,).Decode(&nrAlertInfo); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		log.Println("Malformed request received:", err.Error())
		return
	}
	log.Println("New Incident:",nrAlertInfo.Severity, nrAlertInfo.IncidentID, nrAlertInfo.CurrentState)
	lightDriver(nrAlertInfo.Severity, nrAlertInfo.IncidentID, nrAlertInfo.CurrentState)
}

//Take valid hooks and set the lights.
func lightDriver (nrSeverity string, nrIncidentID int, nrCurrentState string) {
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
			log.Println("Alert does not exist, this shouldn't happen.")
		}
	default:
		log.Println("Received alert with unexpected current state")
	}
	log.Println("Open Critical:", len(alerts["CRITICAL"]), "Open Warning:", len(alerts["WARNING"]))
}
