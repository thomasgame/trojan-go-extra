package compress

import (
	"bytes"
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
	"github.com/thomasgame/trojan-go-extra/common"
	"github.com/thomasgame/trojan-go-extra/log"
	"github.com/thomasgame/trojan-go-extra/tunnel"
)

// Conn 实现带压缩/解压的 tunnel.Conn
// 双向连接：Write 时压缩发送，Read 时解压接收
// 客户端和服务端使用相同的逻辑，两端对称
// 只嵌入 tunnel.Conn（已包含 net.Conn 与 Metadata），不可再嵌入 net.Conn，否则匿名字段名均为 Conn 会冲突
type Conn struct {
	tunnel.Conn

	config *CompressConfig

	// 压缩器：Write 时使用
	encoder  *zstd.Encoder
	writeBuf bytes.Buffer
	writeMu  sync.Mutex
	closed   bool

	// 解压器：Read 时使用
	decoder       *zstd.Decoder
	readBuf       bytes.Buffer
	readBufMu     sync.Mutex
	readExhausted bool // 底层连接已读完
}

// NewCompressConn 创建压缩连接
// conn: 底层 tunnel.Conn
// cfg: 压缩配置
func NewCompressConn(conn tunnel.Conn, cfg *CompressConfig) (*Conn, error) {
	// 根据配置的 Level(1-22) 选择压缩级别，默认使用 SpeedDefault
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, common.NewError("compress failed to create encoder").Base(err)
	}
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		encoder.Close()
		return nil, common.NewError("compress failed to create decoder").Base(err)
	}
	return &Conn{
		Conn:    conn,
		config:  cfg,
		encoder: encoder,
		decoder: decoder,
	}, nil
}

// Write 实现 io.Writer，对数据进行压缩后写入底层连接
// 小数据(<MinSize)不压缩直接发送，大数据分块压缩
func (c *Conn) Write(p []byte) (n int, err error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.closed {
		return 0, io.ErrClosedPipe
	}

	// 直接写入缓冲区
	c.writeBuf.Write(p)

	// 当缓冲区达到 chunkSize 或至少 minSize（满足可压缩条件）时发送
	// 这样在积累 minSize 后即可发送，降低小流量延迟
	chunkSize := c.config.BufferSize
	if chunkSize <= 0 {
		chunkSize = 32768
	}
	minSize := c.config.MinSize
	if minSize <= 0 {
		minSize = 256
	}

	for c.writeBuf.Len() >= minSize {
		toSend := chunkSize
		if c.writeBuf.Len() < chunkSize {
			toSend = c.writeBuf.Len()
		}
		chunk := make([]byte, toSend)
		_, _ = c.writeBuf.Read(chunk)
		if err = c.writeCompressedFrame(chunk, minSize); err != nil {
			return len(p), err
		}
	}

	// 不足 minSize 的数据保留在 buffer，待后续 Write 或 Close 时发送
	return len(p), nil
}

// writeCompressedFrame 将数据压缩（或原样）后写入一帧
func (c *Conn) writeCompressedFrame(data []byte, minSize int) error {
	if len(data) < minSize {
		return writeFrame(c.Conn, FlagUncompressed, data)
	}

	compressed := c.encoder.EncodeAll(data, nil)
	// 若压缩后更大，则发未压缩
	if len(compressed) >= len(data) {
		return writeFrame(c.Conn, FlagUncompressed, data)
	}
	return writeFrame(c.Conn, FlagCompressed, compressed)
}

// flushWriteBuf 将写缓冲区剩余数据刷到底层
func (c *Conn) flushWriteBuf() error {
	minSize := c.config.MinSize
	if minSize <= 0 {
		minSize = 256
	}
	for c.writeBuf.Len() > 0 {
		remaining := c.writeBuf.Len()
		chunkSize := c.config.BufferSize
		if chunkSize <= 0 {
			chunkSize = 32768
		}
		toRead := remaining
		if toRead > chunkSize {
			toRead = chunkSize
		}
		chunk := make([]byte, toRead)
		_, _ = c.writeBuf.Read(chunk)
		if err := c.writeCompressedFrame(chunk, minSize); err != nil {
			return err
		}
	}
	return nil
}

// Read 实现 io.Reader，从底层读取并解压
func (c *Conn) Read(p []byte) (n int, err error) {
	c.readBufMu.Lock()
	defer c.readBufMu.Unlock()

	// 若读缓冲有数据，先返回
	if c.readBuf.Len() > 0 {
		return c.readBuf.Read(p)
	}
	if c.readExhausted {
		return 0, io.EOF
	}

	// 从底层读取一帧
	flags, length, err := readFrameHeader(c.Conn)
	if err != nil {
		if err == io.EOF {
			c.readExhausted = true
		}
		return 0, err
	}

	if length == 0 {
		c.readExhausted = true
		return 0, io.EOF
	}

	payload := make([]byte, length)
	if _, err = io.ReadFull(c.Conn, payload); err != nil {
		return 0, err
	}

	var decoded []byte
	if flags == FlagCompressed {
		decoded, err = c.decoder.DecodeAll(payload, nil)
		if err != nil {
			log.Error(common.NewError("compress decode failed").Base(err))
			return 0, err
		}
	} else {
		decoded = payload
	}

	c.readBuf.Write(decoded)
	return c.readBuf.Read(p)
}

// Close 关闭连接，压缩方向需先刷完写缓冲
func (c *Conn) Close() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true

	if c.writeBuf.Len() > 0 {
		if err := c.flushWriteBuf(); err != nil {
			log.Debug("compress flush on close:", err)
		}
	}

	if c.encoder != nil {
		c.encoder.Close()
	}
	if c.decoder != nil {
		c.decoder.Close()
	}
	return c.Conn.Close()
}

// Metadata 返回底层连接的元数据
func (c *Conn) Metadata() *tunnel.Metadata {
	return c.Conn.Metadata()
}
