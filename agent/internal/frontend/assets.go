package frontend

import "embed"

//go:generate tailwindcss -i assets/styles/input.css -o assets/styles/output.css --minify

//go:embed assets
var AssetDir embed.FS
