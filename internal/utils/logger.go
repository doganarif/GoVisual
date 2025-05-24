package utils

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/mattn/go-isatty"
)

var (
	green   = "\033[32m"
	white   = "\033[37m"
	red     = "\033[31m"
	blue    = "\033[34m"
	yellow  = "\033[33m"
	gray    = "\033[90m"
	black   = "\033[30m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	reset   = "\033[0m"
)

func colorizeMethod(method string) string {
	if method == "" {
		return ""
	}

	var color string
	switch method {
	case http.MethodGet:
		color = blue
	case http.MethodPost:
		color = green
	case http.MethodPut:
		color = yellow
	case http.MethodDelete:
		color = red
	case http.MethodPatch:
		color = magenta
	case http.MethodHead:
		color = gray
	case http.MethodOptions:
		color = cyan
	case http.MethodTrace:
		color = white
	default:
		color = black
	}

	return fmt.Sprintf("[%s%-7s%s]", color, method, reset)
}

func colorizeStatus(status int) string {
	if status < 100 || status > 599 {
		return fmt.Sprintf("[%s%3d%s]", red, status, reset)
	}

	var color string
	switch {
	case status >= http.StatusContinue && status < http.StatusOK:
		color = gray
	case status >= http.StatusOK && status < http.StatusMultipleChoices:
		color = green
	case status >= http.StatusMultipleChoices && status < http.StatusBadRequest:
		color = blue
	case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
		color = yellow
	default:
		color = red
	}

	return fmt.Sprintf("[%s%3d%s]", color, status, reset)
}

func colorizeDuration(duration time.Duration) string {
	if duration < 0 {
		return fmt.Sprintf("%s%13v%s", red, duration, reset)
	}

	var color string
	switch {
	case duration < 500*time.Millisecond:
		color = green
	case duration < 1*time.Second:
		color = yellow
	default:
		color = red
	}

	return fmt.Sprintf("%s%13v%s", color, duration, reset)
}

func LogRequest(reqLog *model.RequestLog) {
	// This function logs the request details based on the configuration
	if reqLog == nil {
		fmt.Println("Warning: Attempted to log nil request log, ignoring")
		return
	}

	fmt.Printf(
		"[VIS] %v %s%s %s %#v\n",
		reqLog.Timestamp.Format("2006-01-02 15:04:05"),
		colorizeMethod(reqLog.Method),
		colorizeStatus(reqLog.StatusCode),
		colorizeDuration(time.Since(reqLog.Timestamp)),
		reqLog.Path,
	)
}

func init() {
	// Initialize any necessary configurations or settings

	if !isatty.IsTerminal(os.Stdout.Fd()) || !isatty.IsTerminal(os.Stderr.Fd()) {
		green = ""
		white = ""
		red = ""
		blue = ""
		yellow = ""
		gray = ""
		black = ""
		magenta = ""
		cyan = ""
		reset = ""
	}
}
