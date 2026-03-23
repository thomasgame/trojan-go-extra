// Package compress 实现 Trojan-Go 的透明压缩隧道
// 在 shadowsocks 与 trojan 之间对 payload 进行 zstd 压缩，减少传输带宽
package compress

import (
	"encoding/binary"
	"io"
)

// 帧标志位定义
const (
	// FlagUncompressed 未压缩标志 (0)
	// 当数据小于 MinSize 或压缩后更大时使用
	FlagUncompressed byte = 0
	// FlagCompressed 压缩标志 (1)
	FlagCompressed byte = 1
)

// frameHeader 帧头大小：1 byte flags + 4 byte length
const frameHeaderSize = 5

// writeFrame 将数据以帧格式写入，可选压缩
// flags: FlagUncompressed 或 FlagCompressed
// w: 底层 io.Writer
// data: 原始或已压缩的数据
func writeFrame(w io.Writer, flags byte, data []byte) error {
	length := uint32(len(data))
	header := make([]byte, frameHeaderSize)
	header[0] = flags
	binary.BigEndian.PutUint32(header[1:5], length)

	if _, err := w.Write(header); err != nil {
		return err
	}
	if length > 0 {
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

// readFrameHeader 读取帧头，返回 flags 和 payload 长度
func readFrameHeader(r io.Reader) (flags byte, length uint32, err error) {
	header := make([]byte, frameHeaderSize)
	if _, err = io.ReadFull(r, header); err != nil {
		return 0, 0, err
	}
	return header[0], binary.BigEndian.Uint32(header[1:5]), nil
}
