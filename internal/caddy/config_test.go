package caddy

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/eleven-sh/agent/proto"
	"github.com/eleven-sh/eleven/entities"
)

func TestCreateConfigFromServedPorts(t *testing.T) {
	testCases := []struct {
		test           string
		servedPorts    map[string]*proto.EnvServedPortBindings
		expectedConfig string
	}{
		{
			test:        "with no served ports",
			servedPorts: map[string]*proto.EnvServedPortBindings{},
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{}
					}
				}
			}`,
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
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{}
					}
				}
			}`,
		},

		{
			test: "with only domain bindings with redirect to HTTPS",
			servedPorts: map[string]*proto.EnvServedPortBindings{
				"8080": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "api.domain.com",
							Type:            string(entities.EnvServedPortBindingTypeDomain),
							RedirectToHttps: true,
						},
					},
				},
			},
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{
							"https-domains":{
								"listen":[
									":443"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"api.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							}
						}
					}
				}
			}`,
		},

		{
			test: "with only domain bindings without redirect to HTTPS",
			servedPorts: map[string]*proto.EnvServedPortBindings{
				"8080": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "api.domain.com",
							Type:            string(entities.EnvServedPortBindingTypeDomain),
							RedirectToHttps: false,
						},
					},
				},
			},
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{
							"https-domains":{
								"listen":[
									":443"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"api.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							},
							"http-domains":{
								"listen":[
									":80"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"api.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							}
						}
					}
				}
			}`,
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
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{
							"https-domains":{
								"listen":[
									":443"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"c.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:4000"
													}
												]
											}
										]
									},
									{
										"match":[
											{
												"host":[
													"a.domain.com",
													"b.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							},
							"http-domains": {
								"listen":[
									":80"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"c.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:4000"
													}
												]
											}
										]
									},
									{
										"match":[
											{
												"host":[
													"a.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							},
							"port-11000":{
								"listen":[
									":8000",
									":2000"
								],
								"routes":[
									{
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:11000"
													}
												]
											}
										]
									}
								]
							}
						}
					}
				}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			caddyConfig := CreateConfigFromServedPorts(tc.servedPorts)

			var expectedConfig *Config
			err := json.Unmarshal([]byte(tc.expectedConfig), &expectedConfig)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if !reflect.DeepEqual(caddyConfig, expectedConfig) {
				t.Fatalf(
					"expected config to equal '%+v', got '%+v'",
					expectedConfig,
					caddyConfig,
				)
			}
		})
	}
}

func TestUpdateConfigToCheckDomainReachability(t *testing.T) {
	testCases := []struct {
		test           string
		servedPorts    map[string]*proto.EnvServedPortBindings
		domain         string
		uniqueID       string
		expectedConfig string
	}{
		{
			test:        "with no served ports",
			servedPorts: map[string]*proto.EnvServedPortBindings{},
			domain:      "test.eleven.sh",
			uniqueID:    "unique_id",
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{
							"http-domains":{
								"listen": [
									":80"
								],
								"routes": [
									{
										"match":[
											{
												"host":[
													"test.eleven.sh"
												]
											}
										],
										"handle":[
											{
												"handler":"static_response",
												"body": "unique_id"
											}
										]
									}
								]
							}
						}
					}
				}
			}`,
		},

		{
			test: "with only domain bindings with redirect to HTTPS",
			servedPorts: map[string]*proto.EnvServedPortBindings{
				"8080": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "api.domain.com",
							Type:            string(entities.EnvServedPortBindingTypeDomain),
							RedirectToHttps: true,
						},
					},
				},
			},
			domain:   "test.eleven.sh",
			uniqueID: "unique_id",
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{
							"https-domains":{
								"listen":[
									":443"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"api.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							},
							"http-domains":{
								"listen": [
									":80"
								],
								"routes": [
									{
										"match":[
											{
												"host":[
													"test.eleven.sh"
												]
											}
										],
										"handle":[
											{
												"handler":"static_response",
												"body": "unique_id"
											}
										]
									}
								]
							}
						}
					}
				}
			}`,
		},

		{
			test: "with only domain bindings without redirect to HTTPS",
			servedPorts: map[string]*proto.EnvServedPortBindings{
				"8080": {
					Bindings: []*proto.EnvServedPortBinding{
						{
							Value:           "api.domain.com",
							Type:            string(entities.EnvServedPortBindingTypeDomain),
							RedirectToHttps: false,
						},
					},
				},
			},
			domain:   "test.eleven.sh",
			uniqueID: "unique_id",
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{
							"https-domains":{
								"listen":[
									":443"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"api.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							},
							"http-domains":{
								"listen":[
									":80"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"api.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									},
									{
										"match":[
											{
												"host":[
													"test.eleven.sh"
												]
											}
										],
										"handle":[
											{
												"handler":"static_response",
												"body":"unique_id"
											}
										]
									}
								]
							}
						}
					}
				}
			}`,
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
			domain:   "test.eleven.sh",
			uniqueID: "unique_id",
			expectedConfig: `{
				"apps":{
					"http":{
						"servers":{
							"https-domains":{
								"listen":[
									":443"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"c.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:4000"
													}
												]
											}
										]
									},
									{
										"match":[
											{
												"host":[
													"a.domain.com",
													"b.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									}
								]
							},
							"http-domains": {
								"listen":[
									":80"
								],
								"routes":[
									{
										"match":[
											{
												"host":[
													"c.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:4000"
													}
												]
											}
										]
									},
									{
										"match":[
											{
												"host":[
													"a.domain.com"
												]
											}
										],
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:8080"
													}
												]
											}
										]
									},
									{
										"match":[
											{
												"host":[
													"test.eleven.sh"
												]
											}
										],
										"handle":[
											{
												"handler":"static_response",
												"body": "unique_id"
											}
										]
									}
								]
							},
							"port-11000":{
								"listen":[
									":8000",
									":2000"
								],
								"routes":[
									{
										"handle":[
											{
												"handler":"reverse_proxy",
												"upstreams":[
													{
														"dial":"127.0.0.1:11000"
													}
												]
											}
										]
									}
								]
							}
						}
					}
				}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			caddyConfig := CreateConfigFromServedPorts(tc.servedPorts)

			UpdateConfigToCheckDomainReachability(
				caddyConfig,
				tc.domain,
				tc.uniqueID,
			)

			var expectedConfig *Config
			err := json.Unmarshal([]byte(tc.expectedConfig), &expectedConfig)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if !reflect.DeepEqual(caddyConfig, expectedConfig) {
				t.Fatalf(
					"expected config to equal '%+v', got '%+v'",
					expectedConfig,
					caddyConfig,
				)
			}
		})
	}
}
