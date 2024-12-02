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
	Message    string  `json:"message"`
	BloodSugar float64 `json:"blood_sugar"`
	Timestamp  string  `json:"timestamp"`
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
	r.GET("/logs", fetch_logs)
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
			Message:    message,
			BloodSugar: bloodSugar,
			Timestamp:  time.Now().Format(time.RFC3339),
		}

		// Marshal the struct to JSON
		document, err := json.Marshal(entry)
		if err != nil {
			fmt.Println("Error marshaling log entry:", err)
			return
		}

		// Index the JSON document in OpenSearch
		_, err = opensearchClient.Index(context.Background(), opensearchapi.IndexReq{
			Index: "blood-sugar-logs",
			Body:  strings.NewReader(string(document)),
		})
		if err != nil {
			fmt.Println("Error indexing document:", err)
		}
	}

	_, err := opensearchClient.Indices.Create(context.Background(), opensearchapi.IndicesCreateReq{
		Index: "blood-sugar-logs",
		Body: strings.NewReader(`{
			"mappings": {
				"properties": {
					"message": {
						"type": "text"
					},
					"blood_sugar": {
						"type": "float"
					},
					"timestamp": {
						"type": "date"
					}
				}
			}
		}`),
	})
	if err != nil && !strings.Contains(err.Error(), "resource_already_exists_exception") {
		panic(err)
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
	Level float64 `json:"level"` // JSON key for the blood sugar level
	State string  `json:"state"` // JSON key for the blood sugar state
}

func insulin_alarm(c *gin.Context) {
	var state string

	// Determine the blood sugar state based on the value
	switch {
	case bloodSugar < 4:
		state = "low"
	case bloodSugar < 9:
		state = "normal"
	case bloodSugar < 13:
		state = "high"
	default:
		state = "critical"
	}

	// Create a response object
	response := InsulinAlarmResponse{
		Level: bloodSugar,
		State: state,
	}

	// Respond with the object serialized as JSON
	c.JSON(http.StatusOK, response)
}

func fetch_logs(c *gin.Context) {
	res, err := opensearchClient.Search(context.Background(), &opensearchapi.SearchReq{
		Indices: []string{"blood-sugar-logs"},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	// Collect log entries
	logs := []LogEntry{}
	for _, hit := range res.Hits.Hits {
		var entry LogEntry
		if err := json.Unmarshal(hit.Source, &entry); err != nil {
			fmt.Println("Error decoding hit:", err)
			continue
		}
		logs = append(logs, entry)
	}

	fmt.Println("Fetched logs:", logs)

	// Respond with the parsed logs
	c.JSON(http.StatusOK, logs)
}
