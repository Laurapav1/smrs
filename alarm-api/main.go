package main

import (
	"fmt"
	"net/http"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

var client MQTT.Client
var blood_sugar float64

func main() {
	broker := "tcp://broker:1883"
	clientID := "go_mqtt_subscriber"

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

	client.Subscribe("test/topic", 0, message_received)

	r := gin.Default()
	r.GET("/insulin-alarm", insulin_alarm)
	r.Run()
}

func message_received(client MQTT.Client, message MQTT.Message) {
	fmt.Println(message.Payload())
	str := string(message.Payload())
	blood_sugar, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	fmt.Println(blood_sugar)
}

func insulin_alarm(c *gin.Context) {
	if blood_sugar > 10 {
		c.JSON(http.StatusOK, gin.H{
			"message": "blood sugar too high",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "blood sugar is okay",
	})
}
