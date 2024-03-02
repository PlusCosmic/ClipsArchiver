package db

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"time"
)

var db *sql.DB

type User struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	ApexUsername string `json:"apexUsername"`
	ApexUid      string `json:"apexUid"`
}

type Clip struct {
	Id          int      `json:"id"`
	OwnerId     int      `json:"ownerId"`
	Filename    string   `json:"filename"`
	IsProcessed bool     `json:"isProcessed"`
	CreatedAt   string   `json:"createdOn"`
	Duration    int      `json:"duration"`
	Tags        []string `json:"tags"`
}

type QueueEntry struct {
	Id         int    `json:"id"`
	ClipId     int    `json:"clipId"`
	Status     string `json:"status"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

type Tag struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func SetupDb() error {
	cfg := mysql.Config{
		User:   "clips_rest_user",
		Passwd: "123",
		Net:    "tcp",
		Addr:   "10.0.0.10",
		DBName: "clips_archiver",
	}
	var dbErr error
	db, dbErr = sql.Open("mysql", cfg.FormatDSN())
	return dbErr
}

func GetAllUsers() ([]User, error) {
	var users []User

	rows, dbErr := db.Query("SELECT * FROM users")
	if dbErr != nil {
		return nil, dbErr
	}

	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.Name, &user.ApexUsername, &user.ApexUid); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func GetClipsQueue() ([]QueueEntry, error) {
	return nil, nil
}

func GetClipsForDate(dateOf time.Time) ([]Clip, error) {
	var clips []Clip

	dateAfter := dateOf.AddDate(0, 0, 1)

	rows, dbErr := db.Query("SELECT * FROM clips WHERE clips.created_at >= ? AND clips.created_at < ?", dateOf, dateAfter)
	if dbErr != nil {
		return nil, dbErr
	}

	defer rows.Close()

	for rows.Next() {
		var clip Clip
		if err := rows.Scan(&clip.Id, &clip.OwnerId, &clip.Filename, &clip.IsProcessed, &clip.CreatedAt, &clip.Duration); err != nil {
			return nil, err
		}

		tags, err := GetTagsForClip(clip.Id)
		if err == nil {
			clip.Tags = tags
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clips, nil
}

func AddClip(ownerId int, filename string) error {
	clip, err := db.Exec("INSERT INTO clips (owner_id, filename, is_processed, created_at) VALUES (?, ?, ?, ?)", ownerId, filename, 0, time.Now())
	if err != nil {
		return err
	}

	id, err := clip.LastInsertId()
	if err != nil {
		return err
	}

	err = AddClipToQueue(int(id))
	return err
}

func AddClipToQueue(clipId int) error {
	_, err := db.Exec("INSERT INTO clips_queue (clip_id, status) VALUES (?, ?)", clipId, "pending")
	return err
}

func GetTagsForClip(clipId int) ([]string, error) {
	var tags []Tag

	rows, dbErr := db.Query("SELECT tag_id,name FROM clips_tags INNER JOIN tags ON clips_tags.tag_id = tags.id WHERE clips_tags.clip_id = ?", clipId)
	if dbErr != nil {
		return nil, dbErr
	}

	defer rows.Close()

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return tagNames, nil
}

func UpdateClip(old Clip, new Clip) error {
	old.Tags = new.Tags
	var err error
	for _, tag := range old.Tags {
		var existingTag Tag
		row := db.QueryRow("SELECT * FROM tags WHERE tags.name = ?", tag)

		err = row.Scan(&existingTag.Id, &existingTag.Name)
		if err != nil {
			_, err = db.Exec("INSERT INTO tags (name) VALUES (?)", tag)
			if err != nil {
				return err
			}
			row = db.QueryRow("SELECT * FROM tags WHERE tags.name = ?", tag)
			err = row.Scan(&existingTag.Id, &existingTag.Name)
			if err != nil {
				return err
			}
		}

		_, err = db.Exec("INSERT INTO clips_tags (clip_id, tag_id) VALUES (?, ?)", old.Id, existingTag.Id)
		if err != nil {
			return err
		}
	}
	return err
}

func GetClipById(clipId int) (Clip, error) {
	var clip Clip
	row := db.QueryRow("SELECT * FROM clips WHERE clips.id = ?", clipId)

	err := row.Scan(&clip.Id, &clip.OwnerId, &clip.Filename, &clip.IsProcessed, &clip.CreatedAt, &clip.Duration)
	return clip, err
}
