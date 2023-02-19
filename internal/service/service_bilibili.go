package service

import (
	"context"
	"error_bot/internal/dao"
	"error_bot/internal/user"
	"fmt"
	"log"
)

// StartBiliBiliService 开始监听BiliBili的账号状态
func StartBiliBiliService(ctx context.Context) {
	infos, err := dao.GetServiceInfo()
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to get service info, err(%s)", err))
		return
	}

	for _, info := range infos {
		object, ok := user.Objects.Load(info.UserID)
		if ok {
			object.(*user.Class).Groups = append(object.(*user.Class).Groups, &user.Group{
				Id:    info.GroupID,
				AtAll: info.AtAll,
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
		// TODO
		return true
	})
}
