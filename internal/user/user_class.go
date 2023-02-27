package user

import (
	"context"
	"encoding/json"
	"error_bot/config"
	"error_bot/internal/bot"
	"error_bot/internal/client"
	"error_bot/internal/types"
	"error_bot/syncgroup"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"log"
	"sync"
	"time"
)

type Class struct {
	Name             string   `json:"name"`
	Uid              int      `json:"uid"`
	RoomId           int      `json:"room_id"`
	LiveTime         int64    `json:"live_time"`
	LiveStatus       int      `json:"live_status"`
	MaxHotValue      int      `json:"max_hot_value"`
	LastDynamicId    string   `json:"last_dynamic_id"`
	LastTopDynamicId string   `json:"last_top_dynamic_id"`
	IsExistingTop    bool     `json:"is_existing_top"`
	LiveFlag         bool     `json:"live_flag"`
	DynamicFlag      bool     `json:"dynamic_flag"`
	Groups           sync.Map `json:"groups"`
	ReadDynamicList  sync.Map `json:"read_dynamic_list"`
}

type Group struct {
	Id                int `json:"id"`
	AtAll             int `json:"atAll"`
	LiveNotification  int `json:"live_notification"`
	SpaceNotification int `json:"space_notification"`
}

const (
	liveRoomInfoUrl = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id="
	//statInfoUrl     = "https://api.bilibili.com/x/relation/stat?vmid="
	spaceInfoUrl = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space"
	userInfoUrl  = "https://api.bilibili.com/x/web-interface/card?mid="
)

var (
	Objects sync.Map
	Cancels sync.Map
)

func (c *Class) StartNewService() {
	ctxLive, cancelLive := context.WithCancel(context.Background())
	keyOfLive := fmt.Sprintf("%d_live", c.Uid)
	Cancels.Store(keyOfLive, cancelLive)

	ctxDynamic, cancelDynamic := context.WithCancel(context.Background())
	keyOfDynamic := fmt.Sprintf("%d_dynamic", c.Uid)
	Cancels.Store(keyOfDynamic, cancelDynamic)

	go c.ListenBiliBiliLiveNotification(ctxLive)
	go c.ListenBiliBiliSpaceNotification(ctxDynamic)
}

