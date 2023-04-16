package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
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

	redisConn := RedisDefaultPool.Get()
	defer redisConn.Close()
	redisConn.Do("HSET", roomId, p.ParticipantStreamId, p.ParticipantId)
	redisConn.Do("EXPIRE", roomId, 86400) 
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

	redisConn := RedisDefaultPool.Get()
	defer redisConn.Close()
	redisConn.Do("HDEL", roomId, p.ParticipantStreamId)
}

func GetAllPcpInRoom(c *fiber.Ctx) error {
	roomUuid := c.Params("roomUuid")
	roomUuid = strings.TrimLeft(roomUuid, ":")
	redisConn := RedisDefaultPool.Get()
	defer redisConn.Close()
	redisData, err := redis.Values(redisConn.Do("HGETALL", roomUuid))
	if err != nil {
		fmt.Println("redis HGETALL failed", err)
	}

	if len(redisData) == 0 {
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
			redisConn.Do("HSET", roomUuid, eachPcp.PcpStreamId, eachPcp.PcpId)
			redisConn.Do("EXPIRE", roomUuid, 86400) 
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"allPcps": pcpInRoomWithAvatar})
	} else {
		pcpInRoomWithAvatarMap := make(map[string]string)
		var pcpInRoomWithAvatar2 []PcpInRoomWithAvatar
		for i := 0; i < len(redisData); i += 2 {

			key := string(redisData[i].([]byte))
			valueString := string(redisData[i+1].([]byte))
			pcpInRoomWithAvatarMap[key] = valueString

			value, err := strconv.Atoi(valueString)
			if err != nil {
				fmt.Println("convert value to int failed", err)
			}

			redisUserData, err := redis.Strings(redisConn.Do("HMGET", valueString, "cacheName", "cacheAvatarUrl"))
			if err != nil {
				fmt.Println("redis get failed", err)
			}

			eachPcp2 := PcpInRoomWithAvatar{
				PcpId:        value,
				PcpName:      redisUserData[0],
				PcpAvatarUrl: redisUserData[1],
				PcpStreamId:  key,
			}
			pcpInRoomWithAvatar2 = append(pcpInRoomWithAvatar2, eachPcp2)
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"allPcps": pcpInRoomWithAvatar2})
	}
}

func GetPcpInfo(c *fiber.Ctx) error {
	roomUuid := c.Params("roomUuid")
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"pcpUserId": specificPcp[0].PcpId, "pcpName": specificPcp[0].PcpName})
}
