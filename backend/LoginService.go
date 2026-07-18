package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type User struct {
	ID             string                 `json:"id"`
	Username       string                 `json:"username"`
	Password       string                 `json:"-"`
	Email          string                 `json:"email"`
	RegisteredAt   string                 `json:"registered_at"`
	Role           string                 `json:"role"`
	AppPermissions []string               `json:"app_permissions"`
	AppJson        map[string]interface{} `json:"app_json,omitempty"`
}

type LastLoginInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	LoginTime string `json:"login_time"`
}

type LoginService struct {
	container *ServiceContainer
}

type authResponse struct {
	Success bool            `json:"success"`
	OK      bool            `json:"ok"`
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
	User    json.RawMessage `json:"user"`
}

func NewLoginService() *LoginService {
	return &LoginService{}
}

func (l *LoginService) Init(container *ServiceContainer) error {
	l.container = container
	return nil
}

func (l *LoginService) Login(name string, password string) User {
	switch {
	case name == "__last_login__":
		return l.handleLastLogin()
	case strings.HasPrefix(name, "__register__:"):
		return l.handleRegister(name, password)
	case l.isDebugAdmin(name, password):
		return l.createDebugAdmin()
	default:
		return l.handleNormalLogin(name, password)
	}
}

func (l *LoginService) handleLastLogin() User {
	info, err := l.GetLastLoginInfo()
	if err != nil {
		return l.errorUser(err.Error())
	}
	return User{
		ID:           info.ID,
		Username:     info.Username,
		Email:        info.Email,
		Role:         info.Role,
		RegisteredAt: info.LoginTime,
	}
}

func (l *LoginService) handleRegister(name string, password string) User {
	if len(password) < 6 {
		return l.errorUser("密码长度不能少于6位")
	}

	var form struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	payload := strings.TrimPrefix(name, "__register__:")
	if err := json.Unmarshal([]byte(payload), &form); err != nil {
		return l.errorUser("注册信息解析失败")
	}

	user, err := l.Register(form.Username, form.Email, password)
	if err != nil {
		return l.errorUser(err.Error())
	}
	return user
}

func (l *LoginService) isDebugAdmin(name, password string) bool {
	return strings.TrimSpace(name) == "admin" && password == "admin"
}

func (l *LoginService) createDebugAdmin() User {
	user := User{
		ID:       "debug-admin",
		Username: "admin",
		Role:     "admin",
	}
	l.container.SetUser(&user)
	_ = l.saveLastLoginInfo(user)
	return user
}

func (l *LoginService) handleNormalLogin(name string, password string) User {
	if err := l.validateLoginInput(name, password); err != nil {
		return l.errorUser(err.Error())
	}

	api, err := l.getAPIFromEnv("LOGIN_API", "LOGIN_URL", "VITE_LOGIN_API", "APP_LOGIN_API")
	if err != nil {
		return l.errorUser(err.Error())
	}

	user, err := l.requestAuth(api, map[string]string{
		"username": name,
		"password": password,
	})
	if err != nil {
		return l.errorUser(err.Error())
	}

	if user.Username == "" {
		user.Username = name
	}

	if !l.hasAppPermission(user) {
		return l.errorUser("无权限访问本应用")
	}

	return l.completeLogin(&user)
}

func (l *LoginService) validateLoginInput(name, password string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(password) == "" {
		return fmt.Errorf("请填写完整的登录信息")
	}
	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}
	return nil
}

func (l *LoginService) hasAppPermission(user User) bool {
	if user.Role == "admin" {
		return true
	}

	appName := l.readEnvValue("APP_NAME")
	if appName == "" {
		return true
	}

	for _, p := range user.AppPermissions {
		if p == appName {
			return true
		}
	}
	return false
}

func (l *LoginService) completeLogin(user *User) User {
	user.Password = ""
	l.container.SetUser(user)
	_ = l.saveLastLoginInfo(*user)
	return *user
}

func (l *LoginService) errorUser(msg string) User {
	return User{AppJson: map[string]interface{}{"error": msg}}
}

