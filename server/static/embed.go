package static

import "embed"

//go:embed color-thief.min.js
//go:embed htmx.min.js
//go:embed tailwind.css
//go:embed icon-*.png
//go:embed chartjs.min.js
var FS embed.FS
