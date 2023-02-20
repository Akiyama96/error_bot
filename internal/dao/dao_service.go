package dao

import (
	"error_bot/internal/table"
	"error_bot/internal/types"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
)

func GetServiceInfo() ([]*types.BilibiliService, error) {
	db := g.DB()
	m := db.Model(table.ServiceInfo)

	infos := make([]*types.BilibiliService, 0)
	m.Where("enable=", 1)
	err := m.Scan(&infos)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func ReplaceServiceInfo(info *types.BilibiliService) error {
	db := g.DB()
	m := db.Model(table.ServiceInfo)

	var replaceData = g.Map{
		"user_group_id":      fmt.Sprintf("%d_%d", info.UserID, info.GroupID),
		"name":               info.Name,
		"user_id":            info.UserID,
		"group_id":           info.GroupID,
		"room_id":            info.RoomID,
		"live_notification":  info.LiveNotification,
		"space_notification": info.SpaceNotification,
		"at_all":             info.AtAll,
		"enable":             info.Enable,
	}

	_, err := m.Replace(replaceData)

	return err
}
