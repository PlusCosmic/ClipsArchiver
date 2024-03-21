package main

import (
	"ClipsArchiver/internal/config"
	"ClipsArchiver/internal/db"
	"fmt"
	"github.com/u2takey/ffmpeg-go"
	"github.com/vansante/go-ffprobe"
	"time"
)

func main() {
	err := db.SetupDb()
	if err != nil {
		return
	}
	jobs := make(chan db.QueueEntry)

	go receiveTranscodeClipVTB(jobs)
	//main loop
	for i := 0; true; i++ {
		time.Sleep(2 * time.Second)
		checkForQueueEntries(jobs)
	}
}

func checkForQueueEntries(jobs chan<- db.QueueEntry) {
	queueEntries, err := db.GetAllPendingQueueEntries()
	if err != nil {
		fmt.Println("Something went wrong in fetching the clips queue :(")
		return
	}

	for _, queueEntry := range queueEntries {
		err := db.UpdateQueueEntryStatusToQueued(queueEntry.ClipId)
		if err != nil {
			continue
		}
		jobs <- queueEntry
	}

}

func receiveTranscodeClipVTB(jobs <-chan db.QueueEntry) {
	for q := range jobs {
		transcodeClipVTB(q)
	}
}

func transcodeClipVTB(queueEntry db.QueueEntry) {
	fmt.Printf("Starting transcode on %s\n", queueEntry.Filename)
	err := db.UpdateQueueEntryStatusToTranscoding(queueEntry.ClipId)
	if err != nil {
		return
	}

	encoder := "h264_videotoolbox"

	encodeErr := ffmpeg_go.Input(config.GetInputPath()+queueEntry.Filename).Output(config.GetOutputPath()+queueEntry.Filename, ffmpeg_go.KwArgs{"c:v": encoder, "q:v": 65, "vf": "scale=1920:1080"}).OverWriteOutput().ErrorToStdOut().Run()
	if encodeErr != nil {
		fmt.Println("Something went wrong when transcoding")
		fmt.Println(encodeErr)
	}

	thumbnailErr := ffmpeg_go.Input(config.GetInputPath()+queueEntry.Filename).Output(config.GetThumbnailsPath()+queueEntry.Filename+".png", ffmpeg_go.KwArgs{"ss": "00:00:01.000", "frames:v": 1}).OverWriteOutput().ErrorToStdOut().Run()
	if thumbnailErr != nil {
		fmt.Println("Something went wrong when creating thumbnail")
		fmt.Println(thumbnailErr)
	}

	result, err := ffprobe.GetProbeData(config.GetOutputPath()+queueEntry.Filename, 120000*time.Millisecond)
	if err != nil {
		return
	}

	err = db.UpdateQueueEntryStatusToFinished(queueEntry.ClipId)
	if err != nil {
		return
	}

	err = db.UpdateClipOnTranscodeFinish(queueEntry.ClipId, result.Format.DurationSeconds)
}
