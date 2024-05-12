package main

import (
	"ClipsArchiver/internal/config"
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/media"
	"ClipsArchiver/internal/rabbitmq"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
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

	jobs := make(chan db.TranscodeRequest)
	go receiveTranscodeClipVTB(jobs)

	var forever chan struct{}

	channel, err := rabbitmq.GetConsumeChannel()
	if err != nil {
		log.Fatalf("Failed to get RabbitMQ queue: %s", err.Error())
	}

	for m := range channel {
		slog.Debug(fmt.Sprintf("Recieved message with queue id: %s", string(m.Body)))
		items := strings.Split(string(m.Body), ",")
		if len(items) != 2 {
			slog.Debug(fmt.Sprintf("Incorrect request format: %s", string(m.Body)))
			err = m.Reject(false)
			continue
		}

		id, err := strconv.Atoi(items[0])
		if err != nil {
			slog.Debug(fmt.Sprintf("Failed to parse item to id: %s", items[0]))
			err = m.Reject(false)
			if err != nil {
				continue
			}
		}

		requestType, err := strconv.Atoi(items[1])
		if err != nil {
			slog.Debug(fmt.Sprintf("Failed to parse item to request type: %s", items[1]))
			err = m.Reject(false)
			if err != nil {
				continue
			}
		}

		if requestType != 0 {
			continue
		}

		queueEntry, err := db.GetTranscodeRequestById(id)
		if err != nil {
			slog.Debug(fmt.Sprintf("Failed to find queueEntry for id: %s", string(m.Body)))
			err = m.Reject(false)
			continue
		}
		err = m.Ack(false)
		if err != nil {
			continue
		}
		jobs <- queueEntry
	}

	<-forever
}

func receiveTranscodeClipVTB(jobs <-chan db.TranscodeRequest) {
	for q := range jobs {
		transcodeClipVTB(q)
	}
}

func transcodeClipVTB(queueEntry db.TranscodeRequest) {
	err := db.UpdateTranscodeRequestStatusToTranscoding(queueEntry.ClipId)
	if err != nil {
		logger.Error("Failed to modify database entry: tried to set queue entry %d to transcoding", queueEntry.Id)
		return
	}

	clip, err := db.GetClipById(queueEntry.ClipId)
	if err != nil {
		logger.Error("Failed to get clip for id: %d", queueEntry.Id)
		return
	}

	inputPath := config.GetInputPath() + clip.Filename
	outputPath := config.GetOutputPath() + clip.Filename
	fmt.Printf("Starting transcode on %s\n", clip.Filename)
	err = media.TranscodeVideoFile(inputPath, outputPath)
	if err != nil {
		err = db.UpdateTranscodeRequestStatusToError(queueEntry.ClipId, "Failed to transcode video file")
		return
	}

	imagePath := config.GetThumbnailsPath() + clip.Filename + ".png"
	err = media.GenerateThumbnailFromVideo(outputPath, imagePath)
	if err != nil {
		err = db.UpdateTranscodeRequestStatusToError(queueEntry.ClipId, "Failed to generate video thumbnail")
		return
	}

	err = db.UpdateTranscodeRequestStatusToFinished(queueEntry.ClipId)
	if err != nil {
		err = db.UpdateTranscodeRequestStatusToError(queueEntry.ClipId, "Failed to modify database entry")
		logger.Error("Failed to modify database entry: tried to set queue entry %d to finished", queueEntry.Id)
		return
	}

	probeData, err := media.GetVideoProbeData(outputPath)
	err = db.UpdateClipOnTranscodeFinish(queueEntry.ClipId, probeData.Format.DurationSeconds)
	if err != nil {
		err = db.UpdateTranscodeRequestStatusToError(queueEntry.ClipId, "Failed to modify database entry")
		logger.Error("Failed to modify database entry: tried to set queue entry %d to error", queueEntry.Id)
		return
	}
}

func combineClips(entry db.TranscodeRequest) {

}

func trimClip(entry db.TranscodeRequest) {

}
