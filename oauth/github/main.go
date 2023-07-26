package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/m1nule/case/oauth/github/types"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	clientId      = os.Getenv("GITHUB_CLIENT_ID")
	callbackUrl   = os.Getenv("GITHUB_CALLBACK_URL")
	scope         = os.Getenv("GITHUB_SCOPE")
	clientSecrets = os.Getenv("GITHUB_CLIENT_SECRET")
)

//go:embed template
var static embed.FS

func main() {
	router := gin.Default()
	templ := template.Must(template.New("").ParseFS(static, "template/*.html"))
	router.SetHTMLTemplate(templ)

	router.GET("/github", handleIndex)
	router.GET("/oauth2/github", handleGithubAuthCallback)
	router.Run(":3000")
}

func handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"callbackUrl": callbackUrl,
		"clientId":    clientId,
		"scope":       scope,
	})
}

func handleGithubAuthCallback(c *gin.Context) {
	githubCode, ok := c.GetQuery("code")
	if !ok {
		c.HTML(http.StatusBadRequest, "index.html", gin.H{
			"callbackUrl": callbackUrl,
			"clientId":    clientId,
			"scope":       scope,
			"msg":         "github code is empty",
		})
		return
	}

	token, err := GetToken(fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientId, clientSecrets, githubCode))
	if err != nil {
		logx.Errorf("get token error:%v", err)
		c.HTML(http.StatusBadRequest, "index.html", gin.H{
			"callbackUrl": callbackUrl,
			"clientId":    clientId,
			"scope":       scope,
			"msg":         "get token error",
		})
		return
	}

	ui, err := GetGithubUserInfo("https://api.github.com/user", token.AccessToken)
	if err != nil || ui == nil {
		logx.Errorf("get user info error:%v", err)
		c.HTML(http.StatusBadRequest, "index.html", gin.H{
			"callbackUrl": callbackUrl,
			"clientId":    clientId,
			"scope":       scope,
			"msg":         "get user info error",
		})
		return
	}
	logx.Error(ui)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"callbackUrl": callbackUrl,
		"clientId":    clientId,
		"scope":       scope,
		"user":        ui,
	})
	return
}

func GetToken(url string) (*types.Token, error) {
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")

	// 发送请求并获得响应
	httpClient := http.Client{}
	var res *http.Response
	if res, err = httpClient.Do(req); err != nil {
		return nil, err
	}

	// 将响应体解析为 token，并返回
	var token types.Token
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func GetGithubUserInfo(url, token string) (*types.UserInfo, error) {
	// 形成请求
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	// 发送请求并获得响应
	httpClient := http.Client{}
	var res *http.Response
	if res, err = httpClient.Do(req); err != nil {
		return nil, err
	}

	// 将响应体解析为 token，并返回
	var userinfo types.UserInfo
	if err != nil {
		return nil, err
	}
	if err = json.NewDecoder(res.Body).Decode(&userinfo); err != nil {
		return nil, err
	}
	return &userinfo, nil
}
