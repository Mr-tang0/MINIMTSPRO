package main

import (
	"MINIMTSPRO/backend"
	"embed"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist/*
var assets embed.FS

func main() {
	container := backend.NewServiceContainer()

	container.Register("User", &backend.User{})
	container.Register("SystemService", backend.NewSystemService())
	container.Register("ProjectService", backend.NewProjectService())
	container.Register("UpdateService", &backend.UpdateService{})
	container.Register("AppService", backend.NewAppService())
	container.Register("LoginService", backend.NewLoginService())
	container.Register("MINIMTSService", backend.NewMINIMTSService())
	container.Register("HIKCameraService", backend.NewHIKCamera())

	app := application.New(application.Options{
		Name:        "MINIMTPRO",
		Description: "MINIMTPRO 是一个基于 MINIMTS 的专业版软件，提供更功能丰富的 MINIMTS 控制和管理功能。",
		Services: []application.Service{
			application.NewService(container),
			application.NewService(container.GetUpdateService()),
			application.NewService(container.GetLoginService()),
			application.NewService(container.GetAppService()),
			application.NewService(container.GetMINIMTSService()),
			application.NewService(container.GetHIKCameraService()),
			application.NewService(container.GetProjectService()),
			application.NewService(container.GetSystemService()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	if err := container.InitAll(); err != nil {
		fmt.Println("初始化服务失败:", err)
		return
	}

	//先检查更新
	_, err := container.GetUpdateService().GetUpdateInfo()
	if err == nil {
		container.GetAppService().CallUpdateWindow()
	}

	//打开登录窗口
	container.GetAppService().CallLoginWindow()

	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now().Format(time.RFC1123)
				app.Event.Emit("time", now)
			case <-stopCh:
				return
			}
		}
	}()

	err = app.Run()

	container.GetMINIMTSService().CleanupHardware()
	container.GetMINIMTSService().Stop()
	container.GetHIKCameraService().Stop()
	close(stopCh)

	if err != nil {
		log.Fatal(err)
	}
}
