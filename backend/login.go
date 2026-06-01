package backend

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type User struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Password   string                 `json:"password"`
	Email      string                 `json:"email"`
	Role       string                 `json:"role"`
	CreatedAt  string                 `json:"created_at"`
	AppJson    map[string]interface{} `json:"app_json"`
	StaticPath string                 `json:"static_path"`
}

type LoginService struct {
	user *User
}

func NewLoginService(user *User) *LoginService {
	return &LoginService{
		user: user,
	}
}

func (l *LoginService) Login(name string, password string) User {
	if l.user == nil {
		l.user = &User{}
	}
	l.user.ID = "1"
	l.user.Name = name
	l.user.Password = password
	l.user.Email = ""
	l.user.Role = "admin"
	l.user.CreatedAt = "2023-08-01"
	l.user.AppJson = nil
	l.user.StaticPath = ""
	return *l.user
}

// 唤起MINIMTS设备窗口
func (l *LoginService) CallMINIMTSWindow() error {
	app := application.Get()
	if app == nil {
		return fmt.Errorf("application not initialized")
	}
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "MINIMTS",
		Width:  1200, // 设置窗口宽度
		Height: 900,  // 设置窗口高度
		// Frameless: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/mts",
	})
	return nil
}
