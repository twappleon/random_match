package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type snapshotMeta struct {
	RoomID    string    `json:"roomId"`
	UserID    string    `json:"userId"`
	PeerID    string    `json:"peerId"`
	Mode      string    `json:"mode"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"createdAt"`
	ImagePath string    `json:"imagePath"`
}

func saveSnapshot(root, userID string, req snapshotRequest) (string, error) {
	mime, payload, err := splitDataURL(req.Image)
	if err != nil {
		return "", err
	}
	if mime != "image/jpeg" && mime != "image/png" {
		return "", errors.New("unsupported image type")
	}
	image, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}
	if len(image) == 0 || len(image) > 2*1024*1024 {
		return "", errors.New("invalid image size")
	}

	now := time.Now().UTC()
	dir := filepath.Join(root, now.Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	ext := ".jpg"
	if mime == "image/png" {
		ext = ".png"
	}
	base := safeFilePart(req.RoomID) + "-" + safeFilePart(userID)
	imagePath := filepath.Join(dir, base+ext)
	if err := os.WriteFile(imagePath, image, 0o644); err != nil {
		return "", err
	}

	meta := snapshotMeta{
		RoomID:    req.RoomID,
		UserID:    userID,
		PeerID:    req.PeerID,
		Mode:      req.Mode,
		Width:     req.Width,
		Height:    req.Height,
		CreatedAt: now,
		ImagePath: imagePath,
	}
	metaPayload, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, base+".json"), metaPayload, 0o644); err != nil {
		return "", err
	}
	return imagePath, nil
}

func splitDataURL(value string) (string, string, error) {
	const marker = ";base64,"
	if !strings.HasPrefix(value, "data:") {
		return "", "", errors.New("image must be data url")
	}
	parts := strings.SplitN(strings.TrimPrefix(value, "data:"), marker, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("invalid data url")
	}
	return parts[0], parts[1], nil
}

func safeFilePart(value string) string {
	var builder strings.Builder
	for _, item := range value {
		if item >= 'a' && item <= 'z' || item >= 'A' && item <= 'Z' || item >= '0' && item <= '9' || item == '-' || item == '_' {
			builder.WriteRune(item)
		}
	}
	if builder.Len() == 0 {
		return "unknown"
	}
	return builder.String()
}
