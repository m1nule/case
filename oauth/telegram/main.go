package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/gin-gonic/gin"
)

var (
	token       = os.Getenv("TELEGRAM_TOKEN")
	botName     = os.Getenv("TELEGRAM_NAME")
	callbackURL = os.Getenv("TELEGRAM_CALLBACK_URL")
)

//go:embed template
var static embed.FS

func main() {
	router := gin.Default()
	templ := template.Must(template.New("").ParseFS(static, "template/*.html"))
	router.SetHTMLTemplate(templ)

	router.GET("/telegram", handleIndex)
	router.GET("/oauth2/telegram", handleTelegramOAuthCallback)
	router.Run(":3000")
}

func handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"botName":     botName,
		"callbackURL": callbackURL,
	})
}

func handleTelegramOAuthCallback(c *gin.Context) {
	params := map[string]string{}
	c.ShouldBindQuery(&params)
	b := checkTelegramHash(params)

	if !b {

		c.HTML(http.StatusOK, "index.html", gin.H{
			"msg":         "认证失败: hash不匹配",
			"botName":     botName,
			"callbackURL": callbackURL,
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"botName":     botName,
		"callbackURL": callbackURL,
	})
}

func checkTelegramHash(params map[string]string) bool {
	strs := []string{}
	hash := ""
	for k, v := range params {
		if k == "hash" {
			hash = v
			continue
		}
		strs = append(strs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(strs)
	imploded := ""
	for _, s := range strs {
		if imploded != "" {
			imploded += "\n"
		}
		imploded += s
	}
	sha256hash := sha256.New()
	io.WriteString(sha256hash, token)
	hmachash := hmac.New(sha256.New, sha256hash.Sum(nil))
	io.WriteString(hmachash, imploded)
	ss := hex.EncodeToString(hmachash.Sum(nil))
	return hash == ss
}
