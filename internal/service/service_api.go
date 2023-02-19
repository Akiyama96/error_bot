package service

import (
	"encoding/json"
	"error_bot/internal/types"
	"error_bot/internal/units"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"io"
	"log"
)

// EventService 监听QQ事件
func EventService(req *ghttp.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to read request body, err(%s)", err))
		units.HttpResponseError(req, g.Map{
			"msg": err.Error(),
		})
		return
	}

	event := &types.Event{}
	err = json.Unmarshal(data, event)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to read request body, err(%s)", err))
		units.HttpResponseError(req, g.Map{
			"msg": err.Error(),
		})
		return
	}

	// TODO:handle event
}
