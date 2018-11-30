# heartbeat-shell-monitor

## How to make the heartbeat ?
- Install golang 1.7+
- Pull the https://github.com/elastic/beats
- Copy the heartbeat-shell-monitor src to beats code.
- Go to github.com/elastic/beats/heartbeat/monitors/defaults/default.go, update its to 
```golang
package defaults

import (
	_ "github.com/elastic/beats/heartbeat/monitors/active/http"
	_ "github.com/elastic/beats/heartbeat/monitors/active/icmp"
	_ "github.com/elastic/beats/heartbeat/monitors/active/shell"
	_ "github.com/elastic/beats/heartbeat/monitors/active/tcp"
)

``` 
- Go to github.com/elastic/beats/heartbeat, run the following command
```
 export GOOS="linux"
 make
```

The shell type provides an Client interface , you can extend it for supporting more client types. 

```golang
type Client interface {
	Connect() error
	Reconnect() error
	Close()
	Run(dir, command string, args ...string) (string, error)
}

```
