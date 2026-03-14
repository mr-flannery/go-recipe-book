package extraction

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type AudioDownloadResult struct {
	FilePath string
	Cleanup  func() error
}

func DownloadYouTubeAudio(videoURL string) (*AudioDownloadResult, error) {
	videoID, err := ExtractVideoID(videoURL)
	if err != nil {
		return nil, err
	}

	tempDir, err := os.MkdirTemp("", "yt-audio-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	outputPath := filepath.Join(tempDir, videoID+".mp3")

	cmd := exec.Command("yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "128K",
		"-o", outputPath,
		"--no-playlist",
		"--no-warnings",
		"--quiet",
		videoURL,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("yt-dlp failed: %w, output: %s", err, string(output))
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("audio file was not created")
	}

	return &AudioDownloadResult{
		FilePath: outputPath,
		Cleanup: func() error {
			return os.RemoveAll(tempDir)
		},
	}, nil
}
