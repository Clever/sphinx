# v0.7.1: Add AWS DynamoDB as a backend

# v0.7.0: Add global limit key
  - Allows global rate limiting irrespective of request content

# v0.6.1: log routing + go 1.7

# v0.6.0: Standardized request logging, X-Request-Id header

# v0.5.3: Report card + go 1.6

# v0.5.2: Move to structured logging using Kayvee

# v0.2.3: Change release process

# v0.2.2: Add resapwn to upstart config
  - Add respawn to upstart config

# v0.2.1: Allow reloading of config file, fix race
  - Send SIGHUP to reload the config file
  - Upgrade leakybucket version to fix redis race condition

# v0.2.0: Handle multiple instances and sensitive headers
  - Sort headers for more consistent bucketnames
  - Allow hashing headers for increased security

# v0.1.1: First release for real scalability testing
 - Fixes after some real world testing to logging and stability

# v0.1.0: Initial Release
  - Supports http and httplogger handlers
  - Support header and path based request matching
  - Support keying requests based on headers and request ip
  - Tests and lint for ALL the things!

# v0.0.1: Pre Release
  - This one is just a test
