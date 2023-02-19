package main

import (
	"error_bot/api"
	"error_bot/config"
	"error_bot/syncgroup"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func main() {
	config.LoadConfig()
	router.SignUp()

	syncgroup.Wait.Add(1)
	syncgroup.Wait.Wait()
}
