# Todo
A list of features which would be nice to implement

* Reset inflight requests on resumption so they are not missed
* Graceful exit when no more requests remain
* Refactor http worker to handle response channel being blocked
* Add backoff if multiple requests fail
* Configurable request delays and related timers
* Configurable logging options
* Configurable recursion
* Configurable 'successful request' criteria
* Save successful responses to the database
* Ability to print status to command line
* Add ability to connect to alternate database