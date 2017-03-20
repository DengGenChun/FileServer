// Used for auto create the file directory, log directory and log file
// corresponding to the date of today.

package main

import (
	"io"
	"log"
	"os"
	"time"
)

var (
	logFile    *os.File
	FileSubDir = ""
	LogSubDir  = ""
)

func init() {
	go autoCreateDirectory()
}

// Update and create file directory, log directory and log file for every day.
func autoCreateDirectory() {
	// Update file directory path
	FileSubDir = time.Now().Format("2006/01/02")
	// Update and create a new log directory and log file
	createLogFile()

	timeOfEarlyMorningToday := time.Now().Truncate(24 * time.Hour)
	// Set a timer, triggered when a new day coming
	timer := time.NewTimer(timeOfEarlyMorningToday.Add(24 * time.Hour).Sub(time.Now()))
	for {
		select {
		case <-timer.C:
			// Update file directory path
			FileSubDir = time.Now().Format("2006/01/02")
			// Update and create a new log directory and log file
			createLogFile()
			// Reset the timer, triggered when the next new day coming
			timer.Reset(24 * time.Hour)
		}
	}
}

func createLogFile() {
	LogSubDir = time.Now().Format("2006/01")

	var err error
	// Create log directory
	if err = os.MkdirAll(LogMainDir+"/"+LogSubDir, 0666); err != nil {
		panic(err.Error())
	}
	// Close last log file
	if logFile != nil {
		logFile.Close()
	}
	// Open new log file
	fileName := LogMainDir + "/" + LogSubDir + "/" + time.Now().Format("2006-01-02") + ".log"
	logFile, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}
	// Bind log output to new log file and StdOut
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
}
