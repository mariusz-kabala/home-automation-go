package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var mqttClient mqtt.Client
var authToken string

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func handler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")

	if auth == "" {
		token, ok := r.URL.Query()["token"]

		log.Info("Authorization token not provided, checking get params")

		if ok && len(token) == 1 {
			auth = string(token[0])
		}
	}

	if auth != authToken {
		log.Warnf("Invalid token provided, ip: %s", r.RemoteAddr)
		w.WriteHeader(401)
		return
	}

	log.Infof("Triggering goodNight scenario, ip: %s", r.RemoteAddr)

	mqttClient.Publish("home/scenario/goodNight/run", byte(0), false, "")
	w.WriteHeader(200)
}

func main() {
	godotenv.Load()

	log.SetOutput(os.Stdout)

	authToken = os.Getenv("AUTH_TOKEN")

	if authToken == "" {
		panic("Set auth token!")
	}

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%s", os.Getenv("MQTT_HOST"), os.Getenv("MQTT_PORT"))).SetClientID("webhook")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	mqttClient = mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	http.HandleFunc("/webhook/goodNight", handler)

	log.Info("Webserver started")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}
