//go:build mysql || full || mini
// +build mysql full mini

package build

import (
	_ "github.com/thomasgame/trojan-go-extra/statistic/mysql"
)
