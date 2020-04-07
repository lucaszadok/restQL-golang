package httpclient

import (
	"bytes"
	"encoding/json"
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"net/http"
)

var (
	ampersand = []byte("&")
	equal     = []byte("=")
)

func setupRequest(request domain.HttpRequest, req *fasthttp.Request) error {
	uri := fasthttp.URI{DisablePathNormalizing: true}
	uri.SetScheme(request.Schema)
	uri.SetHost(request.Uri)
	uri.SetQueryStringBytes(makeQueryArgs(uri.QueryString(), request))

	uriStr := uri.String()
	req.SetRequestURI(uriStr)

	if request.Method == http.MethodPost || request.Method == http.MethodPut {
		data, err := json.Marshal(request.Body)
		if err != nil {
			//fmt.Printf("failed to marshal request body: %v\n", err)
			return errors.Wrap(err, "failed to marshal request body")
		}

		req.SetBody(data)
	}

	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	req.Header.SetMethod(request.Method)
	return nil
}

func readHeaders(res *fasthttp.Response) domain.Headers {
	h := make(domain.Headers)
	res.Header.VisitAll(func(key, value []byte) {
		h[string(key)] = string(value)
	})

	return h
}

func makeQueryArgs(queryArgs []byte, request domain.HttpRequest) []byte {
	buf := bytes.NewBuffer(queryArgs)

	for key, value := range request.Query {
		switch value := value.(type) {
		case string:
			appendStringParam(buf, key, value)
		case []interface{}:
			appendListParam(buf, key, value)
		}
	}

	return buf.Bytes()
}

func appendListParam(buf *bytes.Buffer, key string, value []interface{}) {
	for _, v := range value {
		s, ok := v.(string)
		if !ok {
			continue
		}

		buf.Write(ampersand)
		buf.WriteString(key)
		buf.Write(equal)
		buf.WriteString(s)
	}
}

func appendStringParam(buf *bytes.Buffer, key string, value string) {
	buf.Write(ampersand)
	buf.WriteString(key)
	buf.Write(equal)
	buf.WriteString(value)
}
