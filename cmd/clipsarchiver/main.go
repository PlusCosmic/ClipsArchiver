package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	CreatedOn   string   `json:"createdOn"`
	Tags        []string `json:"tags"`
}

type QueueEntry struct {
	Id         int    `json:"id"`
	ClipId     int    `json:"clipId"`
	Status     string `json:"status"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

func main() {
	cfg := mysql.Config{
		User:   "clips_rest_user",
		Passwd: "123",
		Net:    "tcp",
		Addr:   "10.0.0.10",
		DBName: "clips_archiver",
	}
	var dbErr error
	db, dbErr = sql.Open("mysql", cfg.FormatDSN())
	if dbErr != nil {
		log.Fatal(dbErr)
		return
	}
	// Create Gin router
	router := gin.Default()

	// Register Routes
	router.POST("/clips/upload", uploadClip)
	router.GET("/users", getAllUsers)
	router.GET("/clips/queue", getClipsQueue)
	/*	// YYYY-MM-DD
		router.GET("/clips/date/:date", getClipsForDate)*/

	// Start the server
	routerErr := router.Run()
	if routerErr != nil {
		log.Fatal(routerErr)
		return
	}
}

func uploadClip(c *gin.Context) {
	// Single file
	file, err := c.FormFile("file")
	log.Println(file.Filename)

	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	filename := exPath + "/uploads/" + filepath.Base(file.Filename)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}

func getAllUsers(c *gin.Context) {
	users, err := getAllUsersInternal()
	if err != nil {
		c.String(http.StatusInternalServerError, "Something went wrong :(")
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

func getAllUsersInternal() ([]User, error) {
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

func getClipsQueue(c *gin.Context) {
	queueEntries, err := getAllUsersInternal()
	if err != nil {
		c.String(http.StatusInternalServerError, "Something went wrong :(")
		return
	}
	c.IndentedJSON(http.StatusOK, queueEntries)
}

/*func getClipsQueueInternal(c *gin.Context) {

}*/

/*func getClipsForDate(c *gin.Context) {
	date := c.Param("day")
	values := strings.Split(date, "-")
}

func getClipsForDateInternal(string date) {

}*/
