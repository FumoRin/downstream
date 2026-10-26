package downloader

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type DownloadStatus int

const (
	StateQueue DownloadStatus = iota
	StateDownloading
	StatePaused
	StateCompleted
	StateError
	StateScheduled
)

type DownloadOptions struct {
	URL      string
	Filename string
	Dir      string
	Size     int64
	Progress chan Progress
	Repo     DownloadRepository
	Settings *Settings
	Limiter  *rate.Limiter
}

type DownloadJob struct {
	ID       string
	URL      string
	Dir      string
	Filename string
	Limiter  *rate.Limiter
}

type DownloadManager struct {
	wg              sync.WaitGroup
	progress        chan Progress
	job             chan DownloadJob
	repo            DownloadRepository
	cancellationsMu sync.Mutex
	cancellations   map[string]context.CancelFunc
	ctx             context.Context
	cancel          context.CancelFunc
	settings        *Settings
	workerMu        sync.Mutex
	workerCount     int
	retireWorker    chan struct{}
	limiter         *rate.Limiter
}

type DownloadState struct {
	ID          string         `json:"id"`
	URL         string         `json:"url"`
	Filename    string         `json:"filename"`
	Category    string         `json:"category"`
	TotalSize   int64          `json:"totalSize"`
	CurrentSize int64          `json:"currentSize"`
	Percentage  float64        `json:"percentage"`
	Status      DownloadStatus `json:"status"`
	ScheduledAt *time.Time     `json:"scheduledAt"`
	Parts       []*PartState   `json:"parts,omitempty"`
}

type ProgressWriter struct {
	ID           string
	Filename     string
	Total        int64
	Current      int64
	ByteAtStart  int64
	Destination  io.Writer
	ProgressChan chan Progress

	StartTime  time.Time
	LastUpdate time.Time
}

type Progress struct {
	ID          string        `json:"id"`
	Filename    string        `json:"filename"`
	Percentage  float64       `json:"percentage"`
	CurrentSize int64         `json:"currentSize"`
	TotalSize   int64         `json:"totalSize"`
	Speed       float64       `json:"speed"`
	ETA         time.Duration `json:"eta"`
	Parts       []*PartState  `json:"parts,omitempty"`
}

func (s DownloadStatus) String() string {
	names := [...]string{"Queued", "Downloading", "Paused", "Completed", "Error", "Scheduled"}
	if int(s) < 0 || int(s) >= len(names) {
		return fmt.Sprintf("Undefined(%d)", s)
	}
	return names[s]
}

type PartState struct {
	ID          string `json:"id"`
	DownloadID  string `json:"downloadID"`
	StartByte   int64  `json:"startByte"`
	EndByte     int64  `json:"endByte"`
	CurrentByte int64  `json:"currentByte"`
	WorkerID    string `json:"workerID"`
}

type DownloadRepository interface {
	SaveDownload(state *DownloadState) error
	GetDownload(id string) (*DownloadState, error)
	GetIncompleteDownload() ([]*DownloadState, error)
	GetAllDownloads() ([]*DownloadState, error)
	DeleteDownload(id string) error
	UpdatePartsProgress(partID string, currentByte int64) error
	GetParts(downloadID string) ([]*PartState, error)
	CreatePart(part *PartState) error
	DeletePart(downloadID string) error
	UpdateFilename(id string, newFilename string) error
	GetDueScheduledDownloads(now time.Time) ([]*DownloadState, error)
	GetCategories() ([]*Category, error)
	GetAllSettings() (*Settings, error)
	SetSettings(key, value string) error
}

type TargetInfo struct {
	Filename         string
	TotalSize        int64
	SupportMultiPart bool
}

type Settings struct {
	DownloadDir            string
	MaxConcurrencyDownload int
	PartsPerDownload       int
	MinMultiPartDownload   int64
	SpeedLimit             int64
}

type Category struct {
	Name       string
	FolderPath string
	Extension  []string
}

type DiskStorage struct {
	Free  uint64 `json:"free"`
	Total uint64 `json:"total"`
}
