package logger

type LogEvent struct {
	Timestamp string `json:"ts_orig"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
	Hash      string `json:"hash"`
}
