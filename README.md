[![GoDoc](http://godoc.org/github.com/hanagantig/aerrors?status.png)](http://godoc.org/github.com/hanagantig/aerrors)
[![Build Status](https://travis-ci.org/robfig/cron.svg?branch=master)](https://travis-ci.com/hanagantig/aerrors)

# aerrors

Package aerrors adds **async errors handling** support in GO

This is useful when you want to recover panics in all your goroutines, build an error from it and handle all such errors in one place (logging them or monitor).

To download the package, run:
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

## Usage
```go
package main

import (
    "github.com/hanagantig/aerrors"
)

func crashFunc()  {
    panic("crashFunc panic")
}

func main() {
    aerror := aerrors.New()
    aerror.StartHandle()
    
    server := runHTTP()
    
    // runs crashFunc in panic-safe goroutine and adds error to handle
    aerror.Go(crashFunc)

    server.Stop()
    aerror.Stop()
}
```

You also can use it with your custom handler. 
Just use available option aerrors.WithHandler()

```go
package main

import (
    "github.com/hanagantig/aerrors"
)

type CustomErrorHandler struct {}
func (eh *CustomErrorHandler) Handle(err error)  {
    // do what you want with your error here
}

func main() {
    h := CustomErrorHandler{}
    aerror := aerrors.New(aerrors.WithHandler(&h))
    aerror.StartHandle()
    
    aerror.Stop()
}
```