package compress

import (
	"context"

	"github.com/thomasgame/trojan-go-extra/tunnel"
)

const (
	// Name 隧道名称，用于注册和配置
	Name = "COMPRESS"
)

// Tunnel 实现 tunnel.Tunnel 接口，用于创建 Client 和 Server
type Tunnel struct{}

// Name 返回隧道名称
func (t *Tunnel) Name() string {
	return Name
}

// NewClient 创建压缩客户端，包装底层 tunnel.Client
func (t *Tunnel) NewClient(ctx context.Context, client tunnel.Client) (tunnel.Client, error) {
	return NewClient(ctx, client)
}

// NewServer 创建压缩服务端，包装底层 tunnel.Server
func (t *Tunnel) NewServer(ctx context.Context, server tunnel.Server) (tunnel.Server, error) {
	return NewServer(ctx, server)
}

func init() {
	tunnel.RegisterTunnel(Name, &Tunnel{})
}
