package compress

import "github.com/thomasgame/trojan-go-extra/config"

// CompressConfig 压缩插件配置
// 用于控制压缩算法、级别、阈值等参数
type CompressConfig struct {
	// Enabled 是否启用压缩：未配置或为 true 时默认开启；显式 false 时关闭（需与 proxy 层 CompressionOn 语义一致）
	Enabled *bool `json:"enabled" yaml:"enabled"`
	// Algorithm 压缩算法: "zstd", "snappy", "lz4"
	// zstd 压缩率高，snappy 延迟低，lz4 速度最快
	Algorithm string `json:"algorithm" yaml:"algorithm"`
	// Level 压缩级别，仅 zstd 有效，范围 1-22，默认 3
	// 级别越高压缩率越好但 CPU 占用更高
	Level int `json:"level" yaml:"level"`
	// MinSize 最小压缩阈值（字节）
	// 小于此值的数据不压缩（避免小包越压越大），默认 256
	MinSize int `json:"min_size" yaml:"min-size"`
	// BufferSize 压缩块大小（字节），默认 32768
	// 用于流式压缩时的缓冲
	BufferSize int `json:"buffer_size" yaml:"buffer-size"`
}

// CompressionOn 与 proxy/client.CompressConfig.CompressionOn 语义一致：默认开启
func (c CompressConfig) CompressionOn() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// Config 供 config 系统解析用的配置结构
// 从 JSON/YAML 根节点的 "compress" 字段解析
type Config struct {
	Compress CompressConfig `json:"compress" yaml:"compress"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			Compress: CompressConfig{
				// Enabled 默认 nil，即默认开启压缩
				Algorithm:  "zstd",
				Level:      3,
				MinSize:    256,
				BufferSize: 32768,
			},
		}
	})
}
