package main

import (
	"ClipsArchiver/internal/config"
	"ClipsArchiver/internal/db"
	"ClipsArchiver/internal/rabbitmq"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type matchHistory struct {
	Uid                string `json:"uid"`
	LegendPlayed       string `json:"legendPlayed"`
	GameStartTimestamp int64  `json:"gameStartTimestamp"`
	GameEndTimestamp   int64  `json:"gameEndTimestamp"`
	BrScoreChange      int    `json:"BRScoreChange"`
	BrRankImg          string `json:"BRRankImg"`
	Map                string `json:"map"`
	MatchHash          string `json:"matchHash"`
}

type MapRotationInfo struct {
	Map           string `json:"map"`
	RemainingMins int    `json:"remainingMins"`
}

type MapRotation struct {
	Current MapRotationInfo `json:"current"`
	Next    MapRotationInfo `json:"next"`
}

const logFileLocation = "matchhistoryprocessor.log"

var currentMapString = ""

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

	_ = getMatchHistoryForAllUsers()
	_ = processMatchHistoriesForRecentClips()
	//main loop
	for i := 0; true; i++ {
		time.Sleep(5 * time.Second)
		_ = getMatchHistoryForAllUsers()
		_ = processMatchHistoriesForRecentClips()
		_ = getMapUpdate()
	}
}

func getMapUpdate() error {
	url := fmt.Sprintf("https://api.mozambiquehe.re/maprotation?auth=%s", config.GetApiKey())
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	mapRotation := new(MapRotation)
	err = json.NewDecoder(resp.Body).Decode(mapRotation)
	if err != nil {
		return err
	}

	if mapRotation.Current.Map != currentMapString {
		currentMapString = mapRotation.Current.Map
		mun := rabbitmq.MapUpdateNotification{
			MapName:         currentMapString,
			DurationMinutes: mapRotation.Current.RemainingMins,
		}
		err = rabbitmq.PublishMapUpdateNotification(mun)
	}
	return err
}

func getMatchHistoryForAllUsers() error {
	allUsers, err := db.GetAllUsers()
	if err != nil {
		return err
	}
	for _, user := range allUsers {
		url := fmt.Sprintf("https://api.mozambiquehe.re/games?auth=%s&uid=%s", config.GetApiKey(), user.ApexUid)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		matchHistories := new([]matchHistory)
		err = json.NewDecoder(resp.Body).Decode(matchHistories)
		if err != nil {
			return err
		}
		processMatchHistories(*matchHistories)
	}
	return err
}

func processMatchHistories(matchHistories []matchHistory) {
	for _, history := range matchHistories {
		user, err := db.GetUserByApexUid(history.Uid)
		if err != nil {
			continue
		}
		//do we already have it?
		hash := sha256.New()
		jsonString, err := json.Marshal(history)
		hash.Write(jsonString)
		matchHash := fmt.Sprintf("%x", hash.Sum(nil))
		_, err = db.GetMatchHistoryByMatchHash(matchHash)
		if err == nil {
			// we have it
			continue
		}
		//we don't have it so add it
		gameMap, err := db.GetMapByAlsName(history.Map)
		hasGameMap := true
		if err != nil {
			hasGameMap = false
		}
		legend, err := db.GetLegendByName(history.LegendPlayed)
		hasLegend := true
		if err != nil {
			hasLegend = false
		}
		gameMode := "Pubs"
		isRanked := false
		if history.BrScoreChange != 0 {
			gameMode = "Ranked"
			isRanked = true
		}

		newHist := db.MatchHistory{
			UserId:        user.Id,
			GameStart:     sql.NullTime{Valid: true, Time: time.Unix(history.GameStartTimestamp, 0)},
			GameEnd:       sql.NullTime{Valid: true, Time: time.Unix(history.GameEndTimestamp, 0)},
			Map:           sql.NullInt32{Int32: int32(gameMap.Id), Valid: hasGameMap},
			Legend:        sql.NullInt32{Int32: int32(legend.Id), Valid: hasLegend},
			GameMode:      gameMode,
			BrScoreChange: sql.NullInt32{Valid: isRanked, Int32: int32(history.BrScoreChange)},
			BrRankImg:     sql.NullString{Valid: isRanked, String: history.BrRankImg},
			MatchHash:     matchHash,
		}
		err = db.AddNewMatchHistory(newHist)
		if err != nil {
			continue
		}
	}
}

func processMatchHistoriesForRecentClips() error {
	var err error

	for i := 0; i < 14; i++ {
		clips, err := db.GetClipsForDate(time.Now().AddDate(0, 0, -1*i))
		if err == nil {
			processMatchHistoriesForClips(clips)
		}
	}
	return err
}

func processMatchHistoriesForClips(clips []db.Clip) {
	for _, clip := range clips {
		if clip.MatchHistoryFound {
			continue
		}
		matchHistories, err := db.GetMatchHistoriesForClip(clip)
		if err != nil {
			continue
		}
		if len(matchHistories) == 0 {
			continue
		}

		selectedHistory := matchHistories[len(matchHistories)-1]

		if selectedHistory.Map.Valid {
			clip.Map.Int32 = selectedHistory.Map.Int32
			clip.Map.Valid = true
		}

		if selectedHistory.Legend.Valid {
			clip.Legend.Int32 = selectedHistory.Legend.Int32
			clip.Legend.Valid = true
		}

		if selectedHistory.BrScoreChange.Valid {
			clip.BrScoreChange = selectedHistory.BrScoreChange
		}

		if selectedHistory.BrRankImg.Valid {
			clip.BrRankImg = selectedHistory.BrRankImg
		}

		clip.GameMode.String = selectedHistory.GameMode
		clip.GameMode.Valid = true
		clip.MatchHistoryFound = true
		_ = db.UpdateClip(clip)
	}
}
