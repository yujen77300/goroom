package models

import (
	// "log"

	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/spf13/viper"
)

type pcpsInRoom struct {
	Participant      string `json:"participant"`
	ParticipantId    int    `json:"participantId"`
	ParticipantEmail string `json:"participantEmail"`
}

func UpdateParticipantInfo(participantInfo []byte, roomId string, participantuuid string) {
	var p pcpsInRoom
	err := json.Unmarshal(participantInfo, &p)
	if err != nil {
		fmt.Println("Error during json to struct")
		return
	}

	db, _ := ConnectToMYSQL()
	_, err = db.Exec("INSERT INTO participant(room_id,member_id,participant_uuid) values(?,?,?);", roomId, p.ParticipantId, participantuuid)
	if err != nil {
		fmt.Println("Insert participant failed")
		return
	}
	defer db.Close()
}
