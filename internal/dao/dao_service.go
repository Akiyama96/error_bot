package dao

import (
	"error_bot/internal/table"
	"error_bot/internal/types"
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
