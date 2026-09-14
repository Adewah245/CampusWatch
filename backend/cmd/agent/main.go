package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"CampusWatch/backend/internal/model"
)

type config struct {
	ServerURL  string
	AgentCode  string
	Credential string
	AgentID    string
	SystemID   string
	Version    string
	Interval   time.Duration
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := &http.Client{Timeout: 10 * time.Second}
	if err := sendHeartbeat(ctx, client, cfg); err != nil {
		log.Printf("initial heartbeat failed: %v", err)
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sendHeartbeat(ctx, client, cfg); err != nil {
				log.Printf("heartbeat failed: %v", err)
			}
		}
	}
}

func loadConfig() (config, error) {
	interval := 30 * time.Second
	if value := os.Getenv("CAMPUSWATCH_HEARTBEAT_INTERVAL_SECONDS"); value != "" {
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds < 5 {
			return config{}, errors.New("CAMPUSWATCH_HEARTBEAT_INTERVAL_SECONDS must be at least 5")
		}
		interval = time.Duration(seconds) * time.Second
	}
	cfg := config{
		ServerURL:  strings.TrimRight(os.Getenv("CAMPUSWATCH_SERVER_URL"), "/"),
		AgentCode:  os.Getenv("CAMPUSWATCH_AGENT_CODE"),
		Credential: os.Getenv("CAMPUSWATCH_AGENT_CREDENTIAL"),
		AgentID:    os.Getenv("CAMPUSWATCH_AGENT_ID"),
		SystemID:   os.Getenv("CAMPUSWATCH_SYSTEM_ID"),
		Version:    os.Getenv("CAMPUSWATCH_AGENT_VERSION"),
		Interval:   interval,
	}
	if cfg.Version == "" {
		cfg.Version = "0.1.0"
	}
	if cfg.ServerURL == "" || cfg.AgentCode == "" || cfg.Credential == "" || cfg.AgentID == "" || cfg.SystemID == "" {
		return config{}, errors.New("CAMPUSWATCH_SERVER_URL, CAMPUSWATCH_AGENT_CODE, CAMPUSWATCH_AGENT_CREDENTIAL, CAMPUSWATCH_AGENT_ID, and CAMPUSWATCH_SYSTEM_ID are required")
	}
	return cfg, nil
}

func sendHeartbeat(ctx context.Context, client *http.Client, cfg config) error {
	now := time.Now().UTC()
	report := model.HeartbeatRequest{
		AgentID: cfg.AgentID, SystemID: cfg.SystemID, Timestamp: now, AgentVersion: cfg.Version,
		Health: &model.HealthReport{
			RecordedAt: now, CPUPercent: 0, MemoryPercent: 0, DiskPercent: 0,
			NetworkConnected: true, OperatingSystem: runtime.GOOS, AgentHealth: "healthy",
			UptimeSeconds: 0,
		},
	}
	body, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("encode heartbeat: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.ServerURL+"/api/v1/agents/heartbeat", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create heartbeat request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Agent-Code", cfg.AgentCode)
	request.Header.Set("X-Agent-Credential", cfg.Credential)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("send heartbeat: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("heartbeat returned HTTP %s", response.Status)
	}
	return nil
}
