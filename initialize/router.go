package initialize

import (
	"deeplx-pro/translator"
	"github.com/gin-contrib/cors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func authMiddleware(checkToken func(string) bool, hasToken func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasToken() {
			providedTokenInQuery := c.Query("token")
			providedTokenInHeader := c.GetHeader("Authorization")

			// Compatability with the Bearer token format
			if providedTokenInHeader != "" {
				parts := strings.Split(providedTokenInHeader, " ")
				if len(parts) == 2 {
					if parts[0] == "Bearer" || parts[0] == "DeepL-Auth-Key" {
						providedTokenInHeader = parts[1]
					} else {
						providedTokenInHeader = ""
					}
				} else {
					providedTokenInHeader = ""
				}
			}

			if !checkToken(providedTokenInHeader) && !checkToken(providedTokenInQuery) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "Invalid access token",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// InitRouter initializes the Gin router with all necessary routes and middleware
func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())
	// 根路由，返回欢迎信息
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to deeplx-pro")
	})

	// GET方法不支持翻译请求
	r.GET("/translate", func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "GET method not supported for this endpoint. Please use POST.")
	})

	// POST方法处理翻译请求
	r.POST("/translate", authMiddleware(translator.CheckToken, translator.HasToken), func(c *gin.Context) {
		var reqBody struct {
			TransText   string `json:"text"`
			SourceLang  string `json:"source_lang"`
			TargetLang  string `json:"target_lang"`
			TagHandling string `json:"tag_handling"`
		}

		// 绑定JSON请求体
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"error":   "Invalid request body",
				"message": err.Error()})
			return
		}
		//if reqBody.SourceLang == "" {
		//	reqBody.SourceLang = "auto"
		//}
		//// 检查源语言是否为"auto"
		//if reqBody.SourceLang == "auto" || reqBody.SourceLang == "AUTO" {
		//	reqBody.SourceLang = "EN"
		//}
		if reqBody.TargetLang == "" {
			reqBody.TargetLang = "ZH"
		}
		if reqBody.SourceLang == "" {
			reqBody.SourceLang = "auto"
		}
		if reqBody.TagHandling != "" && reqBody.TagHandling != "html" && reqBody.TagHandling != "xml" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"error":   "Invalid request body",
				"message": "Invalid tag_handling value. Allowed values are 'html' and 'xml'.",
			})
			return
		}

		// 调用翻译函数
		result, err := translator.TranslateByDeepLX(reqBody.SourceLang, reqBody.TargetLang, reqBody.TransText, reqBody.TagHandling, translator.GetNextProxy(), translator.GetNextCookie())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Translation failed", "details": err.Error()})
			return
		}
		if result.Code == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{
				"code":         http.StatusOK,
				"id":           result.ID,
				"data":         result.Data,
				"alternatives": result.Alternatives,
				"source_lang":  result.SourceLang,
				"target_lang":  result.TargetLang,
				"method":       result.Method,
			})
		} else {
			c.JSON(result.Code, gin.H{
				"code":    result.Code,
				"message": result.Message,
			})

		}
	})
	// Catch-all route to handle undefined paths
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Path not found",
		})
	})
	return r
}
