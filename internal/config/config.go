package config

// Config defines the options for gomarklint, typically loaded from a config file.
type Config struct {
	MinHeadingLevel                 int      `json:"minHeadingLevel"`
	EnableLinkCheck                 bool     `json:"enableLinkCheck"`
	SkipLinkPatterns                []string `json:"skipLinkPatterns"`
	Include                         []string `json:"include"`
	Ignore                          []string `json:"ignore"`
	OutputFormat                    string   `json:"output"`
	EnableDuplicateHeadingCheck     bool     `json:"enableDuplicateHeadingCheck"`
	EnableHeadingLevelCheck         bool     `json:"enableHeadingLevelCheck"`
	EnableNoMultipleBlankLinesCheck bool     `json:"enableNoMultipleBlankLinesCheck"`
}

func Default() Config {
	return Config{
		MinHeadingLevel:                 2,
		EnableLinkCheck:                 false,
		SkipLinkPatterns:                []string{},
		Include:                         []string{"README.md", "testdata"},
		Ignore:                          []string{},
		OutputFormat:                    "text",
		EnableDuplicateHeadingCheck:     true,
		EnableHeadingLevelCheck:         true,
		EnableNoMultipleBlankLinesCheck: true,
	}
}
