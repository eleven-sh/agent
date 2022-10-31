package caddy

import (
	"net"
	"sort"

	"github.com/eleven-sh/agent/proto"
	"github.com/eleven-sh/eleven/entities"
)

const (
	configHTTPAppKey             = "http"
	configServersHTTPSDomainsKey = "https-domains"
	configServersHTTPDomainsKey  = "http-domains"
	configServersPortsKeyPrefix  = "port-"
	configServersRPHandler       = "reverse_proxy"
	configServersStaticHandler   = "static_response"
)

type Config struct {
	Apps map[string]ConfigHTTPApp `json:"apps"`
}

type ConfigHTTPApp struct {
	Servers ConfigHTTPServers `json:"servers"`
}

type ConfigHTTPServers map[string]ConfigHTTPServer

type ConfigHTTPServer struct {
	Listen []string                `json:"listen"`
	Routes []ConfigHTTPServerRoute `json:"routes"`
}

type ConfigHTTPServerRoute struct {
	Match  []ConfigHTTPServerMatch  `json:"match,omitempty"`
	Handle []ConfigHTTPServerHandle `json:"handle"`
}

type ConfigHTTPServerMatch struct {
	Host []string `json:"host"`
}

type ConfigHTTPServerHandle struct {
	Handler   string                      `json:"handler"`
	Upstreams []ConfigHTTPServerUpstreams `json:"upstreams,omitempty"`
	Body      string                      `json:"body,omitempty"`
}

type ConfigHTTPServerUpstreams struct {
	Dial string `json:"dial"`
}

type servedPorts []servedPort
type servedPort struct {
	port     string
	bindings servedPortBindings
}
type servedPortBindings struct {
	httpsDomains []string
	httpDomains  []string
	ports        []string
}

func CreateConfigFromServedPorts(
	ports map[string]*proto.EnvServedPortBindings,
) *Config {

	servedPorts := buildServedPorts(ports)
	httpServersConfig := ConfigHTTPServers{}

	for _, servedPort := range servedPorts {
		port := servedPort.port
		bindings := servedPort.bindings

		if len(bindings.httpsDomains) > 0 {
			httpsConfig, hasHTTPSConfig := httpServersConfig[configServersHTTPSDomainsKey]

			if !hasHTTPSConfig {
				httpsConfig = ConfigHTTPServer{
					Listen: []string{":443"},
					Routes: []ConfigHTTPServerRoute{},
				}
			}

			newRoute := func(domains []string) ConfigHTTPServerRoute {
				return ConfigHTTPServerRoute{
					Match: []ConfigHTTPServerMatch{
						{
							Host: domains,
						},
					},
					Handle: []ConfigHTTPServerHandle{
						{
							Handler: configServersRPHandler,
							Upstreams: []ConfigHTTPServerUpstreams{
								{
									Dial: net.JoinHostPort(
										"127.0.0.1",
										port,
									),
								},
							},
						},
					},
				}
			}

			httpsConfig.Routes = append(
				httpsConfig.Routes,
				newRoute(bindings.httpsDomains),
			)

			httpServersConfig[configServersHTTPSDomainsKey] = httpsConfig

			if len(bindings.httpDomains) > 0 {
				httpConfig, hasHTTPConfig := httpServersConfig[configServersHTTPDomainsKey]

				if !hasHTTPConfig {
					httpConfig = ConfigHTTPServer{
						Listen: []string{":80"},
						Routes: []ConfigHTTPServerRoute{},
					}
				}

				httpConfig.Routes = append(
					httpConfig.Routes,
					newRoute(bindings.httpDomains),
				)

				httpServersConfig[configServersHTTPDomainsKey] = httpConfig
			}
		}

		if len(bindings.ports) == 0 {
			continue
		}

		httpServersConfig[configServersPortsKeyPrefix+port] = ConfigHTTPServer{
			Listen: bindings.ports,
			Routes: []ConfigHTTPServerRoute{
				{
					Handle: []ConfigHTTPServerHandle{
						{
							Handler: configServersRPHandler,
							Upstreams: []ConfigHTTPServerUpstreams{
								{
									Dial: net.JoinHostPort(
										"127.0.0.1",
										port,
									),
								},
							},
						},
					},
				},
			},
		}
	}

	return &Config{
		Apps: map[string]ConfigHTTPApp{
			configHTTPAppKey: {
				Servers: httpServersConfig,
			},
		},
	}
}

func UpdateConfigToCheckDomainReachability(
	config *Config,
	domain string,
	uniqueID string,
) {

	httpConfig, hasHTTPConfig := config.Apps[configHTTPAppKey].Servers[configServersHTTPDomainsKey]

	if !hasHTTPConfig {
		httpConfig = ConfigHTTPServer{
			Listen: []string{":80"},
			Routes: []ConfigHTTPServerRoute{},
		}
	}

	httpConfig.Routes = append(httpConfig.Routes, ConfigHTTPServerRoute{
		Match: []ConfigHTTPServerMatch{
			{
				Host: []string{domain},
			},
		},

		Handle: []ConfigHTTPServerHandle{
			{
				Handler: configServersStaticHandler,
				Body:    uniqueID,
			},
		},
	})

	config.Apps[configHTTPAppKey].Servers[configServersHTTPDomainsKey] = httpConfig
}

func buildServedPorts(
	ports map[string]*proto.EnvServedPortBindings,
) servedPorts {

	// To be allowed to write tests,
	// we need to have the same configuration components order,
	// not a random one
	sortedPorts := []string{}
	for port := range ports {
		sortedPorts = append(sortedPorts, port)
	}
	sort.Strings(sortedPorts)

	servedPorts := servedPorts{}

	for _, port := range sortedPorts {
		portBindings := ports[port]

		httpsDomains := []string{}
		httpDomains := []string{}
		ports := []string{}

		for _, binding := range portBindings.Bindings {

			if binding.Type == string(entities.EnvServedPortBindingTypeDomain) {

				httpsDomains = append(
					httpsDomains,
					binding.Value,
				)

				if !binding.RedirectToHttps {
					httpDomains = append(
						httpDomains,
						binding.Value,
					)
				}

				continue
			}

			// Port already bound by user application
			if binding.Value == port {
				continue
			}

			ports = append(ports, ":"+binding.Value)
		}

		servedPorts = append(servedPorts, servedPort{
			port: port,
			bindings: servedPortBindings{
				httpsDomains: httpsDomains,
				httpDomains:  httpDomains,
				ports:        ports,
			},
		})
	}

	return servedPorts
}
