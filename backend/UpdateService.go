/*
 * @Author: tang
 * @Date: 2026-05-23
 * @GitHub: Mr-tang0/CTSystem
 * @Description: 更新服务模块，负责从GitHub获取版本更新信息
 */
package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

//此文件为获取更新信息的接口

type UpdateService struct {
	cachedRelease *GitHubRelease
}

type GitHubRelease struct {
	TagName            string `json:"tag_name"`
	HTMLURL            string `json:"html_url"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Body               string `json:"body"`
	CurrentVersion     string `json:"current_version"`
}

func (u *UpdateService) GetUpdateInfo() (GitHubRelease, error) {
	//根据app.json里的version与name、url获取更新信息
	path := "update.json"
	//读取app.json文件
	appJson, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return GitHubRelease{}, err
	}

	//解析app.json文件
	var appInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		URL     string `json:"url"`
	}
	err = json.Unmarshal(appJson, &appInfo)
	if err != nil {
		return GitHubRelease{}, err
	}

	// 发送HTTP请求获取最新发布信息
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(appInfo.URL)
	if err != nil {
		fmt.Printf("HTTP请求失败: %v\n", err)
		return GitHubRelease{}, err
	}
	defer resp.Body.Close()

	fmt.Printf("GitHub 返回状态码: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		return GitHubRelease{}, errors.New("获取最新发布信息失败")
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal(bodyBytes, &release); err != nil {
		fmt.Printf("JSON 解析失败: %v\n", err)
		return GitHubRelease{}, err
	}
	// 缓存最新发布信息
	u.cachedRelease = &GitHubRelease{
		TagName:            release.TagName,
		HTMLURL:            release.HTMLURL,
		BrowserDownloadURL: release.Assets[0].BrowserDownloadURL,
		Body:               release.Body,
		CurrentVersion:     appInfo.Version,
	}

	println("当前版本:", appInfo.Version)
	println("最新版本:", u.cachedRelease.TagName)
	println("更新内容:", u.cachedRelease.Body)
	println("下载地址:", u.cachedRelease.BrowserDownloadURL)

	if u.cachedRelease.TagName == appInfo.Version {
		return GitHubRelease{}, errors.New("当前版本已经是最新版")
	} else {
		return *u.cachedRelease, nil
	}
}

func (u *UpdateService) GetCachedRelease() GitHubRelease {
	if u.cachedRelease == nil {
		fmt.Println("release info not available")
		return GitHubRelease{}
	}
	fmt.Println("release info available")
	return *u.cachedRelease
}
