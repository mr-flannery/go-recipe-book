#!/bin/bash
#
# Extract transcript from a YouTube video using yt-dlp
#
# Usage: ./extract-transcript.sh "https://www.youtube.com/watch?v=VIDEO_ID"
#
# Prerequisites: pip install yt-dlp
#
# Output: Plain text transcript saved to {VIDEO_ID}.txt

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <youtube-url>" >&2
    echo "Example: $0 \"https://www.youtube.com/watch?v=dQw4w9WgXcQ\"" >&2
    exit 1
fi

URL="$1"

# Check if yt-dlp is installed
if ! command -v yt-dlp &> /dev/null; then
    echo "Error: yt-dlp is not installed" >&2
    echo "Install with: pip install yt-dlp" >&2
    exit 1
fi

# Get video ID first
VIDEO_ID=$(yt-dlp --get-id "$URL" 2>/dev/null)
if [ -z "$VIDEO_ID" ]; then
    echo "Error: Could not extract video ID from URL" >&2
    exit 1
fi

# Download subtitles to current directory (prefer manual subs, fall back to auto-generated)
yt-dlp \
    --skip-download \
    --write-subs \
    --write-auto-subs \
    --sub-langs "en,de,en.*,de.*" \
    -o "%(id)s.%(ext)s" \
    "$URL" 2>/dev/null || true

# Find the subtitle file (prefer non-auto-generated, prefer .srt over .vtt)
SUB_FILE=""
for pattern in "./${VIDEO_ID}".en.srt "./${VIDEO_ID}".de.srt "./${VIDEO_ID}"*.srt \
               "./${VIDEO_ID}".en.vtt "./${VIDEO_ID}".de.vtt "./${VIDEO_ID}"*.vtt; do
    for f in $pattern; do
        if [ -f "$f" ] 2>/dev/null; then
            # Prefer non-auto-generated (no .auto. in filename)
            if [[ ! "$f" =~ \.auto\. ]] || [ -z "$SUB_FILE" ]; then
                SUB_FILE="$f"
                # If we found a non-auto file, stop looking
                if [[ ! "$f" =~ \.auto\. ]]; then
                    break 2
                fi
            fi
        fi
    done
done

if [ -z "$SUB_FILE" ] || [ ! -f "$SUB_FILE" ]; then
    echo "Error: No subtitles found for this video" >&2
    echo "The video may not have subtitles available" >&2
    exit 1
fi

echo "Using subtitle file: $SUB_FILE" >&2

# Convert VTT/SRT to plain text (remove timestamps and formatting)
OUTPUT_FILE="${VIDEO_ID}.txt"

awk '
    /^WEBVTT/ { next }                            # Skip VTT header
    /^Kind:/ { next }                             # Skip VTT metadata
    /^Language:/ { next }                         # Skip VTT metadata
    /^NOTE/ { next }                              # Skip VTT notes
    /^[0-9]+$/ { next }                           # Skip SRT sequence numbers
    /^[0-9]{2}:[0-9]{2}/ { next }                 # Skip timestamps (both VTT and SRT)
    /-->/ { next }                                # Skip timestamp lines
    /^$/ { next }                                 # Skip empty lines
    {
        gsub(/<[^>]*>/, "")                       # Remove HTML/VTT tags
        gsub(/&nbsp;/, " ")                       # Replace HTML entities
        gsub(/&amp;/, "\\&")
        gsub(/&lt;/, "<")
        gsub(/&gt;/, ">")
        gsub(/\[.*\]/, "")                        # Remove [Music] etc
        gsub(/^[ \t]+|[ \t]+$/, "")               # Trim whitespace
        if (length($0) > 0) print                 # Print non-empty lines
    }
' "$SUB_FILE" | \
awk 'NR==1 || $0!=prev { print; prev=$0 }' > "$OUTPUT_FILE"

# Clean up subtitle files
rm -f "./${VIDEO_ID}"*.vtt "./${VIDEO_ID}"*.srt

echo "Transcript saved to: $OUTPUT_FILE"
echo "Source: $URL"
