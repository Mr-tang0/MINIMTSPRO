package main

import (
	"changeme/backend"
	"embed"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	application.RegisterEvent[string]("time")
}

func main() {
	user := &backend.User{}
	updateService := &backend.UpdateService{}
	loginService := backend.NewLoginService(user)

	projectService := backend.NewProjectService(user)
	systemService := backend.NewSystemService()

	minimtsService := backend.NewMINIMTSService(systemService, projectService, user)
	cameraService := backend.NewHIKCamera()

	app := application.New(application.Options{
		Name:        "MINIMTS_3",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
			application.NewService(updateService),
			application.NewService(loginService),
			application.NewService(minimtsService),
			application.NewService(cameraService),
			application.NewService(projectService),
			application.NewService(systemService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	minimtsService.SetApp(app)

	//默认打开登录窗口
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Login",
		// Width:     650, // 设置窗口宽度
		// Height:    450, // 设置窗口高度
		Frameless: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	//检查更新
	_, err := updateService.GetUpdateInfo()
	if err == nil {

		app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title: "更新可用",
			// 控制窗口大小
			Width:            550,  // 设置窗口宽度
			Height:           300,  // 设置窗口高度
			DisableResize:    true, // 禁止用户拖拽改变窗口大小 (推荐弹窗使用)
			AlwaysOnTop:      true, // 可选：让更新弹窗保持在最顶层
			Frameless:        true,
			BackgroundColour: application.NewRGB(27, 38, 54),
			URL:              "/update",
		})
	} else {
		fmt.Println("No update available:", err)
	}

	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// $env:CGO_ENABLED="1"
// $env:PATH = "C:\msys64\ucrt64\bin;" + $env:PATH
