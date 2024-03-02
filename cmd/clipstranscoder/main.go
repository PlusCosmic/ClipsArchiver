package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/u2takey/ffmpeg-go"
	"github.com/vansante/go-ffprobe"
	"log"
	"time"
)

var db *sql.DB

type QueueEntry struct {
	Id         int
	ClipId     int
	Status     string
	StartedAt  sql.NullTime
	FinishedAt sql.NullTime
	Filename   string
}

const inputPath = "/Uploads/"
const outputPath = "/Clips/"
const thumbnailsPath = "/Clips/Thumbnails/"
const storePath = "/Volumes/Big Store/TheArchive"

func main() {
	cfg := mysql.Config{
		User:      "clips_rest_user",
		Passwd:    "123",
		Net:       "tcp",
		Addr:      "10.0.0.10",
		DBName:    "clips_archiver",
		ParseTime: true,
	}
	var dbErr error
	db, dbErr = sql.Open("mysql", cfg.FormatDSN())
	if dbErr != nil {
		log.Fatal(dbErr)
		return
	}

	//main loop
	for i := 0; true; i++ {
		time.Sleep(2 * time.Second)
		checkForQueueEntries()
	}
}

func checkForQueueEntries() {
	queueEntries, err := getClipsQueueInternal()
	if err != nil {
		fmt.Println("Something went wrong in fetching the clips queue :(")
		return
	}
	for _, queueEntry := range queueEntries {
		transcodeClip(queueEntry)
	}

}

func getClipsQueueInternal() ([]QueueEntry, error) {
	println("Checking for new queue entries")
	var queueEntries []QueueEntry

	rows, dbErr := db.Query("SELECT clips_queue.*,clips.filename FROM clips_queue INNER JOIN clips on clips_queue.clip_id = clips.id where status = 'pending'")
	if dbErr != nil {
		return nil, dbErr
	}

	defer rows.Close()

	for rows.Next() {
		var queueEntry QueueEntry
		if err := rows.Scan(&queueEntry.Id, &queueEntry.ClipId, &queueEntry.Status, &queueEntry.StartedAt, &queueEntry.FinishedAt, &queueEntry.Filename); err != nil {
			return nil, err
		}
		queueEntries = append(queueEntries, queueEntry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return queueEntries, nil
}

func transcodeClip(queueEntry QueueEntry) {
	fmt.Printf("Starting transcode on %s\n", queueEntry.Filename)
	_, dbErr := db.Exec("UPDATE clips_queue SET clips_queue.status = 'transcoding', clips_queue.started_at = ? WHERE clips_queue.clip_id = ?", time.Now(), queueEntry.ClipId)
	if dbErr != nil {
		return
	}

	encoder := "hevc_videotoolbox"

	encodeErr := ffmpeg_go.Input(storePath+inputPath+queueEntry.Filename).Output(storePath+outputPath+queueEntry.Filename, ffmpeg_go.KwArgs{"c:v": encoder, "q:v": 65, "vf": "scale=1920:1080"}).OverWriteOutput().ErrorToStdOut().Run()
	if encodeErr != nil {
		fmt.Println("Something went wrong when transcoding")
		fmt.Println(encodeErr)
	}

	thumbnailErr := ffmpeg_go.Input(storePath+inputPath+queueEntry.Filename).Output(storePath+thumbnailsPath+queueEntry.Filename+".png", ffmpeg_go.KwArgs{"ss": "00:00:01.000", "frames:v": 1}).OverWriteOutput().ErrorToStdOut().Run()
	if thumbnailErr != nil {
		fmt.Println("Something went wrong when creating thumbnail")
		fmt.Println(thumbnailErr)
	}

	result, err := ffprobe.GetProbeData(storePath+outputPath+queueEntry.Filename, 120000*time.Millisecond)
	if err != nil {
		return
	}

	_, dbErr = db.Exec("UPDATE clips_queue SET clips_queue.status = 'finished', clips_queue.finished_at = ? WHERE clips_queue.clip_id = ?", time.Now(), queueEntry.ClipId)
	if dbErr != nil {
		return
	}
	_, dbErr = db.Exec("UPDATE clips SET clips.is_processed = 1, clips.duration = ? WHERE clips.id = ?", result.Format.DurationSeconds, queueEntry.ClipId)
	if dbErr != nil {
		return
	}

}
