# settings
[![Build Status](https://travis-ci.org/winjeg/settings.svg?branch=master)](https://travis-ci.org/winjeg/settings)
[![Go Report Card](https://goreportcard.com/badge/github.com/winjeg/settings)](https://goreportcard.com/report/github.com/winjeg/settings)
[![codecov](https://codecov.io/gh/winjeg/settings/branch/master/graph/badge.svg)](https://codecov.io/gh/winjeg/settings)

a simple tool to manage settings for an go app with mysql database

## about settings
settings is a simple go library for users to
set get and delete variables from db easily

you don't even need to create the settings table,
the program will create the settings table automatically for you.

the program is born with cache enabled


## how to use
```
# go get github.com/winjeg/settings
```

```go
import (
	"github.com/winjeg/settings"
)

func TestSettings(){
    // the db connection can be acquired from your app
	dbConn := getDb()
    // Init before use, fix the error if necessary		
	err := settings.Init(dbConn)
	settings.SetVar("a", "b")
	settings.SetVar("a", "bc")
	settings.GetVar("a")
	settings.DelVar("a")
}

```

