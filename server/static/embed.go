package static

import "embed"

//go:embed color-thief.min.js
//go:embed htmx.min.js
//go:embed tailwind.css
var FS embed.FS
