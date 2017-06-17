package config

type T struct {
	K bool `yaml:"debug"`
	B struct {
		  RenamedC int   `yaml:"c"`
		  D        []int `yaml:",flow"`
	  }
}
type Configuration struct {
	Debug bool `yaml:"debug"`
	DisplayMatchedUrl bool `yaml:"display_matched_url"`
	Project string `yaml:"project"`
	WebCount int `yaml:"web_count"`
	RootURL string `yaml:"root_url"`
	RootURLTest string `yaml:"root_url_test"`
	Directory string `yaml:"directory"`
	LinkValidator string `yaml:"link_validator"`
	LinkSanitizer string `yaml:"link_sanitizer"`
	LinkSanitizerReplacement string `yaml:"link_sanitizer_replacement"`
	ContentSelector string `yaml:"content_selector"`
	PageValidator string `yaml:"page_validator"`
}

func (c *Configuration) DataDir() string {
	return c.Directory + "/data"
}
