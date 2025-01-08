package es

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchAdapter is a custom adapter for sending logs to Elasticsearch.
type ElasticsearchAdapter struct {
	client *elasticsearch.Client
	index  string
}

// NewElasticsearchAdapter creates a new ElasticsearchAdapter instance.
func NewElasticsearchAdapter(client *elasticsearch.Client, index string) *ElasticsearchAdapter {
	return &ElasticsearchAdapter{
		client: client,
		index:  index,
	}
}

// Init initializes the adapter (required by LoggerInterface).
func (a *ElasticsearchAdapter) Init(config string) error {
	// No initialization needed for this example
	return nil
}

// WriteMsg sends a log message to Elasticsearch based on log level.
func (a *ElasticsearchAdapter) WriteMsg(msg *logs.LogMsg) error {
	// Determine log level based on the level of the msg.
	level := map[int]int{
		logs.LevelEmergency: 1,
		logs.LevelAlert:     2,
		logs.LevelCritical:  3,
		logs.LevelError:     4,
		logs.LevelWarning:   5,
		logs.LevelNotice:    6,
		logs.LevelInfo:      7,
		logs.LevelDebug:     8,
	}[msg.Level]

	// If no matching level found, default to "info"
	if level == 0 {
		level = 7
	}

	// Create a JSON document for the log
	logDoc := map[string]interface{}{
		"message":     msg.Msg,
		"level":       level,
		"timestamp":   msg.When.Format(time.RFC3339),
		"file_path":   msg.FilePath,
		"line_number": msg.LineNumber,
		"args":        fmt.Sprintf("%v", msg.Args), // Convert Args to string representation
		"prefix":      msg.Prefix,
	}

	// Marshal the log document to JSON
	docJSON, err := json.Marshal(logDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal log document: %w", err)
	}

	// Index the log document in Elasticsearch
	req := esapi.IndexRequest{
		Index: a.index,
		Body:  strings.NewReader(string(docJSON)),
	}

	// Execute the request
	res, err := req.Do(context.Background(), a.client)
	if err != nil {
		return fmt.Errorf("failed to send log to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	return nil
}

// Destroy cleans up resources (required by LoggerInterface).
func (a *ElasticsearchAdapter) Destroy() {
	// No cleanup needed for this example
}

// Flush flushes any buffered logs (required by LoggerInterface).
func (a *ElasticsearchAdapter) Flush() {
	// No buffering in this example
}

// SetFormatter sets a custom formatter for the logs (required by LoggerInterface).
func (a *ElasticsearchAdapter) SetFormatter(f logs.LogFormatter) {
	// No-op: This adapter does not use a custom formatter
}

func NewElasticsearchLogger() logs.Logger {
	// Initialize Elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"}, // Elasticsearch 8.x uses HTTPS
		Username:  "elastic",                         // Default user
		Password:  "helloworld",                      // Password for the elastic user
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logs.Error("Failed to create Elasticsearch client:", err)
		return nil
	}

	// Create and return a new ElasticsearchAdapter
	return NewElasticsearchAdapter(client, "beego-logs")
}
