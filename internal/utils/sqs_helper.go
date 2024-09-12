package utils

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

// ReadSQSEventFromFile lee un archivo JSON y lo parsea a un evento de SQS.
func ReadSQSEventFromFile(filePath string) (events.SQSEvent, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return events.SQSEvent{}, err
	}

	var sqsEvent events.SQSEvent
	if err := json.Unmarshal(data, &sqsEvent); err != nil {
		return events.SQSEvent{}, err
	}

	return sqsEvent, nil
}
