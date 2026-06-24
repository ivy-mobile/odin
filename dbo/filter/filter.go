package filter

import (
	"dubbo.apache.org/dubbo-go/v3/common/extension"

	"github.com/ivy-mobile/odin/dbo/filter/logging"
	xlogv2 "github.com/ivy-mobile/odin/xutil/xlog/v2"
)

const (
	FilterLogging = "logging"
)

// Init 初始化
func Init(logger xlogv2.Logger) {
	extension.SetFilter(FilterLogging, logging.NewLogFilter(logger))
}
