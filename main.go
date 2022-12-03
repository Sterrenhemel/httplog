package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sterrenhemel/common/env"
	"github.com/Sterrenhemel/common/logs"
	"github.com/Sterrenhemel/common/tracex"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"io"
	"moul.io/http2curl"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type PostSchema struct {
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType"`
	References []struct {
		Name    string `json:"name"`
		Subject string `json:"subject"`
		Version int    `json:"version"`
	} `json:"references"`
}

var anyHandler = func(c *gin.Context) {

	curlCommand, err := http2curl.GetCurlCommand(c.Request)
	//c.Request.Host = "redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081"
	logs.CtxInfow(c.Request.Context(), "curl", "curl", curlCommand.String())

	remote, err := url.Parse("http://redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081")
	if err != nil {
		c.Error(err)
		return
	}
	if c.Request.Method == http.MethodPost &&
		strings.HasPrefix(c.Request.URL.Path, "/subjects/") &&
		strings.HasSuffix(c.Request.URL.Path, "-value") {
		// /schemas/ids/1?fetchMaxId=false&subject=tidb_v_user-value
		// /subjects/{subjects}/versions/latest

		subject := c.Query("subject")

		resp, err := http.DefaultClient.Get(fmt.Sprintf("/subjects/%s/versions/latest", subject))
		if err != nil {
			c.Error(err)
			return
		}
		respSchema, err := io.ReadAll(resp.Body)
		if err != nil {
			c.Error(err)
			return
		}

		logs.CtxInfow(c.Request.Context(), "get schema", "schema", respSchema, "subject", subject)
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Error(err)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var postSchema PostSchema
		err = json.Unmarshal([]byte(body), postSchema)
		if err != nil {
			c.Error(err)
			return
		}

		postSchema.Schema = string(respSchema)
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
	r.Use(otelgin.Middleware(env.ServiceName()))

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
