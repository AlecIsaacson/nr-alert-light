# nr-alert-light
This code receives New Relic Alerts webhooks and triggers a Raspberry Pi GPIO that can be used to turn on a light or take some other action.

Your webhook must be directed to the /hook endpoint. By default the app listens on port 9000.
