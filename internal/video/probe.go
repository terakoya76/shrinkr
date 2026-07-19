package video

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/terakoya76/shrinkr/internal/runner"
)

type Info struct {
	Width       int
	Height      int
	VideoCodec  string
	AudioCodec  string
	AudioBitsPS int
}

// Probe reports basic stream info via ffprobe. Returns a zero-value Info if
// ffprobe is not available (caller falls back to default encoding params).
func Probe(ctx context.Context, r runner.Runner, path string) (Info, error) {
	if _, err := r.LookPath("ffprobe"); err != nil {
		return Info{}, nil
	}
	out, err := r.Run(ctx, "ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		path,
	)
	if err != nil {
		return Info{}, fmt.Errorf("ffprobe: %w: %s", err, string(out))
	}
	var raw struct {
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			BitRateStr string `json:"bit_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return Info{}, fmt.Errorf("parse ffprobe json: %w", err)
	}
	var info Info
	for _, s := range raw.Streams {
		switch s.CodecType {
		case "video":
			info.Width = s.Width
			info.Height = s.Height
			info.VideoCodec = s.CodecName
		case "audio":
			info.AudioCodec = s.CodecName
			if s.BitRateStr != "" {
				if v, err := strconv.Atoi(s.BitRateStr); err == nil {
					info.AudioBitsPS = v
				}
			}
		}
	}
	return info, nil
}
