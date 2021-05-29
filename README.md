[![GoDoc](http://godoc.org/github.com/hanagantig/aerrors?status.png)](http://godoc.org/github.com/hanagantig/aerrors)
[![Build Status](https://travis-ci.org/robfig/cron.svg?branch=master)](https://travis-ci.com/hanagantig/aerrors)

# aerrors

Package aerrors adds **async errors handling** support in GO

This is useful when you want to recover panics in all your goroutines, build an error from it and handle all such errors in one place (logging them or monitor).

To download the specific tagged release, run:
```bash
go get github.com/hanagantig/aerrors
```

Import it in your program as:
```go
import "github.com/hanagantig/aerrors"
```

It requires Go 1.13 or later.

Refer to the documentation here:
http://godoc.org/github.com/hanagantig/aerrors
