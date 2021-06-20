[![GoDoc](http://godoc.org/github.com/hanagantig/aerrors?status.png)](http://godoc.org/github.com/hanagantig/aerrors)
[![Build Status](https://travis-ci.org/robfig/cron.svg?branch=master)](https://travis-ci.com/hanagantig/aerrors)

# aerrors

Package aerrors adds **async errors** handling support in GO

This is effective when you want to recover panics in all your goroutines, build an error from it and handle all such errors in one place (logging them or monitor).

There is additionally the `Wrap()` method, which helps you to build manually 'stacktrace' with errors chain. 

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

You also can implement your custom handler. 
Merely use available option aerrors.WithHandler()

```go
package main

import (
    "github.com/hanagantig/aerrors"
)

type CustomErrorHandler struct {}
func (eh *CustomErrorHandler) HandleError(err error)  {
    // do what you want with your error here
}

func main() {
    h := CustomErrorHandler{}
    aerror := aerrors.New(aerrors.WithHandler(&h))
    _ = aerror.StartHandle()
    
    aerror.Close() // for graceful shutdown
}
```

In addition, you are capable to build your own stack trace with needed functions contains all desired information.
```go
package main

import (
    "fmt"
    "errors"
    "github.com/hanagantig/aerrors"
)

func foo(id int) (err error){
    defer aerrors.Wrap(&err, "foo(%v)", id)

    err = errors.New("errors in foo")
    return
}

func bar(id int, tag string) (err error)  {
    defer aerrors.Wrap(&err, "bar(%v, %v)", id, tag)
    
    err = foo(id)
    return
}

func main() {
    err := bar(1, "aerrors_wrap")
    fmt.Printf("%v", err)   
}

// output
// bar(1, aerrors_wrap): foo(1): errors in foo
``` 