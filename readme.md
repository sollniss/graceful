# graceful
Enables graceful shutdown of pretty much anything you can imagine.
All you need is a blocking start function that returns a known error on successful shutdown, and a shutdown function.

Includes utility functions to handle HTTP servers in a one-liner.
Also includes a helper function to handle most common OS process interrupt signals.

# Examples
```go
func main() {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("start Hello World")
		time.Sleep(10 * time.Second)
		w.Write([]byte("Hello World"))
		log.Print("finish Hello World")
	})
	s := &http.Server{
		Handler: m,
		Addr:    "0.0.0.0:8080",
	}

	ctx := graceful.NotifyShutdown()
	err := graceful.ListenAndServe(ctx, s, 60*time.Second)
	if err != nil {
		log.Printf("error during shutdown: %s", err)
		return
	}
	log.Print("gracefully shut down")
}
```

Start the server, open localhost:8080 in the browser and then terminate the process.
Check the fibonacci example in the examples directory for a sample of how to use the customizable Start function.