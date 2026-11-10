package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fumorin/downstream/internal/downloader"
)

func TestGetDiskStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "downstream_disk_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	app := &App{
		settings: downloader.Settings{
			DownloadDir: tempDir,
		},
	}

	storage, err := app.GetDiskStorage()
	if err != nil {
		t.Fatalf("GetDiskStorage failed: %v", err)
	}

	if storage.Total == 0 {
		t.Errorf("expected Total > 0, got %d", storage.Total)
	}
	if storage.Free == 0 {
		t.Errorf("expected Free > 0, got %d", storage.Free)
	}
	if storage.Free > storage.Total {
		t.Errorf("expected Free <= Total, got Free=%d, Total=%d", storage.Free, storage.Total)
	}
}

func sendIPCRequest(t *testing.T, app *App, msg IPCMessage) IPCResponse {
	t.Helper()
	client, server := net.Pipe()
	defer func() {
		_ = client.Close()
	}()

	go app.handleIPCConnection(server)

	if err := json.NewEncoder(client).Encode(msg); err != nil {
		t.Fatalf("failed to encode IPC message: %v", err)
	}

	var resp IPCResponse
	if err := json.NewDecoder(client).Decode(&resp); err != nil {
		t.Fatalf("failed to decode IPC response: %v", err)
	}
	return resp
}

func TestIPC_Ping(t *testing.T) {
	app := &App{}
	resp := sendIPCRequest(t, app, IPCMessage{Action: "ping"})
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
}

