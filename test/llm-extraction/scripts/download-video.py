#!/usr/bin/env python3
"""
Download YouTube audio for recipe extraction testing.

Usage:
    python download-video.py <youtube_url> [output_name]
    
Examples:
    python download-video.py "https://www.youtube.com/watch?v=VIDEO_ID"
    python download-video.py "https://www.youtube.com/watch?v=VIDEO_ID" video-03-cookies-en

Requirements:
    pip install yt-dlp

The script downloads audio-only in MP3 format suitable for Gemini's audio API.
Audio files are typically 2-5MB vs 50-100MB for video, avoiding payload size limits.
"""

import sys
import os
import subprocess
import re

def get_video_id(url: str) -> str | None:
    """Extract video ID from various YouTube URL formats."""
    patterns = [
        r'(?:v=|/v/|youtu\.be/)([a-zA-Z0-9_-]{11})',
        r'(?:embed/)([a-zA-Z0-9_-]{11})',
    ]
    for pattern in patterns:
        match = re.search(pattern, url)
        if match:
            return match.group(1)
    return None

def download_audio(url: str, output_name: str | None = None) -> str:
    """Download YouTube audio using yt-dlp."""
    
    video_id = get_video_id(url)
    if not video_id:
        print(f"Error: Could not extract video ID from URL: {url}", file=sys.stderr)
        sys.exit(1)
    
    # Default output name based on video ID
    if not output_name:
        output_name = f"video-{video_id}"
    
    # Ensure we're in the right directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    samples_dir = os.path.join(script_dir, "..", "samples")
    os.makedirs(samples_dir, exist_ok=True)
    
    output_path = os.path.join(samples_dir, f"{output_name}.mp3")
    
    # yt-dlp command for audio-only download
    # - Extract best audio
    # - Convert to MP3 (supported by Gemini)
    # - Much smaller file size than video
    cmd = [
        "yt-dlp",
        "-x",  # Extract audio
        "--audio-format", "mp3",
        "--audio-quality", "128K",  # 128kbps is plenty for speech
        "-o", output_path,
        "--no-playlist",
        "--progress",
        url
    ]
    
    print(f"Downloading audio: {url}")
    print(f"Output: {output_path}")
    print()
    
    try:
        subprocess.run(cmd, check=True)
        print()
        print(f"Success! Audio saved to: {output_path}")
        
        # Show file size
        if os.path.exists(output_path):
            size_mb = os.path.getsize(output_path) / (1024 * 1024)
            print(f"File size: {size_mb:.1f} MB")
        
        return output_path
        
    except subprocess.CalledProcessError as e:
        print(f"Error downloading audio: {e}", file=sys.stderr)
        sys.exit(1)
    except FileNotFoundError:
        print("Error: yt-dlp not found. Install it with: pip install yt-dlp", file=sys.stderr)
        sys.exit(1)

def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(1)
    
    url = sys.argv[1]
    output_name = sys.argv[2] if len(sys.argv) > 2 else None
    
    download_audio(url, output_name)

if __name__ == "__main__":
    main()
