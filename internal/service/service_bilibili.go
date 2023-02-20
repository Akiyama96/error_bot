package service

import (
	"error_bot/internal/dao"
	"error_bot/internal/types"
	"error_bot/internal/user"
	"fmt"
	"golang.org/x/net/context"
	"log"
)

// StartBiliBiliService 开始监听BiliBili的账号状态
func StartBiliBiliService() {
	infos, err := dao.GetServiceInfo()
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to get service info, err(%s)", err))
		return
	}

	for _, info := range infos {
		CreateObject(info)
	}
}

func CreateObject(info *types.BilibiliService) {
	object, ok := user.Objects.Load(info.UserID)
	if ok {
		object.(*user.Class).Groups = append(object.(*user.Class).Groups, &user.Group{
			Id:                info.GroupID,
			AtAll:             info.AtAll,
			LiveNotification:  info.LiveNotification,
			SpaceNotification: info.SpaceNotification,
		})
	}

	if !ok {
		newObject := &user.Class{
			Name:   info.Name,
			Uid:    info.UserID,
			RoomId: info.RoomID,
			Groups: make([]*user.Group, 0),
		}

		newObject.Groups = append(newObject.Groups, &user.Group{
			Id:    info.GroupID,
			AtAll: info.AtAll,
		})

		object = newObject
	}

	keyOfLive := fmt.Sprintf("%d_live", object.(*user.Class).Uid)
	if v, ok := user.Cancels.Load(keyOfLive); ok {
		v.(context.CancelFunc)()
	}

	keyOfDynamic := fmt.Sprintf("%d_dynamic", object.(*user.Class).Uid)
	if v, ok := user.Cancels.Load(keyOfDynamic); ok {
		v.(context.CancelFunc)()
	}

	object.(*user.Class).StartNewService()
}
