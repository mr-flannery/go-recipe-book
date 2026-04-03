package extraction

import (
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
		return nil, technicalErrorf("failed to create temp directory: %w", err)
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
		return nil, technicalErrorf("yt-dlp failed: %w, output: %s", err, string(output))
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.RemoveAll(tempDir)
		return nil, technicalErrorf("audio file was not created")
	}

	return &AudioDownloadResult{
		FilePath: outputPath,
		Cleanup: func() error {
			return os.RemoveAll(tempDir)
		},
	}, nil
}
