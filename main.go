package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	lib "github.com/dpindur/get-good/libgetgood"
	. "github.com/dpindur/get-good/logger"
	_ "github.com/mattn/go-sqlite3"
	logrus "github.com/sirupsen/logrus"
)

func main() {
	workerCount := flag.Int("workers", 5, "number of worker threads")
	clearDB := flag.Bool("clear-db", false, "clear the database before starting")
	dbFile := flag.String("db", "bust.db", "database file to store results")
	logFileStr := flag.String("log-file", "bust.log", "log file to output progress to")
	logLevelStr := flag.String("log-level", "info", "what level of logs and up should be logged (debug, info, warn, error, fatal, panic)")
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

	// Logging
	logLevel, err := logrus.ParseLevel(strings.ToLower(*logLevelStr))
	if err != nil {
		fmt.Printf("not a valid log level %v\n", *logLevelStr)
		flagsInvalid = true
	}

	logFile, err := os.OpenFile(*logFileStr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		fmt.Printf("error opening logfile %v\n", *logFileStr)
		flagsInvalid = true
	}

	if flagsInvalid {
		os.Exit(1)
	}

	ConfigureLogger(logLevel, logFile)

	Logger.Infof("Starting get-good directory bust of %v", *urlStr)
	Logger.Infof("Worker threads: %v", *workerCount)
	Logger.Infof("Database file: %v", dbFilePath)
	Logger.Infof("Wordlist file: %v", wordsFilePath)
	Logger.Infof("Extensions: (blank)%v", strings.Join(extensions, ", "))
	Logger.Infof("Resuming existing directory bust: %v", !*clearDB)
	Logger.Infof("Logging to file: %v", *logFileStr)
	Logger.Infof("Configured logging level: %v", *logLevelStr)

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	cleanupChan := make(chan struct{})
	go func() {
		<-signalChan
		fmt.Printf("received an interrupt, stopping...\n")
		close(cleanupChan)
	}()
	<-cleanupChan

	updater.Stop()
	for _, worker := range workers {
		worker.Stop()
	}
	poller.Stop()

	if workerErr == nil {
		wg.Wait()
	} else {
		fmt.Printf("terminating without properly halting routines... sorry\n")
	}
}
