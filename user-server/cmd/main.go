package cmd

import "github.com/gangantongxue/knowsync/user-server/internal/app"

func main() {
	// 初始化应用
	err := app.NewApp()
	if err != nil {
		panic(err)
	}
}
