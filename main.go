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
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"io"
	"moul.io/http2curl"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

type PostSchema struct {
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType,omitempty"`
	References []struct {
		Name    string `json:"name,omitempty"`
		Subject string `json:"subject,omitempty"`
		Version int    `json:"version,omitempty"`
	} `json:"references,omitempty"`
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

func getSchema(ctx context.Context, subject string) (string, error) {
	resp, err := otelhttp.Get(ctx,
		fmt.Sprintf("http://redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081/subjects/%s/versions/latest", subject),
	)
	if err != nil {
		logs.CtxErrorw(ctx, "http DefaultClient Get", "err", err)
		return "", err
	}
	respSchema, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.CtxErrorw(ctx, "http io.ReadAll", "err", err)
		return "", err
	}
	schemaStr := gjson.GetBytes(respSchema, "schema").String()
	logs.CtxInfow(ctx, "respSchema", "schema", schemaStr)
	//schemaStr := "{\"type\":\"record\",\"name\":\"user\",\"namespace\":\"default.v\",\"fields\":[{\"name\":\"biz_id\",\"type\":{\"type\":\"long\",\"connect.parameters\":{\"tidb_type\":\"BIGINT UNSIGNED\"}}},{\"name\":\"user_id\",\"type\":{\"type\":\"long\",\"connect.parameters\":{\"tidb_type\":\"BIGINT UNSIGNED\"}}},{\"name\":\"user_name\",\"type\":{\"type\":\"string\",\"connect.parameters\":{\"tidb_type\":\"TEXT\"}}},{\"default\":\"\",\"name\":\"avatar\",\"type\":{\"type\":\"string\",\"connect.parameters\":{\"tidb_type\":\"TEXT\"}}},{\"default\":\"\",\"name\":\"description\",\"type\":{\"type\":\"string\",\"connect.parameters\":{\"tidb_type\":\"TEXT\"}}},{\"default\":0,\"name\":\"gender\",\"type\":{\"type\":\"int\",\"connect.parameters\":{\"tidb_type\":\"INT\"}}},{\"default\":\"\",\"name\":\"phone_number\",\"type\":{\"type\":\"string\",\"connect.parameters\":{\"tidb_type\":\"TEXT\"}}},{\"default\":\"CURRENT_TIMESTAMP\",\"name\":\"ctime\",\"type\":{\"type\":\"string\",\"connect.parameters\":{\"tidb_type\":\"TIMESTAMP\"}}},{\"default\":\"CURRENT_TIMESTAMP\",\"name\":\"mtime\",\"type\":{\"type\":\"string\",\"connect.parameters\":{\"tidb_type\":\"TIMESTAMP\"}}},{\"default\":\"CURRENT_TIMESTAMP\",\"name\":\"btime\",\"type\":{\"type\":\"string\",\"connect.parameters\":{\"tidb_type\":\"TIMESTAMP\"}}}]}"
	return schemaStr, nil
}

var subjectsHandler = func(c *gin.Context) {
	curlCommand, err := http2curl.GetCurlCommand(c.Request)
	//c.Request.Host = "redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081"
	c.Request.URL.Host = "redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081"
	c.Request.URL.Scheme = "http"

	ctx := c.Request.Context()
	logs.CtxInfow(ctx, "curl", "curl", curlCommand.String())

	//remote, err := url.Parse("http://redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081")
	//if err != nil {
	//	logs.CtxErrorw(ctx, "url.Parse", "err", err)
	//	return
	//}

	// /subjects/tidb_v_user-value?normalize=false&deleted=true
	// /schemas/ids/1?fetchMaxId=false&subject=tidb_v_user-value
	// /subjects/{subjects}/versions/latest

	subject := c.Param("subjects")

	//schemaBytes, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(string(respSchema))))
	//if err != nil {
	//	logs.CtxErrorw(ctx, "http io.ReadAll", "err", err)
	//	return
	//}
	//logs.CtxInfow(ctx, "get schema", "schema", string(schemaBytes), "subject", subject)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logs.CtxErrorw(ctx, "http io.ReadAll", "err", err)
		return
	}

	var postSchema PostSchema
	err = json.Unmarshal([]byte(body), &postSchema)
	if err != nil {
		logs.CtxErrorw(ctx, "json.Unmarshal", "err", err)
		return
	}

	postSchema.Schema, err = getSchema(ctx, subject)
	if err != nil {
		logs.CtxErrorw(ctx, "getSchema", "err", err)
		return
	}

	newBody, err := json.Marshal(postSchema)
	if err != nil {
		logs.CtxErrorw(ctx, "json.Marshal", "err", err)
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
	c.Request.Header.Set("Content-Length", strconv.Itoa(len(newBody)))

	url1, err := url.ParseRequestURI(c.Request.RequestURI)
	if err != nil {
		logs.CtxErrorw(ctx, "url.ParseRequestURI", "err", err)
		return
	}
	url1.Scheme = "http"
	url1.Host = "redpanda-cluster-0.redpanda-cluster.redpanda-cluster-default.svc.cluster.local:8081"

	req := &http.Request{
		Header: c.Request.Header,
		Host:   c.Request.Host,
		URL:    url1,
		Body:   io.NopCloser(bytes.NewBuffer(newBody)),
		Method: http.MethodPost,
	}
	curlCommand, err = http2curl.GetCurlCommand(req)
	logs.CtxInfow(ctx, "curl", "curl", curlCommand.String())
	rawResp, err := otelhttp.DefaultClient.Do(req)
	if err != nil {
		logs.CtxErrorw(ctx, "json.Marshal", "err", err)
		return
	}
	respBody, err := io.ReadAll(rawResp.Body)
	if err != nil {
		logs.CtxErrorw(ctx, "io.ReadAll", "err", err)
		return
	}
	logs.CtxInfow(ctx, "post resp", "respBody", string(respBody))
	_, err = c.Writer.Write(respBody)
	if err != nil {
		logs.CtxErrorw(ctx, "io.Write", "err", err)
		return
	}
	//proxy := httputil.NewSingleHostReverseProxy(remote)
	//Define the director func
	//This is a good place to log, for example
	//proxy.Director = func(req *http.Request) {
	//	req.Header = c.Request.Header
	//	req.Host = remote.Host
	//	req.URL.Scheme = remote.Scheme
	//	req.URL.Host = remote.Host
	//	req.URL.Path = c.Request.URL.Path
	//}

	//proxy.ServeHTTP(c.Writer, c.Request)
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
	r.POST("/subjects/:subjects", subjectsHandler)
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
