package service

import (
	"context"
	"encoding/json"
	"error_bot/internal/bot"
	"error_bot/internal/types"
	"error_bot/internal/units"
	"error_bot/internal/user"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
	"io"
	"log"
	"strings"
)

// EventService 监听QQ事件
func EventService(req *ghttp.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to read request body, err(%s)", err))
		units.HttpResponseError(req, g.Map{
			"msg": err.Error(),
		})
		return
	}

	event := &types.Event{}
	err = json.Unmarshal(data, event)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to read request body, err(%s)", err))
		units.HttpResponseError(req, g.Map{
			"msg": err.Error(),
		})
		return
	}

	go handleEvent(event)

}

func handleEvent(event *types.Event) {
	switch event.PostType {
	case "message":
		handleMessage(event)
		return
	case "request":
		if event.SubType == "add" || event.SubType == "invite" {
			AcceptInvite(event)
		}
		return
	case "meta_event":
	}
}

func handleMessage(event *types.Event) {
	switch event.MessageType {
	case "group":
	case "private":
		if len(event.Message) > 5 && event.Message[:8] == "config:" {
			event.Message = strings.Replace(event.Message, "config:", "", -1)
			ReplaceServiceConfig(event)
			return
		}
		bot.SendMessage(event.Sender.UserId, "private", RequestXiaoAi(event.Message))
	}
}

func Manage(req *ghttp.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to read request body, err(%s)", err))
		units.HttpResponseError(req, g.Map{
			"msg": err.Error(),
		})
		return
	}

	manage := &types.Manage{}
	err = json.Unmarshal(data, manage)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to read request body, err(%s)", err))
		units.HttpResponseError(req, g.Map{
			"msg": err.Error(),
		})
		return
	}

	var resData map[string]interface{}
	switch manage.Operation {
	case "getLiveStatus":
		resData = getLiveStatus(manage.Data)
	case "getLastDynamic":
		resData = getLastDynamic(manage.Data)
	}

	units.HttpResponseSuccess(req, resData)
}

func getLiveStatus(data map[string]interface{}) map[string]interface{} {
	roomId := gconv.Int(data["room_id"])

	roomInfo := user.GetLiveRoomInfo(context.Background(), roomId)
	if roomInfo == nil {
		log.Println("INFO: nil live room info")
		return nil
	}

	var resData = g.Map{
		"uid":         roomInfo.Data.Uid,
		"live_status": roomInfo.Data.LiveStatus,
		"cover":       roomInfo.Data.UserCover,
	}

	return resData
}

func getLastDynamic(data map[string]interface{}) map[string]interface{} {
	userId := gconv.Int(data["uid"])

	dynamicInfo := user.GetDynamicInfo(context.Background(), userId)
	if dynamicInfo == nil {
		log.Println("INFO: nil  live room info")
		return nil
	}

	if len(dynamicInfo.Data.Items) < 1 {
		return g.Map{}
	}

	var formattedTopDynamic, formattedDynamic string

	for i, dynamicItems := range dynamicInfo.Data.Items {
		if i == 0 && dynamicItems.Modules.ModuleTag.Text == "置顶" {
			formattedTopDynamic = user.FormatDynamic(dynamicInfo, i)
			continue
		} else {
			formattedDynamic = user.FormatDynamic(dynamicInfo, i)
			break
		}
	}

	return g.Map{
		"top_dynamic":  formattedTopDynamic,
		"last_dynamic": formattedDynamic,
	}
}
