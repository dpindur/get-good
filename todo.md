# Todo

A list of features which would be nice to implement:

* Refactor http worker to handle response channel being blocked
* Add backoff if multiple requests fail
* Configurable request delays and related timers
* Configurable recursion
* Configurable 'successful request' criteria
* Save successful responses to the database
* Add ability to connect to alternate database
* More graceful exit if http workers timeout (can end up waiting a while for halt signal to be handled, maybe just reduce the timeout?)
* Tune performance (batch database writes? find optimal queue and poller sizes?)
* Add alternative (i.e. short) names for flags
* Add support for proxy / Auth
* Add tests
* Add comments for exported functions/variables
* Add detection for soft 404 (i.e. server responding with 200 for everything)
* Detect false positives when recursively scanning files. (i.e. requesting url/info.php will produce same result as requesting url/info.php/{anything here})
* Refactor parsing of flags to use a config struct
* Color log level indicator in terminal output
* Bring back SIGINT handling