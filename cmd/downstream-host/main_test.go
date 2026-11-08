package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func createFramedMessage(t *testing.T, payload []byte) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	length := uint32(len(payload))
	if err := binary.Write(buf, binary.LittleEndian, length); err != nil {
		t.Fatalf("failed to write length prefix: %v", err)
	}
	buf.Write(payload)
	return buf
}

func TestReadMessage_Valid(t *testing.T) {
	inputMsg := NativeMessage{
		Action:   "download",
		URL:      "https://example.com/file.zip",
		Filename: "file.zip",
	}
	data, err := json.Marshal(inputMsg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	buf := createFramedMessage(t, data)
	msg, err := readMessage(buf)
	if err != nil {
		t.Fatalf("unexpected error reading message: %v", err)
	}

	if msg.Action != inputMsg.Action || msg.URL != inputMsg.URL || msg.Filename != inputMsg.Filename {
		t.Errorf("expected %+v, got %+v", inputMsg, msg)
	}
}

func TestReadMessage_ZeroLength(t *testing.T) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))

	_, err := readMessage(buf)
	if err == nil {
		t.Fatal("expected error for 0 length message, got nil")
	}
}

func TestReadMessage_Oversized(t *testing.T) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1024*1024+1))

	_, err := readMessage(buf)
	if err == nil {
		t.Fatal("expected error for message exceeding 1MB, got nil")
	}
}

func TestReadMessage_TruncatedPayload(t *testing.T) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint32(100))
	buf.Write([]byte(`{"action":`)) // Only 10 bytes

	_, err := readMessage(buf)
	if err == nil {
		t.Fatal("expected error for truncated payload, got nil")
	}
}

func TestReadMessage_InvalidJSON(t *testing.T) {
	invalidJSON := []byte("not a json payload")
	buf := createFramedMessage(t, invalidJSON)

	_, err := readMessage(buf)
	if err == nil {
		t.Fatal("expected error for malformed json, got nil")
	}
}

func TestWriteResponse(t *testing.T) {
	resp := NativeResponse{
		Status:  "ok",
		ID:      "test-id-123",
		Message: "download started",
	}

	buf := new(bytes.Buffer)
	if err := writeResponse(buf, resp); err != nil {
		t.Fatalf("writeResponse failed: %v", err)
	}

	var length uint32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		t.Fatalf("failed to read length prefix from output: %v", err)
	}

	payload := buf.Bytes()
	if uint32(len(payload)) != length {
		t.Fatalf("expected payload length %d, got %d", length, len(payload))
	}

	var decoded NativeResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("failed to decode written JSON: %v", err)
	}

	if decoded.Status != resp.Status || decoded.ID != resp.ID || decoded.Message != resp.Message {
		t.Errorf("expected %+v, got %+v", resp, decoded)
	}
}

func TestGetIPCSocketPath(t *testing.T) {
	path := getIPCSocketPath()
	expectedSuffix := fmt.Sprintf("downstream-%d.sock", os.Getuid())
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("expected socket path to end with %s, got %s", expectedSuffix, path)
	}
}
