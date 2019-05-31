![Spiders](https://github.com/therahulprasad/spiders/blob/master/assets/banner.png)

# spiders
An easy to use web crawler for collecting text.  
[![CircleCI](https://circleci.com/gh/therahulprasad/spiders.svg?style=svg)](https://circleci.com/gh/therahulprasad/spiders)  

### Features
1. Indefinitely crawls till end condition is met
2. Configurable Concurrency  
3. SQLite based implementation for easy manual stats
4. Support for resuming broken operation
5. Configurable Regex based URL validation  
_You can decide which URL should be added to Queue based on custom Regex_
6. Configurable Selector based page validation  
_You can decide which part of page to be scrapped based on CSS selector_
7. Configurable Regex based URL Sanitizer  
_For example you can remove everything after #_
8. Batch URL processing  

### Installation
#### If you do not have Go installed (Recommended)  
- Download latest binaries from [release page](https://github.com/therahulprasad/spiders/releases)  
- Copy binaries to executable PATH or run directly from terminal using `./spiders`  

#### If you have Go installed  
`go install github.com/therahulprasad/spiders`  
and run `spiders` from terminal  

#### If you are a windows users  
- Upgrade to Linux  

### Usage
For help use  
`./spiders -h`

Create a `config.yaml` file and run  
`./spiders`

For using config which is not present in current directory use  
`./spiders -config /path/to/config.yaml`

Resume previous project by running   
`./spiders --resume`  

### Customization
Use self explanatory `config.yaml` to configure the project.

### What next?  
- Do not download non html resourcees  
- support link which starts with *//*www.example.com  
- Create a UI  
- Save config data in sqlite and implement --resume with db path instead of config path, let user override parameters using CLI arguments  
- Add new project type for fetching pagniated API data  
- Handle case: When craweling is complete.  
- Add support for parsing set of specified tags and collecting data in json format  
- Automated release on Tag using CircleCI

### Bugs
`Ctrl + C` does not work when workers are less

##### Change Log
_v0.1_  
Initial Release  
_v0.2_  
_v0.3_  
Batch Processing  
_v0.3.1_  
Config is made mandatory flag
Add two parameters in config to decide how to extract `text content_holder` (text/attr) and `content_tag_attr`  
_v.04_  
batch support  
attr grabbing support  
Configuration format updated from json to yaml  
-  encode.php not needed now  

