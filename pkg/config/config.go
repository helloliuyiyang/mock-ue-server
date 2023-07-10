package config

type Config struct {
	Db            string
	ServerNames   []string
	UserCount     int
	TplFilePath   string
	IntervalMs    int
	PrintDebugLog bool
}
