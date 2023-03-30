package models

import (
	"encoding/json"
	"fmt"
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

	// redis使用新增上線的人
	redisConn := RedisDefaultPool.Get()
	defer redisConn.Close()
	redisConn.Do("HSET", roomId, p.ParticipantStreamId, p.ParticipantId)
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

	// redis使用刪除上線的人
	redisConn := RedisDefaultPool.Get()
	defer redisConn.Close()
	redisConn.Do("HDEL", roomId, p.ParticipantStreamId)
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

	//redis使用抓全部的人
	fmt.Println("測試抓全部的人")
	redisConn := RedisDefaultPool.Get()
	defer redisConn.Close()
	redisdata, err := redis.Values(redisConn.Do("HGETALL", roomUuid))
	if err != nil {
		fmt.Println("redis get failed", err)
	}
	fmt.Println(redisdata)
	pcpInRoomWithAvatarMap := make(map[string]string)
	for i := 0; i < len(redisdata); i += 2 {
		key := string(redisdata[i].([]byte))
		value := string(redisdata[i+1].([]byte))
		pcpInRoomWithAvatarMap[key] = value
	}
	fmt.Println(pcpInRoomWithAvatarMap)
	fmt.Println(pcpInRoomWithAvatar)

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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"pcpId": specificPcp[0].PcpId, "pcpName": specificPcp[0].PcpName, "pcpAvatar": specificPcp[0].PcpAvatarUrl})
}

// 暫時======================================

type cacheStruct struct {
	CacheUser cacheUserStruct
}

type cacheUserStruct struct {
	UserId   string `json:"userId"`
	StreamId string `json:"streamId"`
}

func CacheOneUser(c *fiber.Ctx) error {
	testNumber := c.Params("number")
	redisRoom := "testroom"
	redisKey := fmt.Sprintf("user_%s", testNumber)
	conn := RedisDefaultPool.Get()
	defer conn.Close()
	data, err := redis.Bytes(conn.Do("GET", redisRoom))
	if err != nil {
		RedisOneUser(c)
		dbResult := RedisOneUser(c)
		fmt.Println(dbResult)
		redisData, _ := json.Marshal(dbResult)
		fmt.Println(redisData)
		fmt.Println(string(redisData))
		// conn.Do("SETEX", redisKey, 40, redisData)
		conn.Do("HSET", redisRoom, redisKey, redisData)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": "from db", "data": dbResult})
	}
	var cachedData cacheStruct
	json.Unmarshal(data, &cachedData)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": "from redis", "data": cachedData})

}

// Redis User
func RedisOneUser(c *fiber.Ctx) cacheStruct {
	return cacheStruct{CacheUser: cacheUserStruct{UserId: "a1012", StreamId: "poiuyt"}}
}
