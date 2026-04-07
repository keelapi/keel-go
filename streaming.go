package keel

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// parseSSEStream reads an SSE stream and sends events to the returned channel.
// The error channel receives at most one error. Both channels are closed when done.
func parseSSEStream(r io.ReadCloser) (<-chan SSEEvent, <-chan error) {
	events := make(chan SSEEvent)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)
		defer r.Close()

		scanner := bufio.NewScanner(r)
		var eventType string
		var dataLines []string

		for scanner.Scan() {
			line := scanner.Text()

			// Comment line
			if strings.HasPrefix(line, ":") {
				continue
			}

			// Empty line = dispatch event
			if line == "" {
				if len(dataLines) > 0 {
					data := strings.Join(dataLines, "\n")
					events <- SSEEvent{
						EventType: eventType,
						Data:      json.RawMessage(data),
					}
				}
				eventType = ""
				dataLines = nil
				continue
			}

			if strings.HasPrefix(line, "event:") {
				eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			} else if strings.HasPrefix(line, "data:") {
				dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			}
		}

		// Flush remaining event
		if len(dataLines) > 0 {
			data := strings.Join(dataLines, "\n")
			events <- SSEEvent{
				EventType: eventType,
				Data:      json.RawMessage(data),
			}
		}

		if err := scanner.Err(); err != nil {
			errc <- err
		}
	}()

	return events, errc
}
