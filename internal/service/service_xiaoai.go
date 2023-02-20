package service

import (
	"context"
	"error_bot/internal/client"
	"fmt"
	"strings"
)

func RequestXiaoAi(message string) string {
	url := fmt.Sprintf("http://81.70.100.130/api/xiaoai.php?msg=%s&n=text", message)
	response := client.Get(context.Background(), url)

	if response == nil {
		return ""
	}

	responseMessage := string(response.ReadAll())
	responseMessage = strings.Replace(responseMessage, "小爱", "era", -1)
	responseMessage = strings.Replace(responseMessage, "小米", "", -1)

	if responseMessage == "" {
		responseMessage = "[CQ:image,file=8bf34b1019c1419558666ddb73a903d8.image,url=https://c2cpicdw.qpic.cn/offpic_new/1131568220//1131568220-1507055459-8BF34B1019C1419558666DDB73A903D8/0?term=3&amp;is_origin=0]"

	}

	return responseMessage
}
