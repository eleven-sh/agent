package grpcserver

import (
	"reflect"
	"testing"

	"github.com/eleven-sh/agent/internal/env"
	"github.com/eleven-sh/agent/proto"
	"github.com/eleven-sh/eleven/entities"
)

func TestGetConfigServedPortsFromProto(t *testing.T) {
	testCases := []struct {
		test           string
		servedPorts    map[string]*proto.EnvServedPortBindings
		expectedConfig env.ConfigServedPorts
	}{
		{
			test:           "with no served ports",
			servedPorts:    map[string]*proto.EnvServedPortBindings{},
			expectedConfig: env.ConfigServedPorts{},
		},

		{
			test: "without domain bindings",
			servedPorts: map[string]*proto.EnvServedPortBindings{
				"8080": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value: "8080",
							Type:  string(entities.EnvServedPortBindingTypePort),
						},
					},
				},
			},
			expectedConfig: env.ConfigServedPorts{
				env.ConfigServedPort("8080"): true,
			},
		},

		{
			test: "with only domain bindings",
			servedPorts: map[string]*proto.EnvServedPortBindings{
				"8080": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value: "api.domain.com",
							Type:  string(entities.EnvServedPortBindingTypeDomain),
						},
					},
				},
			},
			expectedConfig: env.ConfigServedPorts{},
		},

		{
			test: "with all port binding types",
			servedPorts: map[string]*proto.EnvServedPortBindings{
				"6000": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "6000",
							Type:            string(entities.EnvServedPortBindingTypePort),
							RedirectToHttps: false,
						},
					},
				},

				"4000": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "c.domain.com",
							Type:            string(entities.EnvServedPortBindingTypeDomain),
							RedirectToHttps: false,
						},
					},
				},

				"8080": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "a.domain.com",
							Type:            string(entities.EnvServedPortBindingTypeDomain),
							RedirectToHttps: false,
						},

						{
							Value:           "b.domain.com",
							Type:            string(entities.EnvServedPortBindingTypeDomain),
							RedirectToHttps: true,
						},
					},
				},

				"11000": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "11000",
							Type:            string(entities.EnvServedPortBindingTypePort),
							RedirectToHttps: false,
						},

						{
							Value:           "8000",
							Type:            string(entities.EnvServedPortBindingTypePort),
							RedirectToHttps: false,
						},

						{
							Value:           "2000",
							Type:            string(entities.EnvServedPortBindingTypePort),
							RedirectToHttps: false,
						},
					},
				},
			},
			expectedConfig: env.ConfigServedPorts{
				env.ConfigServedPort("2000"):  true,
				env.ConfigServedPort("6000"):  true,
				env.ConfigServedPort("8000"):  true,
				env.ConfigServedPort("11000"): true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			servedPortsCfg := getConfigServedPortsFromProto(tc.servedPorts)

			if !reflect.DeepEqual(servedPortsCfg, tc.expectedConfig) {
				t.Fatalf(
					"expected config to equal '%+v', got '%+v'",
					tc.expectedConfig,
					servedPortsCfg,
				)
			}
		})
	}
}