func TestIPC_InvalidJSON(t *testing.T) {
	app := &App{}
	client, server := net.Pipe()
	defer func() {
		_ = client.Close()
	}()

	go app.handleIPCConnection(server)

	_, _ = client.Write([]byte("{invalid-json\n"))

	var resp IPCResponse
	if err := json.NewDecoder(client).Decode(&resp); err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestIPC_UnknownAction(t *testing.T) {
	app := &App{}
	resp := sendIPCRequest(t, app, IPCMessage{Action: "nonexistent_action"})
	if resp.Status != "error" {
		t.Errorf("expected status error, got %s", resp.Status)
	}
}

func TestIPC_Settings(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test.db")
	repo, err := downloader.NewSQLiteRepository(tempDB)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	app := &App{repo: repo}

	// Test update settings
	updateResp := sendIPCRequest(t, app, IPCMessage{
		Action: "update_settings",
		Key:    "download_dir",
		Value:  "/tmp/downloads",
	})
	if updateResp.Status != "ok" {
		t.Errorf("update_settings failed: %v", updateResp.Message)
	}

	// Test get settings
	getResp := sendIPCRequest(t, app, IPCMessage{Action: "get_settings"})
	if getResp.Status != "ok" || getResp.Settings == nil {
		t.Fatalf("get_settings failed: %+v", getResp)
	}
	if getResp.Settings.DownloadDir != "/tmp/downloads" {
		t.Errorf("expected /tmp/downloads, got %s", getResp.Settings.DownloadDir)
	}
}

func TestIPC_Download_Validation(t *testing.T) {
	app := &App{}
	resp := sendIPCRequest(t, app, IPCMessage{Action: "download", URL: ""})
	if resp.Status != "error" || !strings.Contains(resp.Message, "cannot be empty") {
		t.Errorf("expected empty URL error, got %+v", resp)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"~", home},
		{"~/Downloads", filepath.Join(home, "Downloads")},
		{"/var/log", "/var/log"},
	}

	for _, tt := range tests {
		got := expandPath(tt.input)
		if got != tt.expected {
			t.Errorf("expandPath(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestApp_UpdateSettings(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "settings_test.db")
	repo, err := downloader.NewSQLiteRepository(tempDB)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	initialSettings := downloader.Settings{
		DownloadDir:            t.TempDir(),
		MaxConcurrencyDownload: 3,
		PartsPerDownload:       4,
		MinMultiPartDownload:   1024 * 1024,
		SpeedLimit:             0,
		BrowserDownloadAction:  "prompt",
		BrowserAutoForeground:  false,
	}
	manager := downloader.NewDownloadManager(repo, &initialSettings)
	app := &App{
		repo:     repo,
		manager:  manager,
		settings: initialSettings,
	}

	newSettings := downloader.Settings{
		DownloadDir:            "/custom/download/path",
		MaxConcurrencyDownload: 6,
		PartsPerDownload:       8,
		MinMultiPartDownload:   5 * 1024 * 1024,
		SpeedLimit:             500 * 1024,
		BrowserDownloadAction:  "auto",
		BrowserAutoForeground:  true,
	}

	if err := app.UpdateSettings(newSettings); err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	// Verify persistence in SQLite repository
	persisted, err := app.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}

	if persisted.DownloadDir != newSettings.DownloadDir ||
		persisted.MaxConcurrencyDownload != newSettings.MaxConcurrencyDownload ||
		persisted.PartsPerDownload != newSettings.PartsPerDownload ||
		persisted.MinMultiPartDownload != newSettings.MinMultiPartDownload ||
		persisted.SpeedLimit != newSettings.SpeedLimit ||
		persisted.BrowserDownloadAction != newSettings.BrowserDownloadAction ||
		persisted.BrowserAutoForeground != newSettings.BrowserAutoForeground {
		t.Errorf("persisted settings mismatch: got %+v, expected %+v", persisted, newSettings)
	}
}

func TestApp_ScheduleDownload_Validation(t *testing.T) {
	app := &App{}

	// Empty URL
	if _, err := app.ScheduleDownload("", "test.zip", "", time.Now().Add(1*time.Hour).Unix()); err == nil {
		t.Error("expected error for empty URL, got nil")
	}

	// Zero timestamp
	if _, err := app.ScheduleDownload("https://example.com/file.zip", "file.zip", "", 0); err == nil {
		t.Error("expected error for zero timestamp, got nil")
	}

	// Past timestamp
	pastTime := time.Now().Add(-1 * time.Hour).Unix()
	if _, err := app.ScheduleDownload("https://example.com/file.zip", "file.zip", "", pastTime); err == nil {
		t.Error("expected error for past timestamp, got nil")
	}
}

func TestApp_ProbeURL_Empty(t *testing.T) {
	app := &App{}
	if _, err := app.ProbeURL(""); err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}

func TestApp_OpenFile_Validation(t *testing.T) {
	app := &App{}

	// Empty path
	if err := app.OpenFile(""); err == nil {
		t.Error("expected error for empty path, got nil")
	}

	// Nonexistent path
	if err := app.OpenFile("/nonexistent/file/path/downstream_test.tmp"); err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestApp_ShowInFolder_Validation(t *testing.T) {
	app := &App{}
	if err := app.ShowInFolder(""); err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestIPC_Download_WithHeadersAndCookies(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "ipc_headers_test.db")
	repo, err := downloader.NewSQLiteRepository(tempDB)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	_ = repo.SetSettings("browser_download_action", "auto")
	settings := downloader.Settings{
		BrowserDownloadAction: "auto",
	}

	manager := downloader.NewDownloadManager(repo, &settings)
	app := &App{
		repo:     repo,
		manager:  manager,
		settings: settings,
	}

	msg := IPCMessage{
		Action:   "download",
		URL:      "https://example.com/file.zip",
		Filename: "file.zip",
		Cookies:  "session=abc123xyz",
		Headers: map[string]string{
			"Authorization": "Bearer test_token",
		},
	}

	resp := sendIPCRequest(t, app, msg)
	if resp.Status != "ok" || resp.ID == "" {
		t.Fatalf("expected status ok with download ID, got %+v", resp)
	}

	// Verify headers and cookies were persisted in SQLite
	persisted, err := repo.GetDownload(resp.ID)
	if err != nil || persisted == nil {
		t.Fatalf("failed to fetch persisted download: %v", err)
	}
	if persisted.Headers["Cookie"] != "session=abc123xyz" {
		t.Errorf("expected Cookie header session=abc123xyz, got %s", persisted.Headers["Cookie"])
	}
	if persisted.Headers["Authorization"] != "Bearer test_token" {
		t.Errorf("expected Authorization header Bearer test_token, got %s", persisted.Headers["Authorization"])
	}
}
