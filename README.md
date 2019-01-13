# chirp

[![Travis Build](https://travis-ci.com/vporoshok/chirp.svg?branch=master)](https://travis-ci.com/vporoshok/chirp)
[![Go Report Card](https://goreportcard.com/badge/github.com/vporoshok/chirp)](https://goreportcard.com/report/github.com/vporoshok/chirp)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/vporoshok/chirp)
[![codecov](https://codecov.io/gh/vporoshok/chirp/branch/master/graph/badge.svg)](https://codecov.io/gh/vporoshok/chirp)
[![MIT License](https://img.shields.io/github/license/mashape/apistatus.svg)](LICENSE)

Request parser for chi router

```
go get -u github.com/vporoshok/chirp
```

## Usage

There are two way to use this library. Parse as function or as middleware. Supported tags are `path` for chi router path parts, `query` for query parameters and `json` for body regardless of type (json / form). For json used standard `encoding/json`, for other parts it is trying to cast `encoding.TextUnmarshaler` and use it on success, otherwise used `fmt.Sscan`.

### Parse

```go
router := chi.NewRouter()
router.Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
    var data struct {
        ID       uuid.UUID `path:"id"`
        Name     string    `json:"name"`
        Part     string    `query:"part"`
        Priority uint8     `json:"priority"`
        Null     string    `json:"-"`
        Hero     string
    }
    if err := chirp.Parse(r, &data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // work with data
})
```

### Middleware

Middleware need a struct instance, not point to struct!

```go
router := chi.NewRouter()
type Data struct {
    ID       uuid.UUID `path:"id"`
    Name     string    `json:"name"`
    Part     string    `query:"part"`
    Priority uint8     `json:"priority"`
    Null     string    `json:"-"`
    Hero     string
}
router.With(
    chirp.Middleware(Data{}, chirp.WithInterrupt(400)),
).Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
    data := chirp.RequestFromContext(r.Context()).(Data)
    // use data
})
```

Be aware to use middleware only on part of route which contained all needed parts. For example, next code is wrong:
```go
router := chi.NewRouter()
type Data struct {
    ID       uuid.UUID `path:"id"`
    Name     string    `json:"name"`
    Part     string    `query:"part"`
    Priority uint8     `json:"priority"`
    Null     string    `json:"-"`
    Hero     string
}
router.Use(chirp.Middleware(Data{}, chirp.WithInterrupt(400)))
router.Put("/user/{id}/name", func(w http.ResponseWriter, r *http.Request) {
    data := chirp.RequestFromContext(r.Context()).(Data)
    // data.ID always be empty
})
```
