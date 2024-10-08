package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var client MQTT.Client
var blood_sugar float64 = 7
var logger Logger

type Logger struct {
	logLevel string
}

func (l *Logger) Debug(str string, a ...any) {
	if l.logLevel == "debug" {
		fmt.Printf(str+"\n", a...)
	}
}

func main() {
	logger = Logger{
		logLevel: os.Getenv("LOG_LEVEL"),
	}

	broker := "tcp://broker:1883"
	clientID := "go_mqtt_subscriber"

	// MQTT client options
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)

	// Create a new MQTT client
	client = MQTT.NewClient(opts)

	// Connect to the broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	client.Subscribe("test/topic", 0, message_received)
	logger.Debug("Successfully subscribed to topic: test/topic")

	// Create a Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Change to your frontend's URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/insulin-alarm", insulin_alarm)
	r.Run()
}

func message_received(client MQTT.Client, message MQTT.Message) {
	// Convert message payload to string
	str := string(message.Payload())

	// Parse the blood sugar level from the message payload
	parsedBloodSugar, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Println("Error parsing blood sugar value:", err)
		return
	}

	// Update the global blood_sugar variable
	blood_sugar = parsedBloodSugar

	// Log the received blood sugar value
	logger.Debug("Received blood sugar level: %f", blood_sugar)
}

func insulin_alarm(c *gin.Context) {
	if blood_sugar < 4 {
		c.JSON(http.StatusOK, gin.H{
			"message": "blood sugar is too low",
		})
		return
	}

	if blood_sugar < 9 {
		c.JSON(http.StatusOK, gin.H{
			"message": "blood sugar is normal",
		})
		return
	}

	if blood_sugar < 13 {
		c.JSON(http.StatusOK, gin.H{
			"message": "blood sugar is too high",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "blood sugar is critical",
	})
}
