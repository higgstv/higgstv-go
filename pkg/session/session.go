package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

// Init 初始化 Session Store
func Init(secretKey string) {
	store = sessions.NewCookieStore([]byte(secretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 天
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// GetSession 取得 session
func GetSession(c *gin.Context) (*sessions.Session, error) {
	return store.Get(c.Request, "higgstv_session")
}

// SetLoggedIn 設定登入狀態
func SetLoggedIn(c *gin.Context, userID, username, email, unclassifiedChannel string) error {
	session, _ := GetSession(c)
	session.Values["logged_in"] = true
	session.Values["uid"] = userID
	session.Values["username"] = username
	session.Values["email"] = email
	if unclassifiedChannel != "" {
		session.Values["unclassified_channel"] = unclassifiedChannel
	}
	return session.Save(c.Request, c.Writer)
}

// IsLoggedIn 檢查是否已登入
func IsLoggedIn(c *gin.Context) bool {
	session, err := GetSession(c)
	if err != nil {
		return false
	}
	loggedIn, ok := session.Values["logged_in"].(bool)
	return ok && loggedIn
}

// GetUserID 取得使用者 ID
func GetUserID(c *gin.Context) string {
	session, _ := GetSession(c)
	if uid, ok := session.Values["uid"].(string); ok {
		return uid
	}
	return ""
}

// GetUsername 取得使用者名稱
func GetUsername(c *gin.Context) string {
	session, _ := GetSession(c)
	if username, ok := session.Values["username"].(string); ok {
		return username
	}
	return ""
}

// GetEmail 取得 Email
func GetEmail(c *gin.Context) string {
	session, _ := GetSession(c)
	if email, ok := session.Values["email"].(string); ok {
		return email
	}
	return ""
}

// GetUnclassifiedChannel 取得未分類頻道 ID
func GetUnclassifiedChannel(c *gin.Context) string {
	session, _ := GetSession(c)
	if ch, ok := session.Values["unclassified_channel"].(string); ok {
		return ch
	}
	return ""
}

// SetUnclassifiedChannel 設定未分類頻道 ID
func SetUnclassifiedChannel(c *gin.Context, channelID string) error {
	session, _ := GetSession(c)
	session.Values["unclassified_channel"] = channelID
	return session.Save(c.Request, c.Writer)
}

// Clear 清除 session
func Clear(c *gin.Context) error {
	session, _ := GetSession(c)
	session.Values = make(map[interface{}]interface{})
	return session.Save(c.Request, c.Writer)
}

