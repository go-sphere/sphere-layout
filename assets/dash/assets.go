package dash

import "embed"

//go:embed index.html
var Assets embed.FS

var AssetsPath = "."
