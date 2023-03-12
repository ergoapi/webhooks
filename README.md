# Library webhooks

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ergoapi/webhooks?filename=go.mod&style=flat-square)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/ergoapi/webhooks?style=flat-square)
![GitHub](https://img.shields.io/github/license/ergoapi/webhooks?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/ergoapi/webhooks)](https://goreportcard.com/report/ergoapi/webhooks)
[![Goproxy.cn](https://goproxy.cn/stats/github.com/ergoapi/webhooks/badges/download-count.svg)](https://goproxy.cn)

Library webhooks allows for easy receiving and parsing of Gitea, GitHub and GitLab Webhook Events

## Installation

```bash
go get -u github.com/ergoapi/webhooks
```

import package into your code

```go
import "github.com/ergoapi/webhooks"
```

## Usage

```go
package main

import (
 "fmt"

 "net/http"

 "github.com/ergoapi/webhooks/gitea"
)

const (
 path = "/webhooks"
)

func main() {
 hook, _ := gitea.New(gitea.Options.Secret("MyGiteaSuperSecretSecrect...?"))
 http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
  payload, err := hook.Parse(r, github.ReleaseEvent, gitea.PullRequestEvent)
  if err != nil {
   if err == gitea.ErrEventNotFound {
    // ok event wasn;t one of the ones asked to be parsed
   }
  }
  switch payload.(type) {

  case gitea.ReleasePayload:
   release := payload.(gitea.ReleasePayload)
   // Do whatever you want from here...
   fmt.Printf("%+v", release)

  case gitea.PullRequestPayload:
   pullRequest := payload.(gitea.PullRequestPayload)
   // Do whatever you want from here...
   fmt.Printf("%+v", pullRequest)
  }
 })
 http.ListenAndServe(":3000", nil)
}
```

## Support

- [x] gitea 1.18.5
- [ ] gogs
- [ ] github
- [ ] gitlab

