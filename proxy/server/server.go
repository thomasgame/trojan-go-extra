package server

import (
	"context"

	"github.com/thomasgame/trojan-go-extra/config"
	"github.com/thomasgame/trojan-go-extra/proxy"
	"github.com/thomasgame/trojan-go-extra/proxy/client"
	"github.com/thomasgame/trojan-go-extra/tunnel/compress"
	"github.com/thomasgame/trojan-go-extra/tunnel/freedom"
	"github.com/thomasgame/trojan-go-extra/tunnel/mux"
	"github.com/thomasgame/trojan-go-extra/tunnel/router"
	"github.com/thomasgame/trojan-go-extra/tunnel/shadowsocks"
	"github.com/thomasgame/trojan-go-extra/tunnel/simplesocks"
	"github.com/thomasgame/trojan-go-extra/tunnel/tls"
	"github.com/thomasgame/trojan-go-extra/tunnel/transport"
	"github.com/thomasgame/trojan-go-extra/tunnel/trojan"
	"github.com/thomasgame/trojan-go-extra/tunnel/websocket"
)

const Name = "SERVER"

func init() {
	proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
		cfg := config.FromContext(ctx, Name).(*client.Config)
		ctx, cancel := context.WithCancel(ctx)
		transportServer, err := transport.NewServer(ctx, nil)
		if err != nil {
			cancel()
			return nil, err
		}
		clientStack := []string{freedom.Name}
		if cfg.Router.Enabled {
			clientStack = []string{freedom.Name, router.Name}
		}

		root := &proxy.Node{
			Name:       transport.Name,
			Next:       make(map[string]*proxy.Node),
			IsEndpoint: false,
			Context:    ctx,
			Server:     transportServer,
		}

		if !cfg.TransportPlugin.Enabled {
			root = root.BuildNext(tls.Name)
		}

		// trojan 子树：transport -> [tls] -> [shadowsocks] -> [compress] -> trojan -> ...
		trojanSubTree := root
		if cfg.Shadowsocks.Enabled {
			trojanSubTree = trojanSubTree.BuildNext(shadowsocks.Name)
		}
		if cfg.Compress.CompressionOn() {
			trojanSubTree = trojanSubTree.BuildNext(compress.Name)
		}
		trojanSubTree.BuildNext(trojan.Name).BuildNext(mux.Name).BuildNext(simplesocks.Name).IsEndpoint = true
		trojanSubTree.BuildNext(trojan.Name).IsEndpoint = true

		// websocket 子树：transport -> [tls] -> websocket -> [shadowsocks] -> [compress] -> trojan -> ...
		wsSubTree := root.BuildNext(websocket.Name)
		if cfg.Shadowsocks.Enabled {
			wsSubTree = wsSubTree.BuildNext(shadowsocks.Name)
		}
		if cfg.Compress.CompressionOn() {
			wsSubTree = wsSubTree.BuildNext(compress.Name)
		}
		wsSubTree.BuildNext(trojan.Name).BuildNext(mux.Name).BuildNext(simplesocks.Name).IsEndpoint = true
		wsSubTree.BuildNext(trojan.Name).IsEndpoint = true

		serverList := proxy.FindAllEndpoints(root)
		clientList, err := proxy.CreateClientStack(ctx, clientStack)
		if err != nil {
			cancel()
			return nil, err
		}
		return proxy.NewProxy(ctx, cancel, serverList, clientList), nil
	})
}
