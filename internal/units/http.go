package units

import (
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"log"
)

const (
	HttpServerErrorCode = 500
	HttpServerOkCode    = 200
)

// HttpResponseError 服务内部错误
func HttpResponseError(req *ghttp.Request, data interface{}) {
	responseWriter := req.Response.Writer

	responseWriter.Status = HttpServerErrorCode

	responseData, err := json.Marshal(data)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to marshal response data, err(%s)", err))
	}

	_, err = req.Response.Writer.Write(responseData)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to response, err(%s)", err))
	}
}

// HttpResponseSuccess 服务内部错误
func HttpResponseSuccess(req *ghttp.Request, data interface{}) {
	responseWriter := req.Response.Writer

	responseWriter.Status = HttpServerOkCode

	responseData, err := json.Marshal(data)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to marshal response data, err(%s)", err))
		HttpResponseError(req, err.Error())
		return
	}

	_, err = req.Response.Writer.Write(responseData)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to response, err(%s)", err))
		HttpResponseError(req, err.Error())
		return
	}
}
