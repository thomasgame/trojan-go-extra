//go:build custom || full
// +build custom full

package build

import (
	_ "github.com/thomasgame/trojan-go-extra/proxy/custom"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/adapter"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/dokodemo"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/freedom"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/http"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/mux"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/router"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/shadowsocks"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/simplesocks"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/socks"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/tls"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/tproxy"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/transport"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/trojan"
	_ "github.com/thomasgame/trojan-go-extra/tunnel/websocket"
)
