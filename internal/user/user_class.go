package user

import (
	"context"
	"encoding/json"
	"error_bot/internal/bot"
	"error_bot/internal/client"
	"error_bot/internal/types"
	"error_bot/syncgroup"
	"fmt"
	"log"
	"sync"
	"time"
)

type Class struct {
	Name        string   `json:"name"`
	Uid         int      `json:"uid"`
	RoomId      int      `json:"room_id"`
	LiveTime    int64    `json:"live_time"`
	LiveStatus  int      `json:"live_status"`
	MaxHotValue int      `json:"max_hot_value"`
	Flag        bool     `json:"flag"`
	Groups      []*Group `json:"groups"`
}

type Group struct {
	Id    int `json:"id"`
	AtAll int `json:"atAll"`
}

const (
	liveRoomInfoUrl = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id="
	statInfoUrl     = "https://api.bilibili.com/x/relation/stat?vmid="
	spaceInfoUrl    = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space"
	userInfoUrl     = "https://api.bilibili.com/x/web-interface/card?mid="
)

var Objects sync.Map

func (c *Class) ListenBiliBiliLiveNotification(ctx context.Context) {
	syncgroup.Wait.Add(1)
	defer syncgroup.Wait.Done()

	ticker := time.NewTicker(time.Second * 3)

	for {
		select {
		case <-ctx.Done():
			log.Println("INFO: bilibili live service returned, context done.")
			return
		case <-ticker.C:
			url := fmt.Sprintf("%s%d", liveRoomInfoUrl, c.RoomId)
			response := client.Get(ctx, url)
			if response == nil {
				log.Println("ERROR: failed to get live info, nil response.")
				continue
			}

			var liveInfo = &types.LiveRoomInfo{}
			err := json.Unmarshal(response.ReadAll(), liveInfo)
			if err != nil {
				log.Println(fmt.Sprintf("ERROR: failed to get live info, err(%s).", err.Error()))
				continue
			}

			if liveInfo.Msg != "ok" || liveInfo.Message != "ok" {
				log.Println("ERROR: failed to get live info, message not ok.")
				continue
			}

			// 直播状态发生改变
			if c.LiveStatus != liveInfo.Data.LiveStatus {
				// 当前状态为直播中
				if liveInfo.Data.LiveStatus == 1 {
					// 记录直播开始时间用于统计直播时间
					c.LiveTime = time.Now().Unix()
				}

				if !c.Flag {
					c.LiveStatus = liveInfo.Data.LiveStatus
					c.Flag = true
				} else if c.Flag {
					if liveInfo.Data.LiveStatus == 1 { // 开播
						c.sendStartLiveNotification(liveInfo)
						// 更新当前直播状态
						c.LiveStatus = liveInfo.Data.LiveStatus
						// 清零最大人气值
						c.MaxHotValue = 0

					} else if liveInfo.Data.LiveStatus == 0 { // 下播
						c.sendStopLiveNotification()
						// 更新当前直播状态
						c.LiveStatus = liveInfo.Data.LiveStatus

					} else if liveInfo.Data.LiveStatus == 2 { // 轮播
						c.sendPlayOtherVideoNotification(liveInfo)
						// 更新当前直播状态
						c.LiveStatus = liveInfo.Data.LiveStatus

					} else {
						log.Println("ERROR: unknown live status")
					}
				}

			} else { // 直播状态未发生改变
				// 当前状态为直播中
				if liveInfo.Data.LiveStatus == 1 {
					// 记录直播最大人气值
					if liveInfo.Data.Online > c.MaxHotValue {
						c.MaxHotValue = liveInfo.Data.Online
					}
				}
			}

		} // 此处为select结束处
	} // 此处为包含定时器的for循环结束处
}

// 发送开播消息到每个订阅群
func (c *Class) sendStartLiveNotification(info *types.LiveRoomInfo) {
	for _, group := range c.Groups {
		message := fmt.Sprintf("%s开播啦!\n", c.Name) +
			fmt.Sprintf("直播间标题：%s\n", info.Data.Title) +
			fmt.Sprintf("直播间地址：https://live.bilibili.com/%d\n", info.Data.RoomId) +
			fmt.Sprintf("当前人气值%d~\n", info.Data.Online) +
			fmt.Sprintf("[CQ:image,file=%s]\n", info.Data.UserCover)

		if group.AtAll == 1 {
			message += fmt.Sprintf("[CQ:at,qq=all]")
		}

		bot.SendMessage(group.Id, "group", message)
	}
}

// 发送下播消息到每个订阅群
func (c *Class) sendStopLiveNotification() {
	for _, group := range c.Groups {
		message := fmt.Sprintf("%s下播啦!\n", c.Name) +
			fmt.Sprintf(
				"本次直播时间：%d小时%d分%d秒\n~~",
				c.LiveTime/3600,
				(c.LiveTime%3600)/60,
				(c.LiveTime%3600)%60,
			) +
			fmt.Sprintf("本次直播最大人气值:%d!", c.MaxHotValue)
		bot.SendMessage(group.Id, "group", message)
	}
}

// 发送轮播消息到每个订阅群
func (c *Class) sendPlayOtherVideoNotification(info *types.LiveRoomInfo) {
	if c.LiveStatus == 0 {
		for _, group := range c.Groups {
			message := fmt.Sprintf("%s正在轮播中\n", c.Name) +
				fmt.Sprintf("直播间地址：https://live.bilibili.com/%d\n", info.Data.RoomId) +
				fmt.Sprintf("当前人气值%d~\n", info.Data.Online) +
				fmt.Sprintf("[CQ:image,file=%s]", info.Data.Keyframe)
			bot.SendMessage(group.Id, "group", message)
		}

	} else if c.LiveStatus == 1 {
		for _, group := range c.Groups {
			message := fmt.Sprintf("%s下播啦!\n", c.Name) +
				fmt.Sprintf(
					"本次直播时间：%d小时%d分%d秒\n~~",
					c.LiveTime/3600,
					(c.LiveTime%3600)/60,
					(c.LiveTime%3600)%60,
				) +
				fmt.Sprintf("本次直播最大人气值:%d!\n", c.MaxHotValue) +
				fmt.Sprintf("当前直播间正在轮播中：https://live.bilibili.com/%d\n", info.Data.RoomId) +
				fmt.Sprintf("[CQ:image,file=%s]", info.Data.Keyframe)
			bot.SendMessage(group.Id, "group", message)
		}
	}
}

func (c *Class) ListenBiliBiliSpaceNotification(ctx context.Context) {
	syncgroup.Wait.Add(1)
	defer syncgroup.Wait.Done()

	ticker := time.NewTicker(time.Second * 3)

	for {
		select {
		case <-ctx.Done():
			log.Println("INFO: bilibili space service returned, context done.")
			return
		case <-ticker.C:

		}
	}
}

func getUserInfo(ctx context.Context, uid int) *types.BilibiliUserInfo {
	url := fmt.Sprintf("%s%d", userInfoUrl, uid)
	response := client.Get(ctx, url)
	if response == nil {
		log.Println("ERROR: failed to get user info, nil response.")
		return nil
	}

	userInfo := &types.BilibiliUserInfo{}
	err := json.Unmarshal(response.ReadAll(), userInfo)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to get user info, err(%s).", err))
		return nil
	}

	if userInfo.Code != 0 {
		log.Println("ERROR: failed to get user info, code not 0.")
		return nil
	}

	return userInfo
}