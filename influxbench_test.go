package main

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

const (
	MyDB     = "testDB"
	username = "root"
)

const sampleSize = 10000000

func BenchmarkWriteToFile(b *testing.B) {
	file, err := os.Create(`log.txt`)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	rand.Seed(time.Now().UnixNano())
	logstr := fmt.Sprintf("time: %v, cpu: cpu-total, host: host123456, region:  tokyo", time.Now())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file.WriteString(logstr)
		file.WriteString(fmt.Sprintf("idle: 50.0, busy: 50.0"))
	}
}

func BenchmarkWrite(b *testing.B) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
	})
	if err != nil {
		b.Fatal(err)
	}
	c.Query(client.Query{Command: `CREATE DATABASE testDB`})
	rand.Seed(time.Now().UnixNano())
	tags := map[string]string{
		"cpu":    "cpu-total",
		"host":   "host123456",
		"region": "tokyo",
	}

	fields := map[string]interface{}{
		"idle": 50.0,
		"busy": 50.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  "testDB",
			Precision: "us",
		})

		pt, _ := client.NewPoint(
			"cpu_usage",
			tags,
			fields,
			time.Now(),
		)
		bp.AddPoint(pt)
		c.Write(bp)
	}
}

func BenchmarkQueryData(b *testing.B) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := client.Query{
			Command:  `SELECT * FROM cpu_usage LIMIT 10`,
			Database: MyDB,
		}
		if response, err := c.Query(q); err == nil {
			if response.Error() != nil {
				b.Fatal(err)
			}
			_ = response.Results
		} else {
			b.Fatal(err)
		}
	}
}
