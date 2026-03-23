package compress

import (
	"context"

	"github.com/thomasgame/trojan-go-extra/common"
	"github.com/thomasgame/trojan-go-extra/config"
	"github.com/thomasgame/trojan-go-extra/log"
	"github.com/thomasgame/trojan-go-extra/tunnel"
)

// Client 压缩 tunnel 的客户端实现
// 位于 shadowsocks 与 trojan 之间，对 trojan payload 进行压缩
type Client struct {
	underlay tunnel.Client
	config   *CompressConfig
}

// DialConn 建立连接，返回的 Conn 在 Write 时压缩、Read 时解压
func (c *Client) DialConn(addr *tunnel.Address, overlay tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(addr, &Tunnel{})
	if err != nil {
		return nil, err
	}
	compressConn, err := NewCompressConn(conn, c.config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return compressConn, nil
}

// DialPacket UDP 暂不支持压缩
func (c *Client) DialPacket(overlay tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("compress tunnel does not support UDP yet")
}

// Close 关闭底层 client
func (c *Client) Close() error {
	return c.underlay.Close()
}

// NewClient 创建压缩客户端
func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name)
	if cfg == nil {
		return nil, common.NewError("compress config not found")
	}
	compressCfg, ok := cfg.(*Config)
	if !ok {
		return nil, common.NewError("invalid compress config type")
	}
	if !compressCfg.Compress.CompressionOn() {
		return nil, common.NewError("compress is disabled")
	}
	// 设置默认值
	cc := &compressCfg.Compress
	if cc.BufferSize <= 0 {
		cc.BufferSize = 32768
	}
	if cc.MinSize <= 0 {
		cc.MinSize = 256
	}
	log.Debug("compress client created, algorithm:", cc.Algorithm, "buffer:", cc.BufferSize)
	return &Client{
		underlay: underlay,
		config:   cc,
	}, nil
}
