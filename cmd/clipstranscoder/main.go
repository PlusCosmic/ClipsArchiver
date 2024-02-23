package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/u2takey/ffmpeg-go"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

const inputPath = "/uploads/"
const outputPath = "/archive/"

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
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	encoder := "hevc_videotoolbox"
	if runtime.GOOS == "windows" {
		encoder = "h264_nvenc"
	}
	encodeErr := ffmpeg_go.Input(exPath+inputPath+queueEntry.Filename).Output(exPath+outputPath+queueEntry.Filename, ffmpeg_go.KwArgs{"c:v": encoder, "q:v": 65}).OverWriteOutput().ErrorToStdOut().Run()
	if encodeErr != nil {
		fmt.Println("Something went wrong when transcoding")
		fmt.Println(err)
	}
}
