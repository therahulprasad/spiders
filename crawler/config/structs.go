package config

// Configuration holds the difinition of configuration which will be parsed from config.yaml
type Configuration struct {
	Debug                    bool   `yaml:"debug"`
	DisplayMatchedURL        bool   `yaml:"display_matched_url"`
	Project                  string `yaml:"project"`
	ProjectType              string `yaml:"project_type"`
	WebCount                 int    `yaml:"web_count"`
	RootURL                  string `yaml:"root_url"`
	RootURLTest              string `yaml:"root_url_test"`
	Directory                string `yaml:"directory"`
	LinkValidator            string `yaml:"link_validator"`
	LinkSanitizer            string `yaml:"link_sanitizer"`
	LinkSanitizerReplacement string `yaml:"link_sanitizer_replacement"`
	ContentSelector          string `yaml:"content_selector"`
	ContentHolder            string `yaml:"content_holder"`   // text,attr,html
	ContentTagAttr           string `yaml:"content_tag_attr"` // optional
	PageValidator            string `yaml:"page_validator"`
	BatchURL                 string `yaml:"batch_url"`
	ProxyAPI				 string `yaml:"proxy_api"`		  // optional
}

// DataDir returns a directory where data will be stored
// TODO: Make it configurable
func (c *Configuration) DataDir() string {
	return c.Directory + "/data"
}

// PROJECTTYPECRAWL crawl
const PROJECTTYPECRAWL = "crawl"

// PROJECTTYPEBACTH batch
const PROJECTTYPEBACTH = "batch"

// CONTENTHOLDERTEXT text
const CONTENTHOLDERTEXT = "text"

// CONTENTHOLDERATTR attr
const CONTENTHOLDERATTR = "attr"

// CONTENTHOLDERATTR html
const CONTENTHOLDERHTML = "html"