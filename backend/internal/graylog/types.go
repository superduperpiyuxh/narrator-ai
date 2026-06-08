package graylog

type Event struct {
	Timestamp   string                 `json:"timestamp"`
	Hostname    string                 `json:"hostname"`
	EventType   string                 `json:"event_type"`
	EventID     string                 `json:"event_id"`
	User        string                 `json:"user"`
	SourceIP    string                 `json:"source_ip"`
	DestIP      string                 `json:"dest_ip"`
	ProcessName string                 `json:"process_name"`
	CommandLine string                 `json:"command_line"`
	ParentProcess string               `json:"parent_process"`
	LogType     string                 `json:"log_type"`
	SessionID   string                 `json:"session_id"`
	Department  string                 `json:"department"`
	Location    string                 `json:"location"`
	DeviceType  string                 `json:"device_type"`
	Success     bool                   `json:"success"`
	Port        string                 `json:"port"`
	Protocol    string                 `json:"protocol"`
	FilePath    string                 `json:"file_path"`
	Severity    string                 `json:"severity"`
	Error       string                 `json:"error"`
	RawJSON     map[string]interface{} `json:"raw_json"`
}
