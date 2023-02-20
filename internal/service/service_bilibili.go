package service

import (
	"error_bot/internal/dao"
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
		}
	}

	user.Objects.Range(func(key, value any) bool {
		obj := value.(*user.Class)

		ctxLive, cancelLive := context.WithCancel(context.Background())
		keyOfLive := fmt.Sprintf("%d_live", obj.Uid)
		user.Cancels.Store(keyOfLive, cancelLive)

		ctxDynamic, cancelDynamic := context.WithCancel(context.Background())
		keyOfDynamic := fmt.Sprintf("%d_dynamic", obj.Uid)
		user.Cancels.Store(keyOfDynamic, cancelDynamic)

		value.(*user.Class).ListenBiliBiliLiveNotification(ctxLive)
		value.(*user.Class).ListenBiliBiliSpaceNotification(ctxDynamic)
		return true
	})
}
