
Unofficial golang library for LogDNA.


# Quick Usage

```go
import (
	"github.com/t0mk/logdna"
)

func someFunction() {
	conf := logdna.Config{
		APIKey:       "",
		App:          "myapp",
		Env:          "production",
		MinimumLevel: logdna.LogLevelInfo,
		Sync:         false,
		Debug:        true,
	}

	cli, err := logdna.New(conf)
	if err != nil {
		panic(err)
	}

	cli.Debug("logging...")
	cli.Trace("logging...")
	cli.Info("logging...")
	cli.Warn("logging...")
	cli.Err("logging...")
	cli.Fatal("logging...")
}

```


# Environment variables

| Name | Description |
|:--|:--|
| `LOGDNA_API_KEY` | API Key of LogDNA. |