func (c *Class) ListenBiliBiliLiveNotification(ctx context.Context) {
	syncgroup.Wait.Add(1)
	defer syncgroup.Wait.Done()

	bot.SendMessage(config.Content.BotServerConfig.QQ, "private", fmt.Sprintf("INFO: bilibili live service started.\nUid:%d", c.Uid))

	ticker := time.NewTicker(time.Second * 3)

	for {
		select {
		case <-ctx.Done():
			log.Println("INFO: bilibili live service returned, context done.")
			bot.SendMessage(config.Content.BotServerConfig.QQ, "private", fmt.Sprintf("INFO: bilibili live service returned, context done.\nUid:%d", c.Uid))
			return
		case <-ticker.C:
			liveInfo := GetLiveRoomInfo(ctx, c.RoomId)
			if liveInfo == nil {
				continue
			}

			// 直播状态发生改变
			if c.LiveStatus != liveInfo.Data.LiveStatus {
				// 当前状态为直播中
				if liveInfo.Data.LiveStatus == 1 {
					// 记录直播开始时间用于统计直播时间
					c.LiveTime = time.Now().Unix()
				}

				if !c.LiveFlag {
					c.LiveStatus = liveInfo.Data.LiveStatus
					c.LiveFlag = true
				} else if c.LiveFlag {
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
	c.Groups.Range(func(key, value any) bool {
		group := value.(*Group)

		if group.LiveNotification != 1 {
			return true
		}

		message := fmt.Sprintf("%s开播啦!\n", c.Name) +
			fmt.Sprintf("直播间标题：%s\n", info.Data.Title) +
			fmt.Sprintf("直播间地址：https://live.bilibili.com/%d\n", info.Data.RoomId) +
			fmt.Sprintf("当前人气值%d~\n", info.Data.Online) +
			fmt.Sprintf("[CQ:image,file=%s]\n", info.Data.UserCover)

		if group.AtAll == 1 {
			message += fmt.Sprintf("[CQ:at,qq=all]")
		}

		bot.SendMessage(group.Id, "group", message)

		return true
	})
}

// 发送下播消息到每个订阅群
func (c *Class) sendStopLiveNotification() {
	c.Groups.Range(func(key, value any) bool {
		group := value.(*Group)

		if group.LiveNotification != 1 {
			return true
		}

		message := fmt.Sprintf("%s下播啦!\n", c.Name) +
			fmt.Sprintf(
				"本次直播时间：%d小时%d分%d秒\n~~",
				(time.Now().Unix()-c.LiveTime)/(3600),
				((time.Now().Unix()-c.LiveTime)%3600)/60,
				((time.Now().Unix()-c.LiveTime)%3600)%60,
			) +
			fmt.Sprintf("本次直播最大人气值:%d!", c.MaxHotValue)

		bot.SendMessage(group.Id, "group", message)

		return true
	})
}

// 发送轮播消息到每个订阅群
func (c *Class) sendPlayOtherVideoNotification(info *types.LiveRoomInfo) {
	if c.LiveStatus == 0 {
		c.Groups.Range(func(key, value any) bool {
			group := value.(*Group)

			if group.LiveNotification != 1 {
				return true
			}

			message := fmt.Sprintf("%s正在轮播中\n", c.Name) +
				fmt.Sprintf("直播间地址：https://live.bilibili.com/%d\n", info.Data.RoomId) +
				fmt.Sprintf("当前人气值%d~\n", info.Data.Online) +
				fmt.Sprintf("[CQ:image,file=%s]", info.Data.Keyframe)

			bot.SendMessage(group.Id, "group", message)

			return true
		})

	} else if c.LiveStatus == 1 {
		c.Groups.Range(func(key, value any) bool {
			group := value.(*Group)

			if group.LiveNotification != 1 {
				return true
			}

			message := fmt.Sprintf("%s下播啦!\n", c.Name) +
				fmt.Sprintf(
					"本次直播时间：%d小时%d分%d秒\n~~",
					(time.Now().Unix()-c.LiveTime)/3600,
					((time.Now().Unix()-c.LiveTime)%3600)/60,
					((time.Now().Unix()-c.LiveTime)%3600)%60,
				) +
				fmt.Sprintf("本次直播最大人气值:%d!\n", c.MaxHotValue) +
				fmt.Sprintf("当前直播间正在轮播中：https://live.bilibili.com/%d\n", info.Data.RoomId) +
				fmt.Sprintf("[CQ:image,file=%s]", info.Data.Keyframe)
			bot.SendMessage(group.Id, "group", message)

			return true
		})

	}
}

func (c *Class) ListenBiliBiliSpaceNotification(ctx context.Context) {
	syncgroup.Wait.Add(1)
	defer syncgroup.Wait.Done()

	bot.SendMessage(config.Content.BotServerConfig.QQ, "private", fmt.Sprintf("INFO: bilibili space service started.\nUid:%d", c.Uid))

	ticker := time.NewTicker(time.Second * 3)

	for {
		select {
		case <-ctx.Done():
			log.Println("INFO: bilibili space service returned, context done.")
			bot.SendMessage(config.Content.BotServerConfig.QQ, "private", fmt.Sprintf("INFO: bilibili space service returned, context done.\nUid:%d", c.Uid))
			return
		case <-ticker.C:
			dynamicInfo := GetDynamicInfo(ctx, c.Uid)
			if dynamicInfo == nil {
				continue
			}

			for i, dynamicItem := range dynamicInfo.Data.Items {
				if !c.DynamicFlag {
					if dynamicItem.Modules.ModuleTag.Text == "置顶" {
						c.LastTopDynamicId = dynamicItem.IdStr
						c.IsExistingTop = true
					} else {
						c.LastDynamicId = dynamicItem.IdStr
						c.DynamicFlag = true
					}

				} else if c.DynamicFlag {
					if dynamicItem.Modules.ModuleTag.Text == "置顶" { //置顶动态
						if c.LastTopDynamicId != dynamicItem.IdStr {
							c.handleTopDynamic(dynamicInfo, i)
							c.LastTopDynamicId = dynamicItem.IdStr
						}
						c.IsExistingTop = true

					} else { // 普通动态
						if c.isNewDynamic(dynamicItem.IdStr) {
							c.handleDynamic(dynamicInfo, i)
							c.LastDynamicId = dynamicItem.IdStr
						}

						// 如果第一条是普通动态，说明没有置顶动态
						if i == 0 {
							c.IsExistingTop = false
							c.LastTopDynamicId = ""
						}
					}
				}

				// 仅处理第一条动态
				if c.IsExistingTop && i == 0 {
					continue
				} else {
					for i, item := range dynamicInfo.Data.Items {
						c.ReadDynamicList.Store(i, item.IdStr)
					}

					break
				}
			}
		}
	}
}

func (c *Class) isNewDynamic(id string) bool {
	is := true
	c.ReadDynamicList.Range(func(key, value any) bool {
		if gconv.String(value) == id {
			is = false
			return false
		}
		return true
	})

	return is
}

func (c *Class) handleDynamic(dynamicInfo *types.SpaceInfo, index int) {
	message := FormatDynamic(dynamicInfo, index)

	c.Groups.Range(func(key, value any) bool {
		group := value.(*Group)

		if group.SpaceNotification != 1 {
			return true
		}

		bot.SendMessage(group.Id, "group", message)

		return true
	})
}

func (c *Class) handleTopDynamic(dynamicInfo *types.SpaceInfo, index int) {
	message := fmt.Sprintf("%s更新了置顶动态\n\n:", c.Name)
	message += FormatDynamic(dynamicInfo, index)

	c.Groups.Range(func(key, value any) bool {
		group := value.(*Group)

		if group.SpaceNotification != 1 {
			return true
		}

		bot.SendMessage(group.Id, "group", message)

		return true
	})
}

func FormatDynamic(dynamicInfo *types.SpaceInfo, index int) string {
	var (
		message string
		dynamic = dynamicInfo.Data.Items[index]
	)

	switch dynamic.Type {
	case "DYNAMIC_TYPE_WORD":
		message = fmt.Sprintf("%s发布了新动态!\n", dynamic.Modules.ModuleAuthor.Name)

		if dynamic.Modules.ModuleDynamic.Topic != nil {
			message += fmt.Sprintf("#%s\n", dynamic.Modules.ModuleDynamic.Topic.Name) +
				fmt.Sprintf("Url:%s\n", dynamic.Modules.ModuleDynamic.Topic.JumpUrl)
		}

		if dynamic.Modules.ModuleDynamic.Desc != nil {
			message += fmt.Sprintf("动态地址:https:%s\n", dynamic.Modules.ModuleAuthor.JumpUrl) +
				fmt.Sprintf("\n%s\n", dynamic.Modules.ModuleDynamic.Desc.Text)
		}

	case "DYNAMIC_TYPE_AV":
		message = fmt.Sprintf("%s发布了新视频!\n", dynamic.Modules.ModuleAuthor.Name)
		if dynamic.Modules.ModuleDynamic.Desc != nil {
			message += fmt.Sprintf("%s\n", dynamic.Modules.ModuleDynamic.Desc.Text)
		}

		if dynamic.Modules.ModuleDynamic.Major.Article.Title == "" && dynamic.Modules.ModuleDynamic.Major.Article.Id == 0 {
			message += fmt.Sprintf("\n%s\n", dynamic.Modules.ModuleDynamic.Major.Archive.Title) +
				fmt.Sprintf("视频地址:https:%s\n", dynamic.Modules.ModuleDynamic.Major.Archive.JumpUrl) +
				fmt.Sprintf("[CQ:image,file=%s]", dynamic.Modules.ModuleDynamic.Major.Archive.Cover)
		} else {
			if dynamic.Modules.ModuleDynamic.Major != nil {
				message += fmt.Sprintf("\n%s\n", dynamic.Modules.ModuleDynamic.Major.Article.Title) +
					fmt.Sprintf("视频地址:https:%s\n", dynamic.Modules.ModuleDynamic.Major.Article.JumpUrl) +
					fmt.Sprintf("[CQ:image,file=%s]", dynamic.Modules.ModuleDynamic.Major.Article.Cover)
			}
		}

	case "DYNAMIC_TYPE_DRAW":
		message = fmt.Sprintf("%s发布了新动态!\n", dynamic.Modules.ModuleAuthor.Name)

		if dynamic.Modules.ModuleDynamic.Topic != nil {
			message += fmt.Sprintf("#%s\n", dynamic.Modules.ModuleDynamic.Topic.Name) +
				fmt.Sprintf("Url:%s\n", dynamic.Modules.ModuleDynamic.Topic.JumpUrl)
		}

		if dynamic.Modules.ModuleDynamic.Desc != nil {
			message += fmt.Sprintf("动态地址:https:%s", dynamic.Modules.ModuleAuthor.JumpUrl) +
				fmt.Sprintf("\n%s\n", dynamic.Modules.ModuleDynamic.Desc.Text)
		}

		if dynamic.Modules.ModuleDynamic.Major != nil {
			if dynamic.Modules.ModuleDynamic.Major.Type == "MAJOR_TYPE_DRAW" {
				for _, draw := range dynamic.Modules.ModuleDynamic.Major.Draw.Items {
					message += fmt.Sprintf("[CQ:image,file=%s]\n", draw.Src)
				}
			}
		}

	case "DYNAMIC_TYPE_FORWARD":
		message = fmt.Sprintf("%s转发了动态!\n", dynamic.Modules.ModuleAuthor.Name)

		if dynamic.Modules.ModuleDynamic.Topic != nil {
			message += fmt.Sprintf("#%s\n", dynamic.Modules.ModuleDynamic.Topic.Name) +
				fmt.Sprintf("Url:%s\n", dynamic.Modules.ModuleDynamic.Topic.JumpUrl)
		}

		if dynamic.Modules.ModuleDynamic.Desc != nil {
			message += fmt.Sprintf("动态地址:https:%s", dynamic.Modules.ModuleAuthor.JumpUrl) +
				fmt.Sprintf("\n%s\n", dynamic.Modules.ModuleDynamic.Desc.Text)
		}

		// 递归处理转发的动态
		if dynamic.Orig != nil {
			item := make([]*types.SpaceInfoItem, 0)
			item = append(item, dynamic.Orig)

			message += "\n" + FormatDynamic(&types.SpaceInfo{
				Data: struct {
					HasMore        bool                   `json:"has_more"`
					Items          []*types.SpaceInfoItem `json:"items"`
					Offset         string                 `json:"offset"`
					UpdateBaseline string                 `json:"update_baseline"`
					UpdateNum      int                    `json:"update_num"`
				}{
					Items: item,
				},
			}, 0)
		}
	}

	return message
}

func GetLiveRoomInfo(ctx context.Context, roomId int) *types.LiveRoomInfo {
	url := fmt.Sprintf("%s%d", liveRoomInfoUrl, roomId)
	response := client.Get(ctx, url)
	if response == nil {
		log.Println("ERROR: failed to get live info, nil response.")
		return nil
	}

	var liveInfo = &types.LiveRoomInfo{}
	err := json.Unmarshal(response.ReadAll(), liveInfo)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to get live info, err(%s).", err.Error()))
		return nil
	}

	if liveInfo.Msg != "ok" || liveInfo.Message != "ok" {
		log.Println("ERROR: failed to get live info, message not ok.")
		return nil
	}

	return liveInfo
}

func GetDynamicInfo(ctx context.Context, uid int) *types.SpaceInfo {
	url := fmt.Sprintf("%s?offset=&host_mid=%d&timezone_offset=-480", spaceInfoUrl, uid)
	response := client.Get(ctx, url)
	if response == nil {
		log.Println("ERROR: failed to get dynamic info, nil response.")
		return nil
	}

	var dynamicInfo = &types.SpaceInfo{}
	err := json.Unmarshal(response.ReadAll(), dynamicInfo)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to get dynamic info, err(%s).", err.Error()))
		return nil
	}

	if dynamicInfo.Code != 0 || dynamicInfo.Message != "0" {
		log.Println("ERROR: failed to get dynamic info, message not ok.")
		return nil
	}

	return dynamicInfo
}

func GetUserInfo(ctx context.Context, uid int) *types.BilibiliUserInfo {
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
