package compress

import (
	"context"

	"github.com/thomasgame/trojan-go-extra/common"
	"github.com/thomasgame/trojan-go-extra/config"
	"github.com/thomasgame/trojan-go-extra/log"
	"github.com/thomasgame/trojan-go-extra/tunnel"
)

// Server 压缩 tunnel 的服务端实现
// 接受来自 shadowsocks 的连接，对读取的数据解压、写入的数据压缩
type Server struct {
	underlay tunnel.Server
	config   *CompressConfig
}

// AcceptConn 接受连接，返回的 Conn 在 Read 时解压、Write 时压缩
func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := s.underlay.AcceptConn(&Tunnel{})
	if err != nil {
		return nil, common.NewError("compress failed to accept").Base(err)
	}
	compressConn, err := NewCompressConn(conn, s.config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return compressConn, nil
}

// AcceptPacket UDP 暂不支持
func (s *Server) AcceptPacket(overlay tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("compress tunnel does not support UDP yet")
}

// Close 关闭底层 server
func (s *Server) Close() error {
	return s.underlay.Close()
}

// NewServer 创建压缩服务端
func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
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
	cc := &compressCfg.Compress
	if cc.BufferSize <= 0 {
		cc.BufferSize = 32768
	}
	if cc.MinSize <= 0 {
		cc.MinSize = 256
	}
	log.Debug("compress server created, algorithm:", cc.Algorithm)
	return &Server{
		underlay: underlay,
		config:   cc,
	}, nil
}
