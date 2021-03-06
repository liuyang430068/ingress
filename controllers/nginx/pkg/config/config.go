/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"runtime"

	"github.com/golang/glog"

	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/defaults"
)

const (
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	bodySize = "1m"

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	errorLevel = "notice"

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be accessed using HTTPS.
	hstsMaxAge = "15724800"

	// If UseProxyProtocol is enabled defIPCIDR defines the default the IP/network address of your external load balancer
	defIPCIDR = "0.0.0.0/0"

	gzipTypes = "application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"

	logFormatUpstream = `%v - [$proxy_add_x_forwarded_for] - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status"`

	logFormatStream = `[$time_local] $protocol [$ssl_preread_server_name] [$stream_upstream] $status $bytes_sent $bytes_received $session_time`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	sslBufferSize = "4k"

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA"

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	sslProtocols = "TLSv1 TLSv1.1 TLSv1.2"

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	sslSessionTimeout = "10m"

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	sslSessionCacheSize = "10m"
)

var (
	// SSLDirectory contains the mounted secrets with SSL certificates, keys and
	SSLDirectory = "/etc/ingress-controller/ssl"
)

// Configuration represents the content of nginx.conf file
type Configuration struct {
	defaults.Backend `json:",squash"`

	// EnableDynamicTLSRecords enables dynamic TLS record sizes
	// https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency
	// By default this is enabled
	EnableDynamicTLSRecords bool `json:"enable-dynamic-tls-records"`

	// ClientHeaderBufferSize allows to configure a custom buffer
	// size for reading client request header
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_buffer_size
	ClientHeaderBufferSize string `json:"client-header-buffer-size"`

	// DisableAccessLog disables the Access Log globally from NGINX ingress controller
	//http://nginx.org/en/docs/http/ngx_http_log_module.html
	DisableAccessLog bool `json:"disable-access-log,omitempty"`

	// EnableStickySessions enabled sticky sessions using cookies
	// https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng
	// By default this is disabled
	EnableStickySessions bool `json:"enable-sticky-sessions,omitempty"`

	// EnableVtsStatus allows the replacement of the default status page with a third party module named
	// nginx-module-vts - https://github.com/vozlt/nginx-module-vts
	// By default this is disabled
	EnableVtsStatus bool `json:"enable-vts-status,omitempty"`

	VtsStatusZoneSize string `json:"vts-status-zone-size,omitempty"`

	// RetryNonIdempotent since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH)
	// in case of an error. The previous behavior can be restored using the value true
	RetryNonIdempotent bool `json:"retry-non-idempotent"`

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	ErrorLogLevel string `json:"error-log-level,omitempty"`

	// Enables or disables the header HSTS in servers running SSL
	HSTS bool `json:"hsts,omitempty"`

	// Enables or disables the use of HSTS in all the subdomains of the servername
	// Default: true
	HSTSIncludeSubdomains bool `json:"hsts-include-subdomains,omitempty"`

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be
	// accessed using HTTPS.
	HSTSMaxAge string `json:"hsts-max-age,omitempty"`

	// Time during which a keep-alive client connection will stay open on the server side.
	// The zero value disables keep-alive client connections
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout
	KeepAlive int `json:"keep-alive,omitempty"`

	// LargeClientHeaderBuffers Sets the maximum number and size of buffers used for reading
	// large client request header.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#large_client_header_buffers
	// Default: 4 8k
	LargeClientHeaderBuffers string `json:"large-client-header-buffers"`

	// Customize upstream log_format
	// http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	LogFormatUpstream string `json:"log-format-upstream,omitempty"`

	// Customize stream log_format
	// http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	LogFormatStream string `json:"log-format-stream,omitempty"`

	// Maximum number of simultaneous connections that can be opened by each worker process
	// http://nginx.org/en/docs/ngx_core_module.html#worker_connections
	MaxWorkerConnections int `json:"max-worker-connections,omitempty"`

	// Sets the bucket size for the map variables hash tables.
	// Default value depends on the processor’s cache line size.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_bucket_size
	MapHashBucketSize int `json:"map-hash-bucket-size,omitempty"`

	// If UseProxyProtocol is enabled ProxyRealIPCIDR defines the default the IP/network address
	// of your external load balancer
	ProxyRealIPCIDR string `json:"proxy-real-ip-cidr,omitempty"`

	// Sets the name of the configmap that contains the headers to pass to the backend
	ProxySetHeaders string `json:"proxy-set-headers,omitempty"`

	// Maximum size of the server names hash tables used in server names, map directive’s values,
	// MIME types, names of request header strings, etcd.
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size
	ServerNameHashMaxSize int `json:"server-name-hash-max-size,omitempty"`

	// Size of the bucket for the server names hash tables
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size
	ServerNameHashBucketSize int `json:"server-name-hash-bucket-size,omitempty"`

	// Enables or disables emitting nginx version in error messages and in the “Server” response header field.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_tokens
	// Default: true
	ShowServerTokens bool `json:"server-tokens"`

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by
	// the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	SSLCiphers string `json:"ssl-ciphers,omitempty"`

	// Base64 string that contains Diffie-Hellman key to help with "Perfect Forward Secrecy"
	// https://www.openssl.org/docs/manmaster/apps/dhparam.html
	// https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam
	SSLDHParam string `json:"ssl-dh-param,omitempty"`

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	SSLProtocols string `json:"ssl-protocols,omitempty"`

	// Enables or disables the use of shared SSL cache among worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCache bool `json:"ssl-session-cache,omitempty"`

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCacheSize string `json:"ssl-session-cache-size,omitempty"`

	// Enables or disables session resumption through TLS session tickets.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	SSLSessionTickets bool `json:"ssl-session-tickets,omitempty"`

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	SSLSessionTimeout string `json:"ssl-session-timeout,omitempty"`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	SSLBufferSize string `json:"ssl-buffer-size,omitempty"`

	// Enables or disables the use of the PROXY protocol to receive client connection
	// (real IP address) information passed through proxy servers and load balancers
	// such as HAproxy and Amazon Elastic Load Balancer (ELB).
	// https://www.nginx.com/resources/admin-guide/proxy-protocol/
	UseProxyProtocol bool `json:"use-proxy-protocol,omitempty"`

	// Enables or disables the use of the nginx module that compresses responses using the "gzip" method
	// http://nginx.org/en/docs/http/ngx_http_gzip_module.html
	UseGzip bool `json:"use-gzip,omitempty"`

	// Enables or disables the HTTP/2 support in secure connections
	// http://nginx.org/en/docs/http/ngx_http_v2_module.html
	// Default: true
	UseHTTP2 bool `json:"use-http2,omitempty"`

	// MIME types in addition to "text/html" to compress. The special value “*” matches any MIME type.
	// Responses with the “text/html” type are always compressed if UseGzip is enabled
	GzipTypes string `json:"gzip-types,omitempty"`

	// Defines the number of worker processes. By default auto means number of available CPU cores
	// http://nginx.org/en/docs/ngx_core_module.html#worker_processes
	WorkerProcesses int `json:"worker-processes,omitempty"`
}

