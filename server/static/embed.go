package static

import "embed"

//go:embed color-thief.min.js
//go:embed htmx.min.js
//go:embed tailwind.css
//go:embed icon-*.png
var FS embed.FS
