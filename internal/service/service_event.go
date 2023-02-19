package service

import (
	"context"
	"error_bot/config"
	"error_bot/internal/client"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"strconv"
)

const (
	reqGroupAddUrl = "/set_group_add_request"
)

func AcceptInvite(flag, subType string) {
	url := fmt.Sprintf(
		config.Content.BotServerConfig.Address +
			":" +
			strconv.Itoa(config.Content.BotServerConfig.Port) +
			reqGroupAddUrl)

	formattedData := g.Map{
		"flag":     flag,
		"sub_type": subType,
		"approve":  true,
	}

	client.Post(context.Background(), url, formattedData)
}
