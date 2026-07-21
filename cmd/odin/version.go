package main

import "runtime/debug"

// version 可在构建时通过 -ldflags "-X main.version=vX.Y.Z" 覆盖。
var version string

// currentVersion 优先使用构建参数，其次读取 go install 写入的模块版本。
func currentVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
