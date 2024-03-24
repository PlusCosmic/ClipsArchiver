package main

import (
	"ClipsArchiver/internal/config"
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/media"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"
)

const logFileLocation = "clipstranscoder.log"

var logger *slog.Logger

func main() {
	options := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	file, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to get log file handle: %s", err.Error())
	}

	var handler slog.Handler = slog.NewJSONHandler(file, options)
	logger = slog.New(handler)

	err = db.SetupDb(logger)
	if err != nil {
		log.Fatalf("Failed to setup database: %s", err.Error())
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
	slog.Debug("Checking for Queue Entries")
	queueEntries, err := db.GetAllPendingQueueEntries()
	if err != nil {
		logger.Error("Failed to get pending Queue Entries")
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
		logger.Error("Failed to modify database entry: tried to set queue entry %d to transcoding", queueEntry.Id)
		return
	}

	inputPath := config.GetInputPath() + queueEntry.Filename
	outputPath := config.GetOutputPath() + queueEntry.Filename
	err = media.TranscodeVideoFile(inputPath, outputPath)
	if err != nil {
		err = db.UpdateQueueEntryStatusToError(queueEntry.ClipId, "Failed to transcode video file")
		return
	}

	imagePath := config.GetThumbnailsPath() + queueEntry.Filename + ".png"
	err = media.GenerateThumbnailFromVideo(outputPath, imagePath)
	if err != nil {
		err = db.UpdateQueueEntryStatusToError(queueEntry.ClipId, "Failed to generate video thumbnail")
		return
	}

	err = db.UpdateQueueEntryStatusToFinished(queueEntry.ClipId)
	if err != nil {
		err = db.UpdateQueueEntryStatusToError(queueEntry.ClipId, "Failed to modify database entry")
		logger.Error("Failed to modify database entry: tried to set queue entry %d to finished", queueEntry.Id)
		return
	}

	probeData, err := media.GetVideoProbeData(outputPath)
	err = db.UpdateClipOnTranscodeFinish(queueEntry.ClipId, probeData.Format.DurationSeconds)
	if err != nil {
		err = db.UpdateQueueEntryStatusToError(queueEntry.ClipId, "Failed to modify database entry")
		logger.Error("Failed to modify database entry: tried to set queue entry %d to error", queueEntry.Id)
		return
	}
}
