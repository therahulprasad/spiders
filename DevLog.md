# Development Log

#### Todo 
- On resume, load configuration from database
- Configurable Regex based page Sanitizer
- step by step config creator
- __Validate URL > mailto: and other protocols__
- Better way to manage database
- Make SQLITE faster
    - Pop multiple items per worker
    - Insert multiple links at once
- Put some example config files
- UI
- Sleep time in configuration
- Proxy and VPN configurations
- Think of a way to make the spiders roam freely and collect specified character-set based data from around the internet

### Done
- Maintain md5 hash of pages to check redundancy
- Validate_URL > Relative URL ../../  bug
- moved Ctrl + C handler to main.go from spider.go
- Find better alternative for storing config (Using Yaml now)
- Resume support 
- Serial mode (To download pages with numeric ids)
- Handle database locked error
