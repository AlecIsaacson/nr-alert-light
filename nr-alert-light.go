// Listens for New Relic webhooks and turns on a status light when one is received.
//
package main

import (
	//"encoding/json"
	"os"
	"log"
	"flag"
	"net/http"
	"fmt"
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

func main() {
	logFile, err := os.OpenFile("nr-alert-light.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
	    log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("New Relic Alert Light v1.0")
	log.Println("")
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
	// For some reason, the below results in a can't convert string to byte error, while the StringVar type doesn't.
	//alertJSON := flag.String("alert", "", "Alert JSON")
	var alertJSON string
	flag.StringVar(&alertJSON, "alert", "", "Alert JSON")
	flag.Parse()

	if *logVerbose {
		log.Println("Verbose logging enabled.")
		log.Println("Received JSON: ",alertJSON)
	}

	http.HandleFunc("/", hookHandler)
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func hookHandler (resp http.ResponseWriter, req *http.Request) {
	fmt.Println(req)
	// var nrAlertInfo nrWebhookStruct
	// if err := json.Unmarshal([]byte(alertJSON), &nrAlertInfo); err != nil {
	// 	panic(err)
	// }
	//
	// if *logVerbose {
	// 	log.Println("Unmarshalling alert into struct")
	// 	log.Println(nrAlertInfo)
	// }
	//log.Println("New Incident:",nrAlertInfo.CurrentState, nrAlertInfo.IncidentID, nrAlertInfo.Severity)
}
