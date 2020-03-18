package httpclient

import (
	"context"
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/b2wdigital/restQL-golang/internal/plataform/logger"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"time"
)

var errExecuteRequestTimeout = errors.New("request timed out")

type HttpClient struct {
	client *fasthttp.Client
	log    *logger.Logger
}

func New(log *logger.Logger) HttpClient {
	c := &fasthttp.Client{
		NoDefaultUserAgentHeader: false,
		ReadTimeout:              3 * time.Second,
		WriteTimeout:             1 * time.Second,
	}

	return HttpClient{client: c, log: log}
}

func (hc HttpClient) Do(ctx context.Context, request domain.HttpRequest) (domain.HttpResponse, error) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	setupRequest(request, req)

	err := hc.executeWithContext(ctx, req, res)
	switch {
	case err == errExecuteRequestTimeout:
		hc.log.Debug("request execution did not complete on time", "request", request)
		return domain.HttpResponse{}, errors.Wrap(err, "request execution failed")
	case err != nil:
		return domain.HttpResponse{}, errors.Wrap(err, "request execution failed")
	}

	response, err := makeResponse(res)
	if err != nil {
		return domain.HttpResponse{}, err
	}

	return response, nil
}

func (hc HttpClient) executeWithContext(ctx context.Context, req *fasthttp.Request, res *fasthttp.Response) error {
	errCh := make(chan error)
	go func() {
		errCh <- hc.client.Do(req, res)
	}()

	select {
	case e := <-errCh:
		return e
	case <-ctx.Done():
		return errExecuteRequestTimeout
	}
}
