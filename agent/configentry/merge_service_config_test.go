// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configentry

import (
	"testing"

	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/consul/agent/structs"
)

func Test_MergeServiceConfig_TransparentProxy(t *testing.T) {
	type args struct {
		defaults *structs.ServiceConfigResponse
		service  *structs.NodeService
	}
	tests := []struct {
		name string
		args args
		want *structs.NodeService
	}{
		{
			name: "inherit transparent proxy settings + kitchen sink",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					Mode: structs.ProxyModeTransparent,
					TransparentProxy: structs.TransparentProxyConfig{
						OutboundListenerPort: 10101,
						DialedDirectly:       true,
					},
					ProxyConfig: map[string]interface{}{
						"foo": "bar",
					},
					MutualTLSMode: structs.MutualTLSModePermissive,
					Expose: structs.ExposeConfig{
						Checks: true,
						Paths: []structs.ExposePath{
							{
								ListenerPort: 8080,
								Path:         "/",
								Protocol:     "http",
							},
						},
					},
					MeshGateway: structs.MeshGatewayConfig{Mode: structs.MeshGatewayModeRemote},
					AccessLogs: structs.AccessLogsConfig{
						Enabled:             true,
						DisableListenerLogs: true,
						Type:                structs.FileLogSinkType,
						Path:                "/tmp/accesslog.txt",
						JSONFormat:          "{ \"custom_start_time\": \"%START_TIME%\" }",
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						Mode:                   structs.ProxyModeDefault,
						TransparentProxy:       structs.TransparentProxyConfig{},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					Mode:                   structs.ProxyModeTransparent,
					TransparentProxy: structs.TransparentProxyConfig{
						OutboundListenerPort: 10101,
						DialedDirectly:       true,
					},
					MutualTLSMode: structs.MutualTLSModePermissive,
					Config: map[string]interface{}{
						"foo": "bar",
					},
					Expose: structs.ExposeConfig{
						Checks: true,
						Paths: []structs.ExposePath{
							{
								ListenerPort: 8080,
								Path:         "/",
								Protocol:     "http",
							},
						},
					},
					MeshGateway: structs.MeshGatewayConfig{Mode: structs.MeshGatewayModeRemote},
					AccessLogs: structs.AccessLogsConfig{
						Enabled:             true,
						DisableListenerLogs: true,
						Type:                structs.FileLogSinkType,
						Path:                "/tmp/accesslog.txt",
						JSONFormat:          "{ \"custom_start_time\": \"%START_TIME%\" }",
					},
				},
			},
		},
		{
			name: "override transparent proxy settings",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					Mode: structs.ProxyModeTransparent,
					TransparentProxy: structs.TransparentProxyConfig{
						OutboundListenerPort: 10101,
						DialedDirectly:       false,
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						Mode:                   structs.ProxyModeDirect,
						TransparentProxy: structs.TransparentProxyConfig{
							OutboundListenerPort: 808,
							DialedDirectly:       true,
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					Mode:                   structs.ProxyModeDirect,
					TransparentProxy: structs.TransparentProxyConfig{
						OutboundListenerPort: 808,
						DialedDirectly:       true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultsCopy, err := copystructure.Copy(tt.args.defaults)
			require.NoError(t, err)

			got, err := MergeServiceConfig(tt.args.defaults, tt.args.service)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			// The input defaults must not be modified by the merge.
			// See PR #10647
			assert.Equal(t, tt.args.defaults, defaultsCopy)
		})
	}
}

func Test_MergeServiceConfig_Extensions(t *testing.T) {
	type args struct {
		defaults *structs.ServiceConfigResponse
		service  *structs.NodeService
	}
	tests := []struct {
		name string
		args args
		want *structs.NodeService
	}{
		{
			name: "inherit extensions",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					EnvoyExtensions: []structs.EnvoyExtension{
						{
							Name:     "ext1",
							Required: true,
							Arguments: map[string]interface{}{
								"arg1": "val1",
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					EnvoyExtensions: []structs.EnvoyExtension{
						{
							Name:     "ext1",
							Required: true,
							Arguments: map[string]interface{}{
								"arg1": "val1",
							},
						},
					},
				},
			},
		},
		{
			name: "replaces existing extensions",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					EnvoyExtensions: []structs.EnvoyExtension{
						{
							Name:     "ext1",
							Required: true,
							Arguments: map[string]interface{}{
								"arg1": "val1",
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						EnvoyExtensions: []structs.EnvoyExtension{
							{
								Name:     "existing-ext",
								Required: true,
								Arguments: map[string]interface{}{
									"arg1": "val1",
								},
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					EnvoyExtensions: []structs.EnvoyExtension{
						{
							Name:     "ext1",
							Required: true,
							Arguments: map[string]interface{}{
								"arg1": "val1",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultsCopy, err := copystructure.Copy(tt.args.defaults)
			require.NoError(t, err)

			got, err := MergeServiceConfig(tt.args.defaults, tt.args.service)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			// The input defaults must not be modified by the merge.
			// See PR #10647
			assert.Equal(t, tt.args.defaults, defaultsCopy)
		})
	}
}

func Test_MergeServiceConfig_UpstreamOverrides(t *testing.T) {
	type args struct {
		defaults *structs.ServiceConfigResponse
		service  *structs.NodeService
	}
	zapUpstreamId := structs.PeeredServiceName{
		ServiceName: structs.NewServiceName("zap", structs.DefaultEnterpriseMetaInDefaultPartition()),
	}
	zapPeeredUpstreamId := structs.PeeredServiceName{
		Peer:        "some-peer",
		ServiceName: structs.NewServiceName("zap", structs.DefaultEnterpriseMetaInDefaultPartition()),
	}
	tests := []struct {
		name string
		args args
		want *structs.NodeService
	}{
		{
			name: "new config fields",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapUpstreamId,
							Config: map[string]interface{}{
								"passive_health_check": map[string]interface{}{
									"Interval":    int64(10),
									"MaxFailures": int64(2),
								},
								"mesh_gateway": map[string]interface{}{
									"Mode": "local",
								},
								"protocol": "grpc",
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationNamespace: "default",
								DestinationPartition: "default",
								DestinationName:      "zap",
								Config: map[string]interface{}{
									"passive_health_check": map[string]interface{}{
										"Interval":    int64(20),
										"MaxFailures": int64(4),
									},
								},
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationNamespace: "default",
							DestinationPartition: "default",
							DestinationName:      "zap",
							Config: map[string]interface{}{
								"passive_health_check": map[string]interface{}{
									"Interval":    int64(20),
									"MaxFailures": int64(4),
								},
								"protocol": "grpc",
							},
							MeshGateway: structs.MeshGatewayConfig{
								Mode: structs.MeshGatewayModeLocal,
							},
						},
					},
				},
			},
		},
		{
			name: "remote upstream config expands local upstream list in transparent mode",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapUpstreamId,
							Config: map[string]interface{}{
								"protocol": "grpc",
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						Mode:                   structs.ProxyModeTransparent,
						TransparentProxy: structs.TransparentProxyConfig{
							OutboundListenerPort: 10101,
							DialedDirectly:       true,
						},
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationNamespace: "default",
								DestinationPartition: "default",
								DestinationName:      "zip",
								LocalBindPort:        8080,
								Config: map[string]interface{}{
									"protocol": "http",
								},
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					Mode:                   structs.ProxyModeTransparent,
					TransparentProxy: structs.TransparentProxyConfig{
						OutboundListenerPort: 10101,
						DialedDirectly:       true,
					},
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationNamespace: "default",
							DestinationPartition: "default",
							DestinationName:      "zip",
							LocalBindPort:        8080,
							Config: map[string]interface{}{
								"protocol": "http",
							},
						},
						structs.Upstream{
							DestinationNamespace: "default",
							DestinationPartition: "default",
							DestinationName:      "zap",
							Config: map[string]interface{}{
								"protocol": "grpc",
							},
							CentrallyConfigured: true,
						},
					},
				},
			},
		},
		{
			name: "remote upstream config not added to local upstream list outside of transparent mode",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapUpstreamId,
							Config: map[string]interface{}{
								"protocol": "grpc",
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						Mode:                   structs.ProxyModeDirect,
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationNamespace: "default",
								DestinationPartition: "default",
								DestinationName:      "zip",
								LocalBindPort:        8080,
								Config: map[string]interface{}{
									"protocol": "http",
								},
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					Mode:                   structs.ProxyModeDirect,
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationNamespace: "default",
							DestinationPartition: "default",
							DestinationName:      "zip",
							LocalBindPort:        8080,
							Config: map[string]interface{}{
								"protocol": "http",
							},
						},
					},
				},
			},
		},
		{
			name: "upstream mode from remote defaults overrides local default",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapUpstreamId,
							Config: map[string]interface{}{
								"mesh_gateway": map[string]interface{}{
									"Mode": "local",
								},
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						MeshGateway: structs.MeshGatewayConfig{
							Mode: structs.MeshGatewayModeRemote,
						},
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationNamespace: "default",
								DestinationPartition: "default",
								DestinationName:      "zap",
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					MeshGateway: structs.MeshGatewayConfig{
						Mode: structs.MeshGatewayModeRemote,
					},
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationNamespace: "default",
							DestinationPartition: "default",
							DestinationName:      "zap",
							Config:               map[string]interface{}{},
							MeshGateway: structs.MeshGatewayConfig{
								Mode: structs.MeshGatewayModeLocal,
							},
						},
					},
				},
			},
		},
		{
			name: "mode in local upstream config overrides all",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapUpstreamId,
							Config: map[string]interface{}{
								"mesh_gateway": map[string]interface{}{
									"Mode": "local",
								},
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						DestinationServiceName: "foo",
						DestinationServiceID:   "foo",
						MeshGateway: structs.MeshGatewayConfig{
							Mode: structs.MeshGatewayModeRemote,
						},
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationNamespace: "default",
								DestinationPartition: "default",
								DestinationName:      "zap",
								MeshGateway: structs.MeshGatewayConfig{
									Mode: structs.MeshGatewayModeNone,
								},
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "foo",
					DestinationServiceID:   "foo",
					MeshGateway: structs.MeshGatewayConfig{
						Mode: structs.MeshGatewayModeRemote,
					},
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationNamespace: "default",
							DestinationPartition: "default",
							DestinationName:      "zap",
							Config:               map[string]interface{}{},
							MeshGateway: structs.MeshGatewayConfig{
								Mode: structs.MeshGatewayModeNone,
							},
						},
					},
				},
			},
		},
		{
			name: "peering upstreams are distinct from local-cluster upstreams",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapUpstreamId,
							Config: map[string]interface{}{
								"connect_timeout_ms": 2222,
							},
						},
						{
							Upstream: zapPeeredUpstreamId,
							Config: map[string]interface{}{
								"connect_timeout_ms": 3333,
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationName: "zap",
							},
							structs.Upstream{
								DestinationPeer: "some-peer",
								DestinationName: "zap",
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationName: "zap",
							Config: map[string]interface{}{
								"connect_timeout_ms": 2222,
							},
						},
						structs.Upstream{
							DestinationPeer: "some-peer",
							DestinationName: "zap",
							Config: map[string]interface{}{
								"connect_timeout_ms": 3333,
							},
						},
					},
				},
			},
		},
		{
			name: "peering upstreams ignore protocol overrides",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapPeeredUpstreamId,
							Config: map[string]interface{}{
								"protocol": "http",
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationPeer: "some-peer",
								DestinationName: "zap",
								Config: map[string]interface{}{
									"protocol": "tcp",
								},
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationPeer: "some-peer",
							DestinationName: "zap",
							Config: map[string]interface{}{
								"protocol": "tcp",
							},
						},
					},
				},
			},
		},
		{
			name: "peering upstreams ignore protocol overrides with unset value",
			args: args{
				defaults: &structs.ServiceConfigResponse{
					UpstreamConfigs: structs.OpaqueUpstreamConfigs{
						{
							Upstream: zapPeeredUpstreamId,
							Config: map[string]interface{}{
								"protocol": "http",
							},
						},
					},
				},
				service: &structs.NodeService{
					ID:      "foo-proxy",
					Service: "foo-proxy",
					Proxy: structs.ConnectProxyConfig{
						Upstreams: structs.Upstreams{
							structs.Upstream{
								DestinationPeer: "some-peer",
								DestinationName: "zap",
								Config:          map[string]interface{}{},
							},
						},
					},
				},
			},
			want: &structs.NodeService{
				ID:      "foo-proxy",
				Service: "foo-proxy",
				Proxy: structs.ConnectProxyConfig{
					Upstreams: structs.Upstreams{
						structs.Upstream{
							DestinationPeer: "some-peer",
							DestinationName: "zap",
							Config:          map[string]interface{}{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultsCopy, err := copystructure.Copy(tt.args.defaults)
			require.NoError(t, err)

			got, err := MergeServiceConfig(tt.args.defaults, tt.args.service)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			// The input defaults must not be modified by the merge.
			// See PR #10647
			assert.Equal(t, tt.args.defaults, defaultsCopy)
		})
	}
}
