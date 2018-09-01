# Get Good
A simple, database-backed HTTP directory buster written in Golang

## Why?
Too many directory busters are either non-recursive or crash often and
cannot be resumed easily when they do.

Get Good offers the following
* Saves state to SQLite for easy query and resumption
* Performs recursive directory searching
* Runs concurrently
* Compiles to a single binary with no runtime

## Usage
```
Usage of ./get-good:
  -clear-db
    	clear the database before starting
  -db string
    	database file to store results (default "bust.db")
  -extensions string
    	comma separated list of extensions to append (default "html,php")
  -log-file string
    	log file to output progress to (default "bust.log")
  -log-level string
    	what level of logs and up should be logged (debug, info, warn, error, fatal, panic) (default "info")
  -poller-batch-size int
    	number of urls the poller can pull from the database in one go (default 5000)
  -queue-size int
    	number of urls that can sit in the queue at one time (default 5000)
  -timeout int
    	http timeout in seconds, specify zero for no timeout (default 10)
  -url string
    	url to perform directory bust against
  -wordlist string
    	wordlist file to use
  -workers int
    	number of worker threads (default 5)
```

Press `q` to halt directory busting. Any in-flight requests will be completed before exiting.

## Examples

### Resuming
```
get-good --db existing-directory-bust.db --url http://localhost --wordlist words.txt
```

### Different extensions
```
get-good --url http://localhost --wordlist words.txt --extensions txt,bak,zip
```

### Running with extra HTTP worker threads
```
get-good --url http://localhost --wordlist words.txt --workers 10
```