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
Usage of get-good:
  -clear-db
        clear the database before starting
  -db <filename>
        database file to store results (default "bust.db")
  -extensions <extensions>
        comma separated list of extensions to append (default "html,php")
  -url <url>
        url to perform directory bust against
  -wordlist <filename>
        wordlist file to use
  -workers <number of workers>
        number of worker threads (default 5)
```

Enter 'q' to halt directory busting. Any in-flight requests will be completed before exiting.

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