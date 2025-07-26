package config

// Config defines the options for gomarklint, typically loaded from a config file.
type Config struct {
	MinHeadingLevel  int      `json:"minHeadingLevel"`
	CheckLinks       bool     `json:"checkLinks"`
	SkipLinkPatterns []string `json:"skipLinkPatterns"`
	Ignore           []string `json:"ignore"`
	OutputFormat     string   `json:"output"`
}

func Default() Config {
	return Config{
		MinHeadingLevel:  2,
		CheckLinks:       false,
		SkipLinkPatterns: []string{},
		Ignore:           []string{},
		OutputFormat:     "text",
	}
}
