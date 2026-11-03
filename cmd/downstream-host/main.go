package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

type NativeMessage struct {
	Action   string `json:"action"`
	URL      string `json:"url"`
	Filename string `json:"filename,omitempty"`
	Referrer string `json:"referrer,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
}

type NativeResponse struct {
	Status   string      `json:"status"`
	ID       string      `json:"id,omitempty"`
	Message  string      `json:"message,omitempty"`
	Settings interface{} `json:"settings,omitempty"`
}

func getIPCSocketPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("downstream-%d.sock", os.Getuid()))
}

// readMessage reads a 4-byte Little-Endian length prefix followed by JSON from stdin.
func readMessage(r io.Reader) (*NativeMessage, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	if length == 0 || length > 1024*1024 { // 1MB sanity check
		return nil, fmt.Errorf("invalid message length: %d", length)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("failed to read full payload: %w", err)
	}

	var msg NativeMessage
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	return &msg, nil
}

// writeResponse writes a 4-byte Little-Endian length prefix followed by JSON to stdout.
func writeResponse(w io.Writer, resp NativeResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	if err := binary.Write(w, binary.LittleEndian, length); err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func main() {
	msg, err := readMessage(os.Stdin)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "downstream-host -- read error : %v\n", err)
		_ = writeResponse(os.Stdout, NativeResponse{
			Status:  "error",
			Message: fmt.Sprintf("failed to read message: %v", err),
		})
		return
	}

	socketPath := getIPCSocketPath()
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "downstream-host -- socket dial timeout: %v\n", err)
		_ = writeResponse(os.Stdout, NativeResponse{
			Status:  "error",
			Message: "Downstream desktop is not running",
		})
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "downstream-host -- socket write error: %v\n", err)
		_ = writeResponse(os.Stdout, NativeResponse{
			Status:  "error",
			Message: fmt.Sprintf("failed to send response to Downstream: %v", err),
		})
		return
	}

	var appResp NativeResponse
	if err := json.NewDecoder(conn).Decode(&appResp); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "downstream-host -- socket decode error: %v\n", err)
		_ = writeResponse(os.Stdout, NativeResponse{
			Status:  "error",
			Message: fmt.Sprintf("failed to read response from Downstream: %v", err),
		})
		return
	}

	if err := writeResponse(os.Stdout, appResp); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "downstream-host -- stdout write error: %v\n", err)
	}
}
