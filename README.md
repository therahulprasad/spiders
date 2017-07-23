# spiders
An easily configurable and easy to use web crawler. 

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
8. Configurable Regex based page Sanitizer  
_For example if you want to keep only english content and remove all other characters_

### Installation
`go install github.com/therahulprasad/spiders`

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

##### Change Log
_v0.1 Initial Release_  
_v0.2_  
_v0.3 Batch Processing_

