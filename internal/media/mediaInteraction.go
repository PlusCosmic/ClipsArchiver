package media

import (
	"github.com/u2takey/ffmpeg-go"
	"github.com/vansante/go-ffprobe"
	"time"
)

const encoder = "h264_videotoolbox"

func TranscodeVideoFile(input string, output string) error {
	err := ffmpeg_go.Input(input).Output(output, ffmpeg_go.KwArgs{"c:v": encoder, "q:v": 65, "vf": "scale=1920:1080"}).OverWriteOutput().ErrorToStdOut().Run()
	return err
}

func GenerateThumbnailFromVideo(input string, output string) error {
	err := ffmpeg_go.Input(input).Output(output, ffmpeg_go.KwArgs{"ss": "00:00:01.000", "frames:v": 1}).OverWriteOutput().ErrorToStdOut().Run()
	return err
}

func TrimVideoFile(input string, output string, startTime int, endTime int) {

}

func GetVideoProbeData(path string) (*ffprobe.ProbeData, error) {
	result, err := ffprobe.GetProbeData(path, 120000*time.Millisecond)
	return result, err
}
