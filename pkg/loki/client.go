package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client is a wrapper around the Loki HTTP API
// It provides helpers that convert Loki responses into JSON-friendly
// structures that are easy for frontends or LLMs to consume.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// LokiResponse represents the response structure from Loki queries
type LokiResponse struct {
	Status string   `json:"status"`
	Data   LokiData `json:"data"`
}

// LokiData contains the actual data from Loki response
type LokiData struct {
	ResultType string       `json:"resultType"`
	Result     []LokiStream `json:"result"`
}

// LokiStream represents a stream of logs with labels
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// LokiEntry represents a single log entry
type LokiEntry struct {
	Timestamp string `json:"timestamp"`
	Line      string `json:"line"`
}

// LabelResponse represents the response from label queries
type LabelResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

// New creates and initializes a Loki client bound to the given lokiURL
// Example lokiURL: "http://loki.monitoring:3100".
func New(lokiURL string) (*Client, error) {
	if lokiURL == "" {
		return nil, fmt.Errorf("loki URL is required")
	}

	// Parse URL to ensure it's valid
	_, err := url.Parse(lokiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid loki URL: %w", err)
	}

	return &Client{
		baseURL: lokiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// QueryInstant performs an instant query at the provided timestamp (or now if ts.IsZero()).
// It returns a map that contains "query", "warnings", "resultType" and a "result"
// that's JSON-serializable and friendly for LLM consumption.
func (c *Client) QueryInstant(ctx context.Context, query string, ts time.Time, limit int) (map[string]interface{}, error) {
	if c == nil {
		return nil, fmt.Errorf("loki client not initialized")
	}

	if ts.IsZero() {
		ts = time.Now()
	}

	// Build query parameters
	params := url.Values{}
	params.Add("query", query)
	params.Add("time", strconv.FormatInt(ts.UnixNano(), 10))
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	// Make request
	url := fmt.Sprintf("%s/loki/api/v1/query?%s", c.baseURL, params.Encode())
	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("loki instant query failed: %w", err)
	}

	// Parse response
	var lokiResp LokiResponse
	if err := json.Unmarshal(resp, &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to parse loki response: %w", err)
	}

	return map[string]interface{}{
		"query":      query,
		"timestamp":  ts,
		"resultType": lokiResp.Data.ResultType,
		"result":     convertLokiStreams(lokiResp.Data.Result),
		"status":     lokiResp.Status,
	}, nil
}

// QueryRange executes a range query between start and end with the given step.
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration, limit int) (map[string]interface{}, error) {
	if c == nil {
		return nil, fmt.Errorf("loki client not initialized")
	}

	// Build query parameters
	params := url.Values{}
	params.Add("query", query)
	params.Add("start", strconv.FormatInt(start.UnixNano(), 10))
	params.Add("end", strconv.FormatInt(end.UnixNano(), 10))
	params.Add("step", strconv.FormatInt(step.Milliseconds(), 10))
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	// Make request
	url := fmt.Sprintf("%s/loki/api/v1/query_range?%s", c.baseURL, params.Encode())
	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("loki range query failed: %w", err)
	}

	// Parse response
	var lokiResp LokiResponse
	if err := json.Unmarshal(resp, &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to parse loki response: %w", err)
	}

	return map[string]interface{}{
		"query":      query,
		"start":      start,
		"end":        end,
		"step":       step.String(),
		"resultType": lokiResp.Data.ResultType,
		"result":     convertLokiStreams(lokiResp.Data.Result),
		"status":     lokiResp.Status,
	}, nil
}

// GetLabels returns all available label names
func (c *Client) GetLabels(ctx context.Context, start, end time.Time) ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("loki client not initialized")
	}

	// Build query parameters
	params := url.Values{}
	if !start.IsZero() {
		params.Add("start", strconv.FormatInt(start.UnixNano(), 10))
	}
	if !end.IsZero() {
		params.Add("end", strconv.FormatInt(end.UnixNano(), 10))
	}

	// Make request
	url := fmt.Sprintf("%s/loki/api/v1/labels?%s", c.baseURL, params.Encode())
	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	// Parse response
	var labelResp LabelResponse
	if err := json.Unmarshal(resp, &labelResp); err != nil {
		return nil, fmt.Errorf("failed to parse labels response: %w", err)
	}

	return labelResp.Data, nil
}

// GetLabelValues returns all possible values for a specific label
func (c *Client) GetLabelValues(ctx context.Context, label string, start, end time.Time) ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("loki client not initialized")
	}

	// Build query parameters
	params := url.Values{}
	if !start.IsZero() {
		params.Add("start", strconv.FormatInt(start.UnixNano(), 10))
	}
	if !end.IsZero() {
		params.Add("end", strconv.FormatInt(end.UnixNano(), 10))
	}

	// Make request - 使用正确的 API 端点：/label/{label}/values
	url := fmt.Sprintf("%s/loki/api/v1/label/%s/values?%s", c.baseURL, label, params.Encode())
	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get label values: %w", err)
	}

	// Parse response
	var labelResp LabelResponse
	if err := json.Unmarshal(resp, &labelResp); err != nil {
		return nil, fmt.Errorf("failed to parse label values response: %w", err)
	}

	return labelResp.Data, nil
}

// SeriesResponse represents the response from series queries
type SeriesResponse struct {
	Status string              `json:"status"`
	Data   []map[string]string `json:"data"`
}

// GetStreams returns all log streams that match the given selector
func (c *Client) GetStreams(ctx context.Context, selector string, start, end time.Time) (map[string]interface{}, error) {
	if c == nil {
		return nil, fmt.Errorf("loki client not initialized")
	}

	// Build query parameters
	params := url.Values{}
	if selector != "" {
		params.Add("match", selector)
	}
	if !start.IsZero() {
		params.Add("start", strconv.FormatInt(start.UnixNano(), 10))
	}
	if !end.IsZero() {
		params.Add("end", strconv.FormatInt(end.UnixNano(), 10))
	}

	// Make request
	url := fmt.Sprintf("%s/loki/api/v1/series?%s", c.baseURL, params.Encode())
	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get streams: %w", err)
	}

	// Parse response using SeriesResponse struct
	var seriesResp SeriesResponse
	if err := json.Unmarshal(resp, &seriesResp); err != nil {
		return nil, fmt.Errorf("failed to parse streams response: %w", err)
	}

	return map[string]interface{}{
		"selector": selector,
		"start":    start,
		"end":      end,
		"result":   seriesResp.Data,
		"status":   seriesResp.Status,
	}, nil
}

// makeRequest is a helper function to make HTTP requests to Loki
func (c *Client) makeRequest(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("loki API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// convertLokiStreams converts Loki streams to JSON-friendly structures
func convertLokiStreams(streams []LokiStream) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(streams))

	for _, stream := range streams {
		entries := make([]map[string]interface{}, 0, len(stream.Values))

		for _, entry := range stream.Values {
			if len(entry) != 2 {
				continue // skip invalid entries
			}

			timestampStr := entry[0]
			line := entry[1]

			// Parse timestamp from string (Loki returns RFC3339 format)
			timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
			if err != nil {
				// If parsing fails, use the original string
				timestamp = time.Time{}
			}

			entries = append(entries, map[string]interface{}{
				"timestamp": timestamp,
				"line":      line,
			})
		}

		out = append(out, map[string]interface{}{
			"labels":  stream.Stream,
			"entries": entries,
		})
	}

	return out
}
