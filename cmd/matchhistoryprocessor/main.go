package main

import (
	"ClipsArchiver/internal/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type matchHistory struct {
	Uid                string `json:"uid"`
	LegendPlayed       string `json:"legendPlayed"`
	GameStartTimestamp int64  `json:"gameStartTimestamp"`
	GameEndTimestamp   int64  `json:"gameEndTimestamp"`
	BrScoreChange      int    `json:"BRScoreChange"`
	Map                string `json:"map"`
}

const apiKey = "_"

func main() {
	err := db.SetupDb()
	if err != nil {
		return
	}
	_ = getMatchHistoryForAllUsers()
	_ = processMatchHistoriesForRecentClips()
	//main loop
	for i := 0; true; i++ {
		time.Sleep(5 * time.Second)
		_ = getMatchHistoryForAllUsers()
		_ = processMatchHistoriesForRecentClips()
	}
}

func getMatchHistoryForAllUsers() error {
	allUsers, err := db.GetAllUsers()
	if err != nil {
		return err
	}
	for _, user := range allUsers {
		url := fmt.Sprintf("https://api.mozambiquehe.re/games?auth=%s&uid=%s", apiKey, user.ApexUid)
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
		_, err = db.GetMatchHistoryByUserIdAndTimeStampRange(user.Id, time.Unix(history.GameStartTimestamp, 0), time.Unix(history.GameEndTimestamp, 0))
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
		if history.BrScoreChange != 0 {
			gameMode = "Ranked"
		}
		newHist := db.MatchHistory{
			UserId:    user.Id,
			GameStart: sql.NullTime{Valid: true, Time: time.Unix(history.GameStartTimestamp, 0)},
			GameEnd:   sql.NullTime{Valid: true, Time: time.Unix(history.GameEndTimestamp, 0)},
			Map:       sql.NullInt32{Int32: int32(gameMap.Id), Valid: hasGameMap},
			Legend:    sql.NullInt32{Int32: int32(legend.Id), Valid: hasLegend},
			GameMode:  gameMode,
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

		clip.GameMode.String = selectedHistory.GameMode
		clip.GameMode.Valid = true
		clip.MatchHistoryFound = true
		_ = db.UpdateClip(clip)
	}
}
