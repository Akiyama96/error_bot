package service

import (
	"context"
	"error_bot/config"
	"error_bot/internal/bot"
	"error_bot/internal/client"
	"error_bot/internal/types"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
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
