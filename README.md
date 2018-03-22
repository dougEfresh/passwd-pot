# Passwdpot - A genric password honey api to store brute-force password attacks

Passwdpot is both a server and client application to capture passwords from nodes running passwd-pot pottter app

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report][report-img]][report]

## Installation 
```shell
$ go get -u github.com/dougEfresh/passwdpot
```

## Quick Start
 
 ```go
package main

import (
  "os"

  "github.com/dougEfresh/zapz"
    )

func main() {
  l, err := zapz.New(os.Args[1]) //logzio token required
  if err != nil {
    panic(err)
  }

  l.Info("tester")
  // Logs are buffered on disk, this will flush it
  if l.Sync() != nil {
      panic("oops")
  }
}
```


Sends a password event to server

## Usage
    
TBD


## Examples
    



## Prerequisites

go 1.x

## Tests
    
```shell
$ go test -v

```

## Contributing
 All PRs are welcome

## Authors

* **Douglas Chimento**  - [dougEfresh][me]

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

  [logz java](https://github.com/logzio/logzio-java-sender)

### TODO 

[doc-img]: https://godoc.org/github.com/dougEfresh/passwd-pot?status.svg
[doc]: https://godoc.org/github.com/dougEfresh/passwd-pot
[ci-img]: https://travis-ci.org/dougEfresh/passwd-pot.svg?branch=master
[ci]: https://travis-ci.org/dougEfresh/passwd-pot
[cov-img]: https://codecov.io/gh/dougEfresh/passwd-pot/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/dougEfresh/passwd-pot
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
[zap]: https://github.com/uber-go/zap
[me]: https://github.com/dougEfresh
[report-img]: https://goreportcard.com/badge/github.com/dougEfresh/passwd-pot
[report]: https://goreportcard.com/report/github.com/dougEfresh/passwd-pot