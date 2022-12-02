package main

import (
	"github.com/MadAppGang/httplog"
	"github.com/Sterrenhemel/common/logs"
	"github.com/Sterrenhemel/common/tracex"
	"github.com/gin-gonic/gin"

	"moul.io/http2curl"
	"net/http"
)

// httplog.ResponseWriter is not fully compatible with gin.ResponseWriter
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		l := httplog.Logger(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// gin uses wrapped ResponseWriter, we don't want to replace it
			// as gin use custom middleware approach with Next() method and context
			c.Next()
			// set result for ResponseWriter manually
			rwr, _ := rw.(httplog.ResponseWriter)
			rwr.Set(c.Writer.Status(), c.Writer.Size())
		}))
		l.ServeHTTP(c.Writer, c.Request)
	}
}

var anyHandler = func(c *gin.Context) {
	curlCommand, err := http2curl.GetCurlCommand(c.Request)
	ctx := c.Request.Context()
	if err != nil {
		logs.CtxErrorw(ctx, "request error", "err", err)
	}
	logs.CtxInfow(ctx, "curl", "curl", curlCommand.String())
	c.JSON(200, nil)
}

func main() {
	tracex.Init()
	// setup routes
	r := gin.New()
	r.Use(LoggerMiddleware())
	r.NoRoute(anyHandler)

	r.Run(":3333")
}
