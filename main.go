package main

import (
	"context"
	"fmt"
	"github.com/Sterrenhemel/common/logs"
	"github.com/Sterrenhemel/common/tracex"
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	"moul.io/http2curl"
	"net/http"
)

var anyHandler = func(c *gin.Context) {
	another := c.Request.Clone(c.Request.Context())

	curlCommand, err := http2curl.GetCurlCommand(another)
	another.Host = "redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081"
	another.URL = c.Request.URL
	another.URL.Host = another.Host
	logs.CtxInfow(c.Request.Context(), "curl", "curl", curlCommand.String())

	remote, err := url.Parse("http://redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	//Define the director func
	//This is a good place to log, for example
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Request.URL.Path
	}

	proxy.ServeHTTP(c.Writer, c.Request)

}

var BuildTime = ""

func main() {
	tracex.Init()
	ctx := context.Background()
	defer tracex.ShutDown(ctx)

	logs.CtxInfo(ctx, "BuildTime:%s", BuildTime)
	// setup routes
	r := gin.New()
	r.NoRoute(anyHandler)

	portInt := int64(80)
	port := os.Getenv("PORT")
	if port != "" {
		var err error
		portInt, err = strconv.ParseInt(port, 10, 64)
		if err != nil {
			portInt = 80
		}
	}

	r.Run(fmt.Sprintf(":%d", portInt))
}
