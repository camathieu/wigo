package utils

func StatusCodeToString(status int) string {
	if status == 100 {
		return "OK"
	} else if status > 100 && status < 200 {
		return "INFO"
	} else if status >= 200 && status < 300 {
		return "WARN"
	} else if status >= 300 && status < 500 {
		return "CRIT"
	} else if status >= 500 {
		return "ERROR"
	}
	return "ERROR"
}