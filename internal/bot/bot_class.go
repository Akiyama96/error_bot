package bot

import (
	"context"
	"error_bot/config"
	"error_bot/internal/client"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"strconv"
	"sync"
)

type class struct {
	Id          int    `json:"id"`
	MessageType string `json:"message_type"`
	AutoEscape  bool   `json:"auto_escape"`
}

const (
	sendMsgUrl = "/send_msg"
)

// ObjPool 对象池，用于减少分配开销
var objPool = sync.Pool{
	New: func() any {
		return new(class)
	},
}

func (c *class) sendMessage(message string) {
	url := fmt.Sprintf(
		config.Content.BotServerConfig.Address +
			":" +
			strconv.Itoa(config.Content.BotServerConfig.Port) +
			sendMsgUrl)

	formattedData := g.Map{
		"message_type": c.MessageType,
		"message":      message,
		"auto_escape":  false,
	}

	switch c.MessageType {
	case "group":
		formattedData["group_id"] = c.Id
	case "private":
		formattedData["user_id"] = c.Id
	}

	client.Post(context.Background(), url, formattedData)
}

// SendMessage 发送消息
func SendMessage(id int, messageType, message string) {
	obj := objPool.Get()
	objEntry := obj.(*class)
	objEntry.Id = id
	objEntry.MessageType = messageType
	objEntry.sendMessage(message)
}
