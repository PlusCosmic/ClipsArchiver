package media

import (
	"github.com/u2takey/ffmpeg-go"
	"github.com/vansante/go-ffprobe"
	"math"
	"strconv"
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

func TrimVideoFile(input string, output string, startTimeSeconds int, endTimeSeconds int) error {
	startSeconds := startTimeSeconds % 60
	startMinutes := math.Floor(float64(startTimeSeconds) / float64(60))
	endSeconds := endTimeSeconds % 60
	endMinutes := math.Floor(float64(endTimeSeconds) / float64(60))
	err := ffmpeg_go.Input(input).Output(output, ffmpeg_go.KwArgs{"ss": "00:" + strconv.Itoa(int(startMinutes)) + ":" + strconv.Itoa(startSeconds), "to": "00:" + strconv.Itoa(int(endMinutes)) + ":" + strconv.Itoa(endSeconds), "c": "copy"}).OverWriteOutput().ErrorToStdOut().Run()
	return err
}

func CombineVideoFiles(input1 string, input2 string, output string) error {
	stream1 := ffmpeg_go.Input(input1, nil)
	stream2 := ffmpeg_go.Input(input2, nil)
	err := ffmpeg_go.Concat([]*ffmpeg_go.Stream{stream1, stream2}).Output(output).OverWriteOutput().ErrorToStdOut().Run()
	return err
}

func GetVideoProbeData(path string) (*ffprobe.ProbeData, error) {
	result, err := ffprobe.GetProbeData(path, 120000*time.Millisecond)
	return result, err
}
