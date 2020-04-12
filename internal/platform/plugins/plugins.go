package plugins

import (
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/b2wdigital/restQL-golang/internal/platform/logger"
	"github.com/b2wdigital/restQL-golang/pkg/restql"
	"github.com/pkg/errors"
)

type Manager interface {
	RunBeforeQuery(query string, queryCtx domain.QueryContext)
	RunAfterQuery(query string, result domain.Resources)
	RunBeforeRequest(request domain.HttpRequest)
	RunAfterRequest(request domain.HttpRequest, response domain.HttpResponse, err error)
}

type manager struct {
	log              *logger.Logger
	availablePlugins []restql.Plugin
}

func NewManager(log *logger.Logger, pluginsLocation string) (Manager, error) {
	ps, err := loadPlugins(log, pluginsLocation)
	if err != nil {
		return noOpManager{}, err
	}

	return manager{log: log, availablePlugins: ps}, nil
}

func (m manager) RunBeforeQuery(query string, queryCtx domain.QueryContext) {
	for _, p := range m.availablePlugins {
		m.safeExecute(p.Name(), "BeforeQuery", func() {
			p.BeforeQuery(query, queryCtx)
		})
	}
}

func (m manager) RunAfterQuery(query string, result domain.Resources) {
	for _, p := range m.availablePlugins {
		m.safeExecute(p.Name(), "AfterQuery", func() {
			m := convertQueryResult(result)
			p.AfterQuery(query, m)
		})
	}
}

func (m manager) RunBeforeRequest(request domain.HttpRequest) {
	for _, p := range m.availablePlugins {
		m.safeExecute(p.Name(), "BeforeRequest", func() {
			p.BeforeRequest(request)
		})
	}
}

func (m manager) RunAfterRequest(request domain.HttpRequest, response domain.HttpResponse, err error) {
	for _, p := range m.availablePlugins {
		m.safeExecute(p.Name(), "AfterRequest", func() {
			p.AfterRequest(request, response, err)
		})
	}
}

func (m manager) safeExecute(pluginName string, hook string, fn func()) {
	go func() {
		defer func() {
			if reason := recover(); reason != nil {
				err := errors.Errorf("reason : %v", reason)
				m.log.Error("plugin produced a panic", err, "name", pluginName, "hook", hook)
			}
		}()

		fn()
	}()
}

func convertQueryResult(resource interface{}) map[string]interface{} {
	switch resource := resource.(type) {
	case domain.Resources:
		m := make(map[string]interface{})
		for k, v := range resource {
			m[string(k)] = convertDoneResource(v)
		}
		return m
	case domain.Details:
		return map[string]interface{}{
			"status":       resource.Status,
			"success":      resource.Success,
			"ignoreErrors": resource.IgnoreErrors,
			"debugging":    convertQueryResult(resource.Debug),
		}
	case *domain.Debugging:
		return map[string]interface{}{
			"method":          resource.Method,
			"url":             resource.Url,
			"requestHeaders":  resource.RequestHeaders,
			"responseHeaders": resource.ResponseHeaders,
			"params":          resource.Params,
			"requestBody":     resource.RequestBody,
			"responseTime":    resource.ResponseTime,
		}
	default:
		return nil
	}
}

func convertDoneResource(doneResource interface{}) interface{} {
	switch resource := doneResource.(type) {
	case domain.DoneResource:
		return map[string]interface{}{
			"details": convertQueryResult(resource.Details),
			"result":  resource.Result,
		}
	case domain.DoneResources:
		l := make([]interface{}, len(resource))
		for i, r := range resource {
			l[i] = convertQueryResult(r)
		}
		return l
	default:
		return resource
	}
}

type noOpManager struct{}

func (n noOpManager) RunBeforeQuery(query string, queryCtx domain.QueryContext) {}
func (n noOpManager) RunAfterQuery(query string, result domain.Resources)       {}
func (n noOpManager) RunBeforeRequest(request domain.HttpRequest)               {}
func (n noOpManager) RunAfterRequest(request domain.HttpRequest, response domain.HttpResponse, err error) {
}