package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/rtlong/web-spider/spider"
)

type Logger interface {
	PrintResult(*spider.Result)
	Fatal(string)
	SetOutput(io.Writer)
}

type PlaintextLogger struct {
	Log *log.Logger
}

func (l *PlaintextLogger) PrintResult(r *spider.Result) {
	if r.Error.Error != nil {
		l.Log.Printf("| ERR %s: %s\n", r.Job, r.Error.String())
	} else {
		ms := int64(r.Duration / time.Millisecond)
		l.Log.Printf("| %3d %s [%dms]\n", r.Response.StatusCode, r.Job.String(), ms)
	}
}

func (l *PlaintextLogger) Fatal(message string) {
	l.Log.Fatalf("%s\n", message)
}

func (l *PlaintextLogger) SetOutput(w io.Writer) {
	l.Log = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)
}

type JSONEvent struct {
	Time           time.Time
	Method         string
	URL            string
	ResponseStatus int
	Duration       time.Duration
	Result         *spider.Result
}

type JSONLogger struct {
	Encoder *json.Encoder
}

func (l JSONLogger) PrintResult(r *spider.Result) {
	err := l.Encoder.Encode(JSONEvent{
		Method:         r.Job.Method,
		URL:            r.Job.URL.String(),
		ResponseStatus: r.Response.StatusCode,
		Duration:       r.Duration,
		Time:           r.Time,
		Result:         r,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (l JSONLogger) Fatal(message string) {
	err := l.Encoder.Encode(map[string]string{"Error": message})
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(1)
}

func (l *JSONLogger) SetOutput(w io.Writer) {
	l.Encoder = json.NewEncoder(w)
}