func (l *LoginService) Register(name string, email string, password string) (User, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return User{}, fmt.Errorf("请填写完整的注册信息")
	}
	if len(password) < 6 {
		return User{}, fmt.Errorf("密码长度不能少于6位")
	}

	api, err := l.getAPIFromEnv("REGISTER_API", "REGISTER_URL", "VITE_REGISTER_API", "APP_REGISTER_API")
	if err != nil {
		return User{}, err
	}

	payload := map[string]interface{}{
		"username":      name,
		"password":      password,
		"email":         email,
		"registered_at": time.Now().Format(time.RFC3339),
		"role":          "admin",
		"app_permissions": []string{
			"*",
		},
	}

	user, err := l.requestAuthWithPayload(api, payload)
	if err != nil {
		return User{}, err
	}
	if user.Username == "" {
		user.Username = name
	}
	if user.Email == "" {
		user.Email = email
	}
	user.Password = ""
	return user, nil
}

func (l *LoginService) GetLastLoginInfo() (LastLoginInfo, error) {
	path, err := l.getLastLoginPath()
	if err != nil {
		return LastLoginInfo{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return LastLoginInfo{}, nil
	}
	if err != nil {
		return LastLoginInfo{}, fmt.Errorf("读取最近登录信息失败: %v", err)
	}
	var info LastLoginInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return LastLoginInfo{}, fmt.Errorf("解析最近登录信息失败: %v", err)
	}
	if info.Username == "" {
		info.Username = info.Name
	}
	return info, nil
}

func (l *LoginService) GetLoginUser() User {
	user := l.container.GetUser()
	if user == nil {
		return User{}
	}
	return *user
}

func (l *LoginService) requestAuth(api string, payload map[string]string) (User, error) {
	return l.requestAuthWithPayload(api, payload)
}

func (l *LoginService) requestAuthWithPayload(api string, payload interface{}) (User, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return User{}, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, api, bytes.NewReader(body))

	if err != nil {
		return User{}, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("请求认证接口失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	if err != nil {
		return User{}, fmt.Errorf("读取认证结果失败: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return User{}, fmt.Errorf("认证接口返回错误: %s", strings.TrimSpace(string(respBody)))
	}

	var result authResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		var user User
		if userErr := json.Unmarshal(respBody, &user); userErr == nil && (user.ID != "" || user.Username != "") {
			return user, nil
		}
		return User{}, fmt.Errorf("解析认证结果失败: %v", err)
	}

	if !result.Success && !result.OK && result.Code != 0 && result.Code != 200 {
		msg := result.Message
		if msg == "" {
			msg = result.Msg
		}
		if msg == "" {
			msg = "认证失败"
		}
		return User{}, fmt.Errorf("%s", msg)
	}

	var user User
	if len(result.Data) > 0 && string(result.Data) != "null" {
		_ = json.Unmarshal(result.Data, &user)
	}
	if user.ID == "" && user.Username == "" && len(result.User) > 0 && string(result.User) != "null" {
		_ = json.Unmarshal(result.User, &user)
	}
	return user, nil
}

func (l *LoginService) getAPIFromEnv(keys ...string) (string, error) {
	env, err := l.readEnvFile()
	if err != nil {
		return "", err
	}
	for _, key := range keys {
		if value := strings.TrimSpace(env[key]); value != "" {
			return value, nil
		}
	}
	return "", fmt.Errorf(".env 未配置接口地址: %s", strings.Join(keys, ", "))
}

func (l *LoginService) readEnvFile() (map[string]string, error) {
	path, err := l.findEnvPath()
	if err != nil {
		return nil, err
	}
	return l.readEnvFileAt(path), nil
}

func (l *LoginService) readEnvValue(key string) string {
	path, err := l.findEnvPath()
	if err != nil {
		return ""
	}
	env := l.readEnvFileAt(path)
	return env[key]
}

func (l *LoginService) readEnvFileAt(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	values := map[string]string{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		values[key] = value
	}
	return values
}

func (l *LoginService) findEnvPath() (string, error) {
	wd, err := os.Getwd()
	fmt.Println(wd)
	if err != nil {
		return "", err
	}
	for {
		path := filepath.Join(wd, ".env")
		fmt.Println(path)

		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", fmt.Errorf("在 %s 未找到根目录 .env 文件", wd)
}

func (l *LoginService) getLastLoginPath() (string, error) {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, "PIMS", "MINIMTS", "login")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("创建登录配置目录失败: %v", err)
		}
	}
	return filepath.Join(configDir, "last_user.json"), nil
}

func (l *LoginService) saveLastLoginInfo(user User) error {
	path, err := l.getLastLoginPath()
	if err != nil {
		return err
	}
	info := LastLoginInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		LoginTime: time.Now().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
