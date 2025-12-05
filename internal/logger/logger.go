/*
Copyright (c) 2025 selene466 <selene.banderas.466@gmail.com>
*/

// Package logger.
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/selene466/go-fingerprint-backend/internal/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogFile struct {
	Name       string     `json:"Name"`
	ServerHost string     `json:"Server Host"`
	ServerPort int        `json:"Server Port"`
	Log        []LogEntry `json:"log"`
}

type LogEntry struct {
	Time      time.Time `json:"time"`
	Level     string    `json:"level"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
}

var (
	lumberjackLogger *lumberjack.Logger
	logData          LogFile
	currentDate      string
)

func Init() error {
	cfg := config.Get()
	dir := "/tmp/" + cfg.App.Name + "/"
	dir = strings.TrimRight(dir, "/") + "/"

	logData = LogFile{
		Name:       "Fingerprint Backend",
		ServerHost: cfg.Server.Host,
		ServerPort: cfg.Server.Port,
		Log:        []LogEntry{},
	}

	_ = cleanupOldLogs(dir, 30)

	currentDate = time.Now().Format("2006-01-02")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	filename := fmt.Sprintf("logfile_%s.log.json", currentDate)
	fullPath := filepath.Join(dir, filename)
	lumberjackLogger = &lumberjack.Logger{
		Filename:   fullPath,
		MaxSize:    2, // MB
		MaxBackups: 30,
		MaxAge:     30, // Days
		LocalTime:  true,
	}

	fmt.Printf("Log path: %s\n", fullPath)

	return nil
}

func log(level, component, message string) {
	cfg := config.Get()
	dir := "/tmp/" + cfg.App.Name + "/"
	dir = strings.TrimRight(dir, "/") + "/"
	now := time.Now()
	date := now.Format("2006-01-02")

	if date != currentDate {
		currentDate = date
		filename := fmt.Sprintf("logfile_%s.log.json", currentDate)
		fullPath := filepath.Join(dir, filename)
		fmt.Println(fullPath)
		lumberjackLogger = &lumberjack.Logger{
			Filename:   fullPath,
			MaxSize:    2, // MB
			MaxBackups: 30,
			MaxAge:     30, // Days
			LocalTime:  true,
		}
		// Reset new day
		logData.Log = []LogEntry{}
	}

	entry := LogEntry{
		Time:      now,
		Level:     level,
		Component: component,
		Message:   message,
	}

	logData.Log = append(logData.Log, entry)
	entryData, err := json.Marshal(entry)
	if err != nil {
		panic(err)
	}

	if _, err := lumberjackLogger.Write(entryData); err != nil {
		panic(err)
	}

	if _, err := lumberjackLogger.Write([]byte("\n")); err != nil {
		panic(err)
	}

	fmt.Printf("[%s] [%5s] %-15s: %s\n", now.Format(time.RFC3339), level, component, message)
}

func Debug(component, message string) {
	log("debug", component, message)
}

func Info(component, message string) {
	log("info", component, message)
}

func Warning(component, message string) {
	log("warn", component, message)
}

func Error(component, message string) {
	log("error", component, message)
}

func Fatal(component, message string) {
	log("fatal", component, message)
	os.Exit(1)
}

func cleanupOldLogs(dir string, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		path := filepath.Join(dir, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			_ = os.Remove(path)
		}
	}

	return nil
}
