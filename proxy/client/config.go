package client

import "github.com/thomasgame/trojan-go-extra/config"

type MuxConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type WebsocketConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type RouterConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type ShadowsocksConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type TransportPluginConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// CompressConfig 压缩插件配置
// 用于在 proxy 层判断是否启用压缩并传递给 tunnel
// Enabled 为指针：未配置或为 true 时默认开启压缩；仅当显式设置 enabled 为 false 时关闭
type CompressConfig struct {
	Enabled *bool `json:"enabled" yaml:"enabled"`
}

// CompressionOn 是否启用传输层压缩（默认 true；显式 "enabled": false 时关闭）
func (c CompressConfig) CompressionOn() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

type Config struct {
	Mux             MuxConfig             `json:"mux" yaml:"mux"`
	Websocket       WebsocketConfig       `json:"websocket" yaml:"websocket"`
	Router          RouterConfig          `json:"router" yaml:"router"`
	Shadowsocks     ShadowsocksConfig     `json:"shadowsocks" yaml:"shadowsocks"`
	TransportPlugin TransportPluginConfig `json:"transport_plugin" yaml:"transport-plugin"`
	Compress        CompressConfig        `json:"compress" yaml:"compress"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
