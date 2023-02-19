package client

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"log"
	"time"
)

func Post(ctx context.Context, url string, data interface{}) *gclient.Response {
	c := g.Client()
	c.SetRetry(3, time.Second*3)
	c.SetTimeout(time.Second * 2)

	res, err := c.Post(ctx, url, data)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to post url(%s), err(%s)", url, err))
		return nil
	}

	return res
}

func Get(ctx context.Context, url string) *gclient.Response {
	c := g.Client()
	c.SetRetry(3, time.Second*3)
	c.SetTimeout(time.Second * 2)

	res, err := c.Get(ctx, url)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR: failed to post url(%s), err(%s)", url, err))
		return nil
	}

	if res.StatusCode != 200 {
		return nil
	}

	return res
}
