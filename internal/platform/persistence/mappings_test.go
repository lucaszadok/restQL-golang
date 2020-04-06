package persistence

import (
	"context"
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/b2wdigital/restQL-golang/internal/platform/logger"
	"io/ioutil"
	"reflect"
	"testing"
)

const defaultTenant = "default"

func TestMappingsReader_Env(t *testing.T) {
	envSource := stubEnvSource{
		getAll: map[string]string{
			"RESTQL_MAPPING_HERO":     "http://hero.api/",
			"RESTQL_MAPPING_SIDEKICK": "http://sidekick.api/",
			"TEST_VAR":                "foo",
		},
	}

	reader := NewMappingReader(noOpLogger, envSource, map[string]string{}, nil)

	expected := map[string]domain.Mapping{
		"hero": {
			ResourceName:  "hero",
			Schema:        "http",
			Uri:           "hero.api/",
			PathParams:    []string{},
			PathParamsSet: map[string]struct{}{},
		},
		"sidekick": {
			ResourceName:  "sidekick",
			Schema:        "http",
			Uri:           "sidekick.api/",
			PathParams:    []string{},
			PathParamsSet: map[string]struct{}{},
		},
	}

	mappings, err := reader.FromTenant(context.Background(), defaultTenant)

	if err != nil {
		t.Fatalf("FromTenant returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mappings, expected) {
		t.Fatalf("FromTenant = %+#v, want = %+#v", mappings, expected)
	}
}

func TestMappingsReader_Local(t *testing.T) {
	envSource := stubEnvSource{getAll: map[string]string{}}
	local := map[string]string{
		"hero":     "http://hero.api/",
		"sidekick": "http://sidekick.api/",
	}

	reader := NewMappingReader(noOpLogger, envSource, local, nil)

	expected := map[string]domain.Mapping{
		"hero": {
			ResourceName:  "hero",
			Schema:        "http",
			Uri:           "hero.api/",
			PathParams:    []string{},
			PathParamsSet: map[string]struct{}{},
		},
		"sidekick": {
			ResourceName:  "sidekick",
			Schema:        "http",
			Uri:           "sidekick.api/",
			PathParams:    []string{},
			PathParamsSet: map[string]struct{}{},
		},
	}

	mappings, err := reader.FromTenant(context.Background(), defaultTenant)

	if err != nil {
		t.Fatalf("FromTenant returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mappings, expected) {
		t.Fatalf("FromTenant = %+#v, want = %+#v", mappings, expected)
	}
}

func TestMappingsReader_ShouldOverwriteMappings(t *testing.T) {
	envSource := stubEnvSource{
		getAll: map[string]string{
			"RESTQL_MAPPING_HERO": "https://hero.com/api/",
		},
	}
	local := map[string]string{
		"hero":     "http://hero.api/",
		"sidekick": "http://sidekick.api/",
	}

	reader := NewMappingReader(noOpLogger, envSource, local, nil)

	expected := map[string]domain.Mapping{
		"hero": {
			ResourceName:  "hero",
			Schema:        "https",
			Uri:           "hero.com/api/",
			PathParams:    []string{},
			PathParamsSet: map[string]struct{}{},
		},
		"sidekick": {
			ResourceName:  "sidekick",
			Schema:        "http",
			Uri:           "sidekick.api/",
			PathParams:    []string{},
			PathParamsSet: map[string]struct{}{},
		},
	}

	mappings, err := reader.FromTenant(context.Background(), defaultTenant)

	if err != nil {
		t.Fatalf("FromTenant returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mappings, expected) {
		t.Fatalf("FromTenant = %+#v, want = %+#v", mappings, expected)
	}
}

var noOpLogger = logger.New(ioutil.Discard, logger.LogOptions{})

type stubEnvSource struct {
	getAll map[string]string
}

func (s stubEnvSource) GetString(key string) string {
	return ""
}

func (s stubEnvSource) GetAll() map[string]string {
	return s.getAll
}