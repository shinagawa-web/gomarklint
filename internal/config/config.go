package config

// Config defines the options for gomarklint, typically loaded from a config file.
type Config struct {
	MinHeadingLevel  int      `json:"minHeadingLevel"`
	CheckLinks       bool     `json:"checkLinks"`
	SkipLinkPatterns []string `json:"skipLinkPatterns"`
}

func Default() Config {
	return Config{
		MinHeadingLevel:  2,
		CheckLinks:       false,
		SkipLinkPatterns: []string{},
	}
}
