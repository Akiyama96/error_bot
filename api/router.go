package router

import (
	"error_bot/config"
	"error_bot/internal/service"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"log"
)

// SignUp 路由注册
func SignUp() {
	s := g.Server()
	s.SetPort(config.Content.HttpServerConfig.Port)

	a := s.Group("/api")

	//Bot-server 处理事件
	a.POST("/", service.EventService)

	err := s.Start()
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to start api server, err(%s)", err))
	}
}
