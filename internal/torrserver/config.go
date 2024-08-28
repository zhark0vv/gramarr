package torrserver

type Config struct {
	Hostname    string  `json:"hostname"`
	Port        int     `json:"port"`
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	TrackerHost *string `json:"trackerHost"`
	TrackerPort *int    `json:"trackerPort"`
}
