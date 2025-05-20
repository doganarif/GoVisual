package model

import (
	"net/http"
	"time"
)

type RequestLog struct {
	ID              string                   `json:"ID" bson:"_id"`
	Timestamp       time.Time                `json:"Timestamp"`
	Method          string                   `json:"Method"`
	Path            string                   `json:"Path"`
	Query           string                   `json:"Query"`
	RequestHeaders  http.Header              `json:"RequestHeaders"`
	ResponseHeaders http.Header              `json:"ResponseHeaders"`
	StatusCode      int                      `json:"StatusCode"`
	Duration        int64                    `json:"Duration"`
	RequestBody     string                   `json:"RequestBody,omitempty"`
	ResponseBody    string                   `json:"ResponseBody,omitempty"`
	Error           string                   `json:"Error,omitempty"`
	MiddlewareTrace []map[string]interface{} `json:"MiddlewareTrace,omitempty"`
	RouteTrace      map[string]interface{}   `json:"RouteTrace,omitempty"`
}

func NewRequestLog(req *http.Request) *RequestLog {
	return &RequestLog{
		ID:             generateID(),
		Timestamp:      time.Now(),
		Method:         req.Method,
		Path:           req.URL.Path,
		Query:          req.URL.RawQuery,
		RequestHeaders: req.Header,
	}
}

func generateID() string {
	return time.Now().Format("20060102-150405.000000")
}
