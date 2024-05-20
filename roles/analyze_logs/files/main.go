package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	logFilePath = "/tmp/data_transformer.log"
	layout      = "2006-01-02 15:04:05.999"
	statusOK    = "status 200"
)

type Log struct {
	Time    time.Time
	Status  string
	Message string
}

type Disconnect struct {
	StartTime time.Time
	EndTime   time.Time
	Type      string
}

type Disconnections []Disconnect

func main() {
	disconnections, logs, err := findDisconnects(logFilePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(disconnections) == 0 {
		fmt.Println("No missing status 200s in the log.")
	} else {
		fmt.Println("Detected Events:")
		for _, d := range disconnections {
			d.ClassifyEvent(logs)
			duration := d.EndTime.Sub(d.StartTime)
			fmt.Printf("Start time: %v, Duration: %v, Classification: %s\n", d.StartTime, duration, d.Type)
		}
	}
}

func findDisconnects(filePath string) (Disconnections, []Log, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening log file: %w", err)
	}
	defer file.Close()

	var (
		prevTime       time.Time
		disconnections Disconnections
		logs           []Log
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, " ", 4)
		if len(fields) < 4 {
			continue
		}

		logTime, err := time.Parse(layout, fields[0]+" "+fields[1][:len(fields[1])-1])
		if err != nil {
			continue
		}

		log := Log{
			Time:    logTime,
			Status:  fields[2],
			Message: fields[3],
		}

		logs = append(logs, log)

		if strings.Contains(log.Message, statusOK) {
			currTime := log.Time

			if !prevTime.IsZero() {
				duration := currTime.Sub(prevTime)
				if duration > (2*time.Second + 900*time.Millisecond) {
					disconnections = append(disconnections, Disconnect{StartTime: prevTime, EndTime: currTime})
				}
			}

			prevTime = currTime
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading log file: %w", err)
	}

	return disconnections, logs, nil
}

func (d *Disconnect) ClassifyEvent(logs []Log) {
	const delayThreshold = 3
	delayBufferTime := d.EndTime.Add(2 * time.Second)

	var count int
	for _, log := range logs {
		if log.Time.After(d.EndTime) && log.Time.Before(delayBufferTime) && strings.Contains(log.Message, statusOK) {
			count++
		}
	}

	if count > delayThreshold {
		d.Type = "Delay"
	} else {
		d.Type = "Disconnect"
	}
}
