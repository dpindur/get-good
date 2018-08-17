package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	lib "github.com/dpindur/get-good/libgetgood"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	workerCount := flag.Int("workers", 5, "number of worker threads")
	clearDB := flag.Bool("clear-db", false, "clear the database before starting")
	dbFile := flag.String("db", "bust.db", "database file to store results")
	urlStr := flag.String("url", "", "url to perform directory bust against")
	wordsFile := flag.String("wordlist", "", "wordlist file to use")
	extensionsFlag := flag.String("extensions", "html,php", "comma separated list of extensions to append")

	flag.Parse()
	flagsInvalid := false

	// Workers
	if *workerCount < 1 {
		fmt.Println("please specify 1 or more worker threads")
		flagsInvalid = true
	}

	// Database File
	if !strings.HasSuffix(*dbFile, ".db") {
		*dbFile += ".db"
	}
	dbFilePath, err := filepath.Abs(*dbFile)
	if err != nil {
		fmt.Println("error resolving path %v\n", *dbFile)
		flagsInvalid = true
	}

	// Url
	urlProvided := true
	if *urlStr == "" {
		fmt.Println("please provide a URL to perform the directory bust against")
		flagsInvalid = true
		urlProvided = false
	}
	if !strings.HasSuffix(*urlStr, "/") {
		*urlStr += "/"
	}
	_, err = url.ParseRequestURI(*urlStr)
	if err != nil && urlProvided {
		fmt.Println("error parsing url, please ensure it includes the protocol for example http://google.com/")
		flagsInvalid = true
	}

	// Wordlist
	if *wordsFile == "" {
		fmt.Println("please provide a wordlist file")
		flagsInvalid = true
	}
	wordsFilePath, err := filepath.Abs(*wordsFile)
	if err != nil {
		fmt.Println("error resolving path %v\n", *wordsFile)
		flagsInvalid = true
	}

	// Extensions
	splitExtensions := strings.Split(*extensionsFlag, ",")
	extensions := make([]string, 0)
	extensions = append(extensions, "")
	for _, ext := range splitExtensions {
		if !strings.HasPrefix(ext, ".") {
			extensions = append(extensions, "."+ext)
		} else {
			extensions = append(extensions, ext)
		}
	}

	if flagsInvalid {
		os.Exit(1)
	}

	log.Printf("Starting get-good directory bust of %v\n", *urlStr)
	log.Printf("Worker threads: %v\n", *workerCount)
	log.Printf("Database file: %v\n", dbFilePath)
	log.Printf("Wordlist file: %v\n", wordsFilePath)
	log.Printf("Extensions: (blank)%v\n", strings.Join(extensions, ", "))
	log.Printf("Resuming existing directory bust: %v\n", !*clearDB)

	db, err := lib.OpenDatabaseConnection(dbFilePath)
	if err != nil {
		fmt.Println("error opening database connection")
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	defer func() {
		err = db.CloseDatabaseConnection()
		if err != nil {
			fmt.Printf("error closing database connection")
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}()

	err = db.CreateSchema()
	if err != nil {
		fmt.Println("error creating database schema")
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if *clearDB {
		err = db.Clear()
		if err != nil {
			fmt.Println("error clearing database")
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}

	// Process the wordlist
	words := make([]string, 0)
	wordlist, err := os.Open(wordsFilePath)
	if err != nil {
		fmt.Println("error opening wordlist file")
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(wordlist)
	for scanner.Scan() {
		word := scanner.Text()
		words = append(words, word)
	}
	err = scanner.Err()
	if err != nil {
		fmt.Println("error reading wordlist file")
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Handler for worker errors
	errChan := make(chan *lib.WorkerError)
	var workerErr *lib.WorkerError
	go func() {
		workerErr = <-errChan
		fmt.Printf("error in worker routine: %v\n", workerErr.Worker)
		fmt.Printf("%v\n", workerErr.Error)
		fmt.Printf("enter any key to continue\n")
	}()

	// Start database workers
	wg := &sync.WaitGroup{}
	requestChan := make(chan *lib.Request, 100)
	responseChan := make(chan *lib.Response, 100)
	updater := lib.StartUpdater(wg, db, errChan, responseChan, words, extensions)
	poller := lib.StartPoller(wg, db, errChan, requestChan)

	// Start http workers
	workers := make([]*lib.HttpWorker, 0)
	for i := 0; i < *workerCount; i++ {
		worker := lib.StartHttpWorker(wg, db, requestChan, responseChan)
		workers = append(workers, worker)
	}

	// Enqueue initial request
	updater.EnqueueRequest(&lib.Request{*urlStr})

	reader := bufio.NewReader(os.Stdin)
	running := true
	for running {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading from stdin: %v\n", err)
			running = false
			break
		}

		// Don't bother processing commands if a worker has failed and
		// cannot be recovered
		if workerErr != nil {
			running = false
			break
		}

		cmd = strings.TrimSuffix(cmd, "\n")
		switch cmd {
		case "q":
			running = false
			updater.Stop()
			for _, worker := range workers {
				worker.Stop()
			}
			poller.Stop()
			break
		}
	}

	if workerErr == nil {
		wg.Wait()
	} else {
		fmt.Printf("terminating without properly halting routines... sorry\n")
	}
}
