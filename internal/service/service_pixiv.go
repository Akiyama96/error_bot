package service

import (
	"context"
	"encoding/json"
	"error_bot/internal/client"
	"error_bot/internal/types"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"log"
)

const picUrl = "https://setu.yuban10703.xyz/setu"

func GetPic(tags []string, r18 int) *types.PicInfo {
	var picInfo = &types.PicInfo{}

	req := g.Map{
		"r18":         r18,
		"num":         1,
		"replace_url": "https://i.pixiv.re",
		"tags":        tags,
	}

	response := client.Post(context.Background(), picUrl, req)
	if response == nil {
		return nil
	}

	err := json.Unmarshal(response.ReadAll(), picInfo)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to get pic, err(%s)", err.Error()))
		return nil
	}

	return picInfo
}
