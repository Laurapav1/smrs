package main

import (
	"net/http"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

var client MQTT.Client

func main() {
	broker := "tcp://broker:1883"
	clientID := "go_mqtt_publisher"

	// Opret MQTT-klientens forbindelsesindstillinger
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)

	// Opret en ny MQTT-klient
	client = MQTT.NewClient(opts)

	// Opret forbindelse til brokeren
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	r := gin.Default()
	r.POST("/blood_sugar_request", blood_sugar_request)
	r.Run()
}

func blood_sugar_request(c *gin.Context) {
	bloodsugar := c.Query("blood_sugar")
	token := client.Publish("test/topic", 0, false, bloodsugar)
	token.Wait()

	c.Status(http.StatusOK)
}
