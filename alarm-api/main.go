package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

var client MQTT.Client
var blood_sugar float64 = 7
var logger Logger

type Logger struct {
	logLevel string
	logFunc  func(string)
}

func (l *Logger) Debug(message string) {
	if l.logLevel == "debug" {
		l.logFunc(message)
	}
}

func main() {
	init_broker()
	init_logger()

	// Create a Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Temporarily allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Cache-Control"}, // Include Cache-Control
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/insulin-alarm", insulin_alarm)
	r.Run()
}

func init_broker() {
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
}

func init_logger() {
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{"https://opensearch:9200"},
		Username:  "admin", // For testing only. Don't store credentials in code.
		Password:  "Hej123456789!",
	})
	if err != nil {
		panic(err)
	}

	logFunc := func(message string) {
		document := strings.NewReader(fmt.Sprintf(`{
			"message": "%s"
		}`, message))

		req := opensearchapi.IndexRequest{
			Index: "alarm-api-logs",
			Body:  document,
		}
		res, err := req.Do(context.Background(), client)
		if err != nil {
			fmt.Println(err)
		}
		if res.IsError() {
			fmt.Println(res.String())
		}
	}

	logger = Logger{
		logLevel: os.Getenv("LOG_LEVEL"),
		logFunc:  logFunc,
	}
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
	logger.Debug(fmt.Sprintf("Received blood sugar level: %f", blood_sugar))
}

type InsulinAlarmResponse struct {
	Level float64 `json:"level"` // JSON key for the blood sugar level
	State string  `json:"state"` // JSON key for the blood sugar state
}

func insulin_alarm(c *gin.Context) {
	var state string

	// Determine the blood sugar state based on the value
	switch {
	case blood_sugar < 4:
		state = "low"
	case blood_sugar < 9:
		state = "normal"
	case blood_sugar < 13:
		state = "high"
	default:
		state = "critical"
	}

	// Create a response object
	response := InsulinAlarmResponse{
		Level: blood_sugar,
		State: state,
	}

	// Respond with the object serialized as JSON
	c.JSON(http.StatusOK, response)
}
