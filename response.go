package main

import (
	"fmt"
	"time"
)

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("15:04:05"))
	return []byte(stamp), nil
}

// Field not used by navitia aren't serialized

type Response struct {
	CorrelationID     string            `json:"-"`
	MessageResponse   []MessageResponse `json:"-"`
	StopTimesResponse []StopTimesResponse
}

type MessageResponse struct {
	ResponseCode    int
	ResponseComment string
}

type StopTimesResponse struct {
	StopID               string `json:"-"`
	StopTimeoCode        int    `json:"-"`
	StopSAECode          string `json:"-"`
	StopLongNmae         string `json:"-"`
	StopShortName        string `json:"-"`
	StopVocalName        string `json:"-"`
	ReferenceTime        Time   `json:"-"`
	NextStopTimesMessage NextStopTimesMessage
}

type NextStopTimesMessage struct {
	LineID               string `json:"-"`
	LineTimeoCode        int    `json:"-"`
	LineSAECode          string `json:"-"`
	Way                  string `json:"-"`
	LineMainDirection    string `json:"-"`
	NextExpectedStopTime []ExpectedStopTime
}

type ExpectedStopTime struct {
	Destination       string
	Terminus          string
	TerminusID        string `json:"-"`
	TerminusTimeoCode int    `json:"-"`
	TerminusSAECode   string `json:"-"`
	NextStop          Time
}
