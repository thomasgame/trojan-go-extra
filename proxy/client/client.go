package client

import (
	"context"

	"github.com/thomasgame/trojan-go-extra/config"
	"github.com/thomasgame/trojan-go-extra/proxy"
	"github.com/thomasgame/trojan-go-extra/tunnel/adapter"
	"github.com/thomasgame/trojan-go-extra/tunnel/compress"
	"github.com/thomasgame/trojan-go-extra/tunnel/http"
	"github.com/thomasgame/trojan-go-extra/tunnel/mux"
	"github.com/thomasgame/trojan-go-extra/tunnel/router"
	"github.com/thomasgame/trojan-go-extra/tunnel/shadowsocks"
	"github.com/thomasgame/trojan-go-extra/tunnel/simplesocks"
	"github.com/thomasgame/trojan-go-extra/tunnel/socks"
	"github.com/thomasgame/trojan-go-extra/tunnel/tls"
	"github.com/thomasgame/trojan-go-extra/tunnel/transport"
	"github.com/thomasgame/trojan-go-extra/tunnel/trojan"
	"github.com/thomasgame/trojan-go-extra/tunnel/websocket"
)

const Name = "CLIENT"

// GenerateClientTree generate general outbound protocol stack
// compressEnabled: 是否启用压缩（默认 true；显式 false 时关闭），压缩层位于 shadowsocks 与 trojan 之间
func GenerateClientTree(transportPlugin bool, muxEnabled bool, wsEnabled bool, ssEnabled bool, routerEnabled bool, compressEnabled bool) []string {
	clientStack := []string{transport.Name}
	if !transportPlugin {
		clientStack = append(clientStack, tls.Name)
	}
	if wsEnabled {
		clientStack = append(clientStack, websocket.Name)
	}
	if ssEnabled {
		clientStack = append(clientStack, shadowsocks.Name)
	}
	if compressEnabled {
		clientStack = append(clientStack, compress.Name)
	}
	clientStack = append(clientStack, trojan.Name)
	if muxEnabled {
		clientStack = append(clientStack, []string{mux.Name, simplesocks.Name}...)
	}
	if routerEnabled {
		clientStack = append(clientStack, router.Name)
	}
	return clientStack
}

func init() {
	proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
		cfg := config.FromContext(ctx, Name).(*Config)
		adapterServer, err := adapter.NewServer(ctx, nil)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithCancel(ctx)

		root := &proxy.Node{
			Name:       adapter.Name,
			Next:       make(map[string]*proxy.Node),
			IsEndpoint: false,
			Context:    ctx,
			Server:     adapterServer,
		}

		root.BuildNext(http.Name).IsEndpoint = true
		root.BuildNext(socks.Name).IsEndpoint = true

		clientStack := GenerateClientTree(cfg.TransportPlugin.Enabled, cfg.Mux.Enabled, cfg.Websocket.Enabled, cfg.Shadowsocks.Enabled, cfg.Router.Enabled, cfg.Compress.CompressionOn())
		c, err := proxy.CreateClientStack(ctx, clientStack)
		if err != nil {
			cancel()
			return nil, err
		}
		s := proxy.FindAllEndpoints(root)
		return proxy.NewProxy(ctx, cancel, s, c), nil
	})
}
