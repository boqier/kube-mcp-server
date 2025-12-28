package prometheus

import (
	"context"
	"fmt"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client is a thin wrapper around the Prometheus HTTP API (prometheus/v1)
// It provides helpers that convert Prometheus model.Value into JSON-friendly
// structures that are easy for frontends or LLMs to consume.
type Client struct {
	api promv1.API
}

// New creates and initializes a Prometheus client bound to the given promURL
// Example promURL: "http://prometheus.monitoring:9090".
func New(promURL string) (*Client, error) {
	if promURL == "" {
		return nil, fmt.Errorf("prometheus URL is required")
	}
	apiClient, err := promapi.NewClient(promapi.Config{Address: promURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus http client: %w", err)
	}
	return &Client{api: promv1.NewAPI(apiClient)}, nil
}

// QueryInstant performs an instant query at the provided timestamp (or now if ts.IsZero()).
// It returns a map that contains "query", "warnings", "resultType" and a "result"
// that's JSON-serializable and friendly for LLM consumption.
func (c *Client) QueryInstant(ctx context.Context, query string, ts time.Time) (map[string]interface{}, error) {
	if c == nil || c.api == nil {
		return nil, fmt.Errorf("prometheus client not initialized")
	}
	if ts.IsZero() {
		ts = time.Now()
	}
	val, warnings, err := c.api.Query(ctx, query, ts)
	if err != nil {
		return nil, fmt.Errorf("prometheus instant query failed: %w", err)
	}
	return map[string]interface{}{
		"query":      query,
		"timestamp":  ts,
		"warnings":   warnings,
		"resultType": string(val.Type()),
		"result":     convertPromModelValue(val),
	}, nil
}

// QueryRange executes a range query between start and end with the given step.
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (map[string]interface{}, error) {
	if c == nil || c.api == nil {
		return nil, fmt.Errorf("prometheus client not initialized")
	}
	r := promv1.Range{Start: start, End: end, Step: step}
	val, warnings, err := c.api.QueryRange(ctx, query, r)
	if err != nil {
		return nil, fmt.Errorf("prometheus range query failed: %w", err)
	}
	return map[string]interface{}{
		"query":      query,
		"start":      start,
		"end":        end,
		"step":       step.String(),
		"warnings":   warnings,
		"resultType": string(val.Type()),
		"result":     convertPromModelValue(val),
	}, nil
}

// convertPromModelValue converts github.com/prometheus/common/model.Value into
// simple JSON-friendly structures (maps/slices/primitives).
func convertPromModelValue(v model.Value) interface{} {
	switch vv := v.(type) {
	case model.Vector:
		out := []map[string]interface{}{}
		for _, s := range vv {
			out = append(out, map[string]interface{}{
				"metric":    metricToMap(s.Metric),
				"value":     s.Value.String(),
				"timestamp": s.Timestamp.Time(),
			})
		}
		return out
	case model.Matrix:
		out := []map[string]interface{}{}
		for _, stream := range vv {
			vals := []map[string]interface{}{}
			for _, p := range stream.Values {
				vals = append(vals, map[string]interface{}{
					"value":     p.Value.String(),
					"timestamp": p.Timestamp.Time(),
				})
			}
			out = append(out, map[string]interface{}{
				"metric": metricToMap(stream.Metric),
				"values": vals,
			})
		}
		return out
	case *model.Scalar:
		return map[string]interface{}{
			"value":     vv.Value.String(),
			"timestamp": vv.Timestamp.Time(),
		}
	case *model.String:
		return map[string]interface{}{
			"value": string(vv.Value),
		}
	default:
		// fallback to string representation
		return vv.String()
	}
}

func metricToMap(m model.Metric) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		out[string(k)] = string(v)
	}
	return out
}

func (c *Client) GetMetricNames(ctx context.Context) ([]string, error) {
	if c == nil || c.api == nil {
		return nil, fmt.Errorf("prometheus client not initialized")
	}
	labelValues, _, err := c.api.LabelValues(ctx, "__name__", []string{}, time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get metric names: %w", err)
	}
	names := make([]string, 0, len(labelValues))
	for _, v := range labelValues {
		names = append(names, string(v))
	}
	return names, nil
}

func (c *Client) GetAlerts(ctx context.Context) (map[string]interface{}, error) {
	if c == nil || c.api == nil {
		return nil, fmt.Errorf("prometheus client not initialized")
	}
	result, err := c.api.Alerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	alerts := []map[string]interface{}{}
	for _, alert := range result.Alerts {
		alerts = append(alerts, map[string]interface{}{
			"status":      alert.State,
			"labels":      labelSetToMap(alert.Labels),
			"annotations": labelSetToMap(alert.Annotations),
			"startsAt":    alert.ActiveAt,
		})
	}
	return map[string]interface{}{
		"alerts": alerts,
	}, nil
}

func labelSetToMap(ls model.LabelSet) map[string]string {
	out := make(map[string]string)
	for k, v := range ls {
		out[string(k)] = string(v)
	}
	return out
}
