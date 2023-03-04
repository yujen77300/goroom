package models

import (
	// "log"

	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	// "github.com/spf13/viper"
)

type pcpInRoom struct {
	Participant         string `json:"pcpName"`
	ParticipantId       int    `json:"pcpId"`
	ParticipantEmail    string `json:"pcpEmail"`
	ParticipantStreamId string `json:"streamId"`
}

type pcpOutRoom struct {
	ParticipantStreamId string `json:"streamId"`
}

type PcpInRoomWithAvatar struct {
	PcpId        int    `json:"id"`
	PcpName      string `json:"username"`
	PcpAvatarUrl string `json:"avatar_url"`
	PcpStreamId  string `json:"pcp_stream_url"`
}

type SpecificPcp struct {
	PcpId        int    `json:"id"`
	PcpName      string `json:"username"`
	PcpAvatarUrl string `json:"avatar_url"`
}

func UpdateParticipantInfo(participantInfo []byte, roomId string) {
	var p pcpInRoom
	err := json.Unmarshal(participantInfo, &p)
	if err != nil {
		fmt.Println("Error during json to struct")
		return
	}

	db, _ := ConnectToMYSQL()
	_, err = db.Exec("INSERT INTO participant(room_id,member_id,pcp_stream_id) values(?,?,?);", roomId, p.ParticipantId, p.ParticipantStreamId)
	if err != nil {
		fmt.Println("Insert participant failed")
		return
	}
	defer db.Close()
}

func DeleteParticipantInfo(participantInfo []byte, roomId string) {
	var p pcpOutRoom
	err := json.Unmarshal(participantInfo, &p)
	if err != nil {
		fmt.Println("Error during json to struct")
		return
	}

	db, _ := ConnectToMYSQL()
	_, err = db.Exec("DELETE FROM participant WHERE pcp_stream_id = ? and room_id= ? ;", p.ParticipantStreamId, roomId)
	if err != nil {
		fmt.Println("Delete participant failed")
		return
	}
	defer db.Close()
}

func GetAllPcpInRoom(c *fiber.Ctx) error {
	roomUuid := c.Params("uuid")
	roomUuid = strings.TrimLeft(roomUuid, ":")
	db, _ := ConnectToMYSQL()
	rows, err := db.Query("SELECT member.id,member.username,member.avatar_url,participant.pcp_stream_id FROM member JOIN participant ON member.id = participant.member_id where room_id=?;", roomUuid)
	if err != nil {
		fmt.Printf("Database query failed, error:%v\n", err)
	}
	defer rows.Close()
	defer db.Close()
	var pcpInRoomWithAvatar []PcpInRoomWithAvatar
	for rows.Next() {
		var eachPcp PcpInRoomWithAvatar
		if dberr := rows.Scan(&eachPcp.PcpId, &eachPcp.PcpName, &eachPcp.PcpAvatarUrl, &eachPcp.PcpStreamId); dberr != nil {
			fmt.Printf("scan failed, err:%v\n", dberr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scan failed"})
		}
		pcpInRoomWithAvatar = append(pcpInRoomWithAvatar, eachPcp)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"allpcps": pcpInRoomWithAvatar})
}

func GetPcpInfo(c *fiber.Ctx) error {
	roomUuid := c.Params("uuid")
	streamId := c.Params("streamId")
	roomUuid = strings.TrimLeft(roomUuid, ":")
	streamId = strings.TrimLeft(streamId, ":")
	db, _ := ConnectToMYSQL()
	row, err := db.Query("SELECT member.id,member.username,member.avatar_url FROM member JOIN participant ON member.id = participant.member_id where room_id=? and pcp_stream_id=?;", roomUuid, streamId)
	if err != nil {
		fmt.Printf("Database query failed, error:%v\n", err)
	}
	defer row.Close()
	defer db.Close()
	var specificPcp []SpecificPcp
	for row.Next() {
		var Pcp SpecificPcp
		if dberr := row.Scan(&Pcp.PcpId, &Pcp.PcpName, &Pcp.PcpAvatarUrl); dberr != nil {
			fmt.Printf("scan failed, err:%v\n", dberr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scan failed"})
		}
		specificPcp = append(specificPcp, Pcp)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"pcpId": specificPcp[0].PcpId,"pcpName":specificPcp[0].PcpName,"pcpAvatar":specificPcp[0].PcpAvatarUrl})
}
