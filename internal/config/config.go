package config

// Config defines the options for gomarklint, typically loaded from a config file.
type Config struct {
	MinHeadingLevel             int      `json:"minHeadingLevel"`
	EnableLinkCheck             bool     `json:"enableLinkCheck"`
	SkipLinkPatterns            []string `json:"skipLinkPatterns"`
	Ignore                      []string `json:"ignore"`
	OutputFormat                string   `json:"output"`
	EnableDuplicateHeadingCheck bool     `json:"enableDuplicateHeadingCheck"`
}

func Default() Config {
	return Config{
		MinHeadingLevel:             2,
		EnableLinkCheck:             false,
		SkipLinkPatterns:            []string{},
		Ignore:                      []string{},
		OutputFormat:                "text",
		EnableDuplicateHeadingCheck: true,
	}
}
