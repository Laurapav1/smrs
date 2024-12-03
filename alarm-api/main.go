package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

var client MQTT.Client
var opensearchClient *opensearchapi.Client
var bloodSugar float64 = 7
var logger Logger

type Logger struct {
	logLevel string
	logFunc  func(string)
}

type LogEntry struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func (l *Logger) Debug(message string) {
	if l.logLevel == "debug" {
		l.logFunc(message)
	}
}

func main() {
	init_broker()
	init_opensearch_client()
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

func init_opensearch_client() {
	var err error
	opensearchClient, err = opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Addresses: []string{"https://opensearch:9200"},
			Username:  "admin", // For testing only. Don't store credentials in code.
			Password:  "Hej123456789!",
		}},
	)
	if err != nil {
		panic(err)
	}
}

func init_logger() {
	logFunc := func(message string) {
		entry := LogEntry{
			Message:   message,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// Marshal the struct to JSON
		document, err := json.Marshal(entry)
		if err != nil {
			fmt.Println("Error marshaling log entry:", err)
			return
		}

		// Index the JSON document in OpenSearch
		_, err = opensearchClient.Index(context.Background(), opensearchapi.IndexReq{
			Index: "logs",
			Body:  strings.NewReader(string(document)),
		})
		if err != nil {
			fmt.Println("Error indexing document:", err)
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
	bloodSugar = parsedBloodSugar

	// Log the received blood sugar value
	logger.Debug(fmt.Sprintf("Received blood sugar level: %f", bloodSugar))
}

type InsulinAlarmResponse struct {
	Level float64 `json:"level"`
	State string  `json:"state"`
	Age   int     `json:"age"`
}

type Thresholds struct {
	Low      float64
	Normal   float64
	High     float64
	Critical float64
}

func determineState(level float64, thresholds Thresholds) string {
	switch {
	case level < thresholds.Low:
		return "low"
	case level < thresholds.Normal:
		return "normal"
	case level < thresholds.High:
		return "high"
	default:
		return "critical"
	}
}

func getThresholdsByAge(age int) Thresholds {
	switch {
	case age < 18:
		// Children thresholds
		return Thresholds{Low: 3.5, Normal: 7.8, High: 11.1, Critical: 11.1}
	case age <= 65:
		// Adults thresholds
		return Thresholds{Low: 4, Normal: 9, High: 13, Critical: 13}
	default:
		// Elderly thresholds
		return Thresholds{Low: 4.5, Normal: 10, High: 14, Critical: 14}
	}
}

func insulin_alarm(c *gin.Context) {
	// Parse age from query parameters
	ageStr := c.Query("age")
	age := 30
	if ageStr != "" {
		fmt.Sscanf(ageStr, "%d", &age)
	}

	// Create a response object
	response := InsulinAlarmResponse{
		Level: bloodSugar,
		State: determineState(bloodSugar, getThresholdsByAge(age)),
		Age:   age,
	}

	// Respond with the object serialized as JSON
	c.JSON(http.StatusOK, response)
}
