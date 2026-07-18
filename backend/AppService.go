package backend

import (
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppService 集中管理应用中所有窗口的创建、显示与关闭，
// 避免各业务服务各自持有 application.App 或窗口引用。
type AppService struct {

	// 窗口引用
	mu        sync.RWMutex
	loginWin  *application.WebviewWindow
	mtsWin    *application.WebviewWindow
	cameraWin *application.WebviewWindow
	extWin    *application.WebviewWindow
	updateWin *application.WebviewWindow
}

func NewAppService() *AppService {
	return &AppService{}
}

func (a *AppService) app() *application.App {
	return application.Get()
}

func (a *AppService) openWindow(win *application.WebviewWindow, opts application.WebviewWindowOptions) (*application.WebviewWindow, error) {
	app := a.app()
	if app == nil {
		return nil, fmt.Errorf("application not initialized")
	}

	if win != nil {
		if err := win.Show(); err != nil {
			fmt.Printf("尝试显示隐藏窗口失败: %v, 正在尝试重新创建...\n", err)
		} else {
			win.Focus()
			return win, nil
		}
	}
	return app.Window.NewWithOptions(opts), nil
}

// CallLoginWindow 唤起登录窗口。
func (a *AppService) CallLoginWindow() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	fmt.Println("CallLoginWindow")
	win, err := a.openWindow(a.loginWin, application.WebviewWindowOptions{
		Title:     "MINIMTS Login",
		Width:     850,
		Height:    650,
		Frameless: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})
	if err != nil {
		return err
	}
	a.loginWin = win
	return nil
}

// CallMINIMTSWindow 唤起主窗口。
func (a *AppService) CallMINIMTSWindow() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	fmt.Println("CallMINIMTSWindow")
	win, err := a.openWindow(a.mtsWin, application.WebviewWindowOptions{
		Title:  "MINIMTS",
		Width:  1200,
		Height: 900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/#/mts",
	})
	if err != nil {
		return err
	}
	a.mtsWin = win
	return nil
}

// CallVideoExtensometerWindow 唤起视频引伸计窗口。
func (a *AppService) CallVideoExtensometerWindow() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	win, err := a.openWindow(a.extWin, application.WebviewWindowOptions{
		Title:  "视频引伸计",
		Width:  1200,
		Height: 900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/#/extensometer",
	})
	if err != nil {
		return err
	}
	a.extWin = win
	return nil
}

// CallHIKCameraWindow 唤起相机窗口。
func (a *AppService) CallHIKCameraWindow() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	win, err := a.openWindow(a.cameraWin, application.WebviewWindowOptions{
		Title:  "Camera",
		Width:  1200,
		Height: 900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/#/camera",
	})
	if err != nil {
		return err
	}
	a.cameraWin = win
	return nil
}

// CallUpdateWindow 唤起更新窗口。
func (a *AppService) CallUpdateWindow() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	win, err := a.openWindow(a.updateWin, application.WebviewWindowOptions{
		Title:            "更新可用",
		Width:            550,
		Height:           300,
		DisableResize:    true,
		AlwaysOnTop:      true,
		Frameless:        true,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/#/update",
	})
	if err != nil {
		return err
	}
	a.updateWin = win
	return nil
}

// CloseHIKCameraWindow 关闭相机窗口并释放引用。
func (a *AppService) CloseHIKCameraWindow() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cameraWin != nil {
		a.cameraWin.Close()
		a.cameraWin = nil
	}
}
