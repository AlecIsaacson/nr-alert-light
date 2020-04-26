// Handles the work of driving the RaspberryPi GPIO pins.
// Takes a map of alerts (ID, severity, status).

package lights

import (
	"log"

	"github.com/stianeikeland/go-rpio"
)

func LightDriver (alertCount map[string]int) (err error){
	// Which light is on what pin on the Pi
	var critPin int = 18
	var warnPin int = 14

	if err := rpio.Open(); err != nil {
		log.Println("Failed to open RPI for IO", err)
		return err
	}

	//Yes, this code is repetitive and I probably could do something more elegant.
	for alertSeverity, count := range alertCount {
		switch alertSeverity {
		case "CRITICAL":
			log.Println("Evaluating critical alert count")
			if count > 0 {
				log.Println("Critical alert count > 0, light on.")
				pin := rpio.Pin(critPin)
				pin.Output()
				pin.High()
			} else if count == 0 {
				log.Println("Critical alert count = 0, light off.")
				pin := rpio.Pin(critPin)
				pin.Output()
				pin.Low()
			} else {
				log.Println("Unexpected value for critical alert count:", count)
			}
		case "WARNING":
			log.Println("Evaluating warning alert count")
			if count > 0 {
				log.Println("Warning alert count > 0, light on.")
				pin := rpio.Pin(warnPin)
				pin.Output()
				pin.High()
			} else if count == 0 {
				log.Println("Warning alert count = 0, light off.")
				pin := rpio.Pin(warnPin)
				pin.Output()
				pin.Low()
			} else {
				log.Println("Unexpected value for warning alert count:", count)
			}
		default:
			log.Println("Unexpected alert severity received:", alertSeverity)
		}
	}
	return err
}
