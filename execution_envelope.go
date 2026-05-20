package keel

import (
	"encoding/json"
	"errors"
)

func executionResponseFromError(err error) (*ExecutionResponse, bool) {
	var keelErr *KeelError
	if !errors.As(err, &keelErr) || keelErr.Details == nil {
		return nil, false
	}

	data, marshalErr := json.Marshal(keelErr.Details)
	if marshalErr != nil {
		return nil, false
	}

	var envelope ExecutionResponse
	if unmarshalErr := json.Unmarshal(data, &envelope); unmarshalErr != nil {
		return nil, false
	}
	if envelope.Object != "execution" || envelope.ID == "" || envelope.Status == "" {
		return nil, false
	}
	return &envelope, true
}

func executionTerminalEventName(envelope *ExecutionResponse) string {
	switch envelope.Status {
	case "completed":
		return "execution.completed"
	case "denied":
		return "execution.denied"
	default:
		return "execution.error"
	}
}

func executionEnvelopeStreamEvents(envelope *ExecutionResponse) ([]ExecutionStreamEvent, error) {
	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	done, err := json.Marshal(map[string]any{
		"id":     envelope.ID,
		"object": "execution.done",
	})
	if err != nil {
		return nil, err
	}
	return []ExecutionStreamEvent{
		{EventType: executionTerminalEventName(envelope), Data: data},
		{EventType: "done", Data: done},
	}, nil
}
