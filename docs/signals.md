# signal.go

## func SetupSignalHandler
```go
func SetupSignalHandler() (stopCh <-chan struct{})
```
`SetupSignalHandler` registered for `SIGTERM` and `SIGINT`. A stop channel is returned which is closed on one of these signals. If a second signal is caught, the program is terminated with exit code 1.
