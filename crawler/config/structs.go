package config

type Configuration struct {
	Debug bool `json:"debug"`
	Project string `json:"project"`
	WebCount int `json:"web_count"`
	RootURL string `json:"root_url"`
	RootURLTest string `json:"root_url_test"`
	Directory string `json:"directory"`
	LinkValidator string `json:"link_validator"`
	LinkSanitizer string `json:"link_sanitizer"`
	LinkSanitizerReplacement string `json:"link_sanitizer_replacement"`
	ContentSelector string `json:"content_selector"`
	PageValidator string `json:"page_validator"`
}

func (c *Configuration) DataDir() string {
	return c.Directory + "/data"
}