// NewDefault returns the default nginx configuration
func NewDefault() Configuration {
	cfg := Configuration{
		ClientHeaderBufferSize:  "1k",
		DisableAccessLog:        false,
		EnableDynamicTLSRecords: true,
		ErrorLogLevel:           errorLevel,
		HSTS:                    true,
		HSTSIncludeSubdomains:    true,
		HSTSMaxAge:               hstsMaxAge,
		GzipTypes:                gzipTypes,
		KeepAlive:                75,
		LargeClientHeaderBuffers: "4 8k",
		LogFormatStream:          logFormatStream,
		LogFormatUpstream:        logFormatUpstream,
		MaxWorkerConnections:     16384,
		MapHashBucketSize:        64,
		ProxyRealIPCIDR:          defIPCIDR,
		ServerNameHashMaxSize:    512,
		ServerNameHashBucketSize: 64,
		ShowServerTokens:         true,
		SSLBufferSize:            sslBufferSize,
		SSLCiphers:               sslCiphers,
		SSLProtocols:             sslProtocols,
		SSLSessionCache:          true,
		SSLSessionCacheSize:      sslSessionCacheSize,
		SSLSessionTickets:        true,
		SSLSessionTimeout:        sslSessionTimeout,
		UseProxyProtocol:         false,
		UseGzip:                  true,
		WorkerProcesses:          runtime.NumCPU(),
		VtsStatusZoneSize:        "10m",
		UseHTTP2:                 true,
		Backend: defaults.Backend{
			ProxyBodySize:        bodySize,
			ProxyConnectTimeout:  5,
			ProxyReadTimeout:     60,
			ProxySendTimeout:     60,
			ProxyBufferSize:      "4k",
			ProxyCookieDomain:    "off",
			ProxyCookiePath:      "off",
			SSLRedirect:          true,
			ForceSSLRedirect:     false,
			CustomHTTPErrors:     []int{},
			WhitelistSourceRange: []string{},
			SkipAccessLogURLs:    []string{},
			UsePortInRedirects:   false,
		},
	}

	if glog.V(5) {
		cfg.ErrorLogLevel = "debug"
	}

	return cfg
}

// BuildLogFormatUpstream format the log_format upstream using
// proxy_protocol_addr as remote client address if UseProxyProtocol
// is enabled.
func (cfg Configuration) BuildLogFormatUpstream() string {
	if cfg.LogFormatUpstream == logFormatUpstream {
		if cfg.UseProxyProtocol {
			return fmt.Sprintf(cfg.LogFormatUpstream, "$proxy_protocol_addr")
		}
		return fmt.Sprintf(cfg.LogFormatUpstream, "$remote_addr")
	}

	return cfg.LogFormatUpstream
}

// TemplateConfig contains the nginx configuration to render the file nginx.conf
type TemplateConfig struct {
	ProxySetHeaders     map[string]string
	MaxOpenFiles        int
	BacklogSize         int
	Backends            []*ingress.Backend
	PassthroughBackends []*ingress.SSLPassthroughBackend
	Servers             []*ingress.Server
	TCPBackends         []ingress.L4Service
	UDPBackends         []ingress.L4Service
	HealthzURI          string
	CustomErrors        bool
	Cfg                 Configuration
}
