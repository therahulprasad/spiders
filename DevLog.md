# Development Log

#### Todo 
- Serial mode (To download pages with numeric ids)
- On resume, load configuration from database
- Configurable Regex based page Sanitizer
- step by step config creator
- __Validate URL > mailto: and other protocols__
- Handle database locked error
- Better way to manage database
- Make SQLITE faster
    - Pop multiple items per worker
    - Insert multiple links at once

### Done
- Maintain md5 hash of pages to check redundancy
- Validate_URL > Relative URL ../../  bug
- moved Ctrl + C handler to main.go from spider.go
- Find better alternative for storing config (Using Yaml now)
- Resume support