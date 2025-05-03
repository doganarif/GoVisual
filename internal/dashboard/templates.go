package dashboard

import (
	"embed"
)

//go:embed templates/*.html templates/components/*.html
var templateFS embed.FS
