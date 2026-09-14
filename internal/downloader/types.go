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
}

type DownloadState struct {
	ID          string
	URL         string
	Filename    string
	Category    string
	TotalSize   int64
	Status      DownloadStatus
	ScheduledAt *time.Time
}

type ProgressWriter struct {
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
	Filename    string
	Percentage  float64
	CurrentSize int64
	TotalSize   int64
	Speed       float64
	ETA         time.Duration
}

func (s DownloadStatus) String() string {
	names := [...]string{"Queued", "Downloading", "Paused", "Completed", "Error", "Scheduled"}
	if int(s) < 0 || int(s) >= len(names) {
		return fmt.Sprintf("Undefined(%d)", s)
	}
	return names[s]
}

type PartState struct {
	ID          string
	DownloadID  string
	StartByte   int64
	EndByte     int64
	CurrentByte int64
	WorkerID    string
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
}

type Category struct {
	Name       string
	FolderPath string
	Extension  []string
}
