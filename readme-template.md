# {{ .title.name }}

{{ .title.description }}

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report][report-img]][report]

## Installation 
{{ .installation }}

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
{{ .quickStart.code}}

{{ .quickStart.description }}

## Getting Started

### Get Logzio token
1. Go to Logzio website
2. Sign in with your Logzio account
3. Click the top menu gear icon (Account)
4. The Logzio token is given in the account page

## Usage 

{{- range .usages }}
    
{{.}}
    
{{- end }}


## Examples 

{{- range .examples }}
    
{{.}}
    
{{- end }}


## Prerequisites

go 1.x

## Tests 

{{- range .tests }}
    
{{.}}
    
{{- end }}

## Contributing
 All PRs are welcome

## Authors

* **Douglas Chimento**  - [{{.user}}][me]

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

  [logz java](https://github.com/logzio/logzio-java-sender)

### TODO 

[doc-img]: https://godoc.org/github.com/{{.user}}/{{.project}}?status.svg
[doc]: https://godoc.org/github.com/{{.user}}/{{.project}}
[ci-img]: https://travis-ci.org/{{.user}}/{{.project}}.svg?branch=master
[ci]: https://travis-ci.org/{{.user}}/{{.project}}
[cov-img]: https://codecov.io/gh/{{.user}}/{{.project}}/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/{{.user}}/{{.project}}
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
[zap]: https://github.com/uber-go/zap
[me]: https://github.com/{{.user}}
[report-img]: https://goreportcard.com/badge/github.com/{{.user}}/{{.project}}
[report]: https://goreportcard.com/report/github.com/{{.user}}/{{.project}}