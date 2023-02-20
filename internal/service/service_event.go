package service

import (
	"context"
	"encoding/json"
	"error_bot/config"
	"error_bot/internal/bot"
	"error_bot/internal/client"
	"error_bot/internal/dao"
	"error_bot/internal/types"
	"error_bot/internal/user"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"strconv"
)

const (
	reqGroupAddUrl = "/set_group_add_request"
)

func AcceptInvite(event *types.Event) {
	url := fmt.Sprintf(
		config.Content.BotServerConfig.Address +
			":" +
			strconv.Itoa(config.Content.BotServerConfig.Port) +
			reqGroupAddUrl)

	formattedData := g.Map{
		"flag":     event.Flag,
		"sub_type": event.SubType,
		"approve":  true,
	}

	client.Post(context.Background(), url, formattedData)

	bot.SendMessage(event.Sender.UserId, "private", "我接受邀请啦")
}

func ReplaceServiceConfig(event *types.Event) {
	var serviceConfig = g.Map{}
	err := json.Unmarshal([]byte(event.Message), &serviceConfig)
	if err != nil {
		bot.SendMessage(
			event.UserId,
			event.MessageType,
			fmt.Sprintf("ERROR: failed to unmarshal command, err(%s)", err.Error()),
		)
	}

	if event.GroupId != 0 && gconv.Int(serviceConfig["group_id"]) == 0 {
		serviceConfig["group_id"] = event.GroupId
	}

	if gconv.Int(serviceConfig["group_id"]) == 0 {
		bot.SendMessage(
			event.UserId,
			event.MessageType,
			fmt.Sprintf("ERROR: not found group ID."),
		)
	}

	if gconv.Int(serviceConfig["room_id"]) == 0 {
		bot.SendMessage(
			event.UserId,
			event.MessageType,
			fmt.Sprintf("ERROR: not found room ID."),
		)
	}

	liveRoomInfo := getLiveStatus(serviceConfig)
	if liveRoomInfo == nil {
		bot.SendMessage(
			event.UserId,
			event.MessageType,
			fmt.Sprintf("ERROR: not found live room."),
		)
	}

	userInfo := user.GetUserInfo(context.Background(), gconv.Int(liveRoomInfo["uid"]))
	if userInfo == nil {
		bot.SendMessage(
			event.UserId,
			event.MessageType,
			fmt.Sprintf("ERROR: not found user."),
		)
	}

	var allBilibiliServiceConfig = &types.BilibiliService{
		Name:              userInfo.Data.Card.Name,
		UserID:            gconv.Int(liveRoomInfo["uid"]),
		GroupID:           gconv.Int(serviceConfig["group_id"]),
		RoomID:            gconv.Int(serviceConfig["room_id"]),
		LiveNotification:  gconv.Int(serviceConfig["live_notification"]),
		SpaceNotification: gconv.Int(serviceConfig["space_notification"]),
		AtAll:             gconv.Int(serviceConfig["at_all"]),
		Enable:            1,
	}

	err = dao.ReplaceServiceInfo(allBilibiliServiceConfig)
	if err != nil {
		bot.SendMessage(
			event.UserId,
			event.MessageType,
			fmt.Sprintf("ERROR: failed replace service config to database, err(%s).", err),
		)
	}

	CreateObject(allBilibiliServiceConfig)
}
