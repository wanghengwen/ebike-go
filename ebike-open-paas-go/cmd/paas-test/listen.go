package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

func callbackListen(cfg Config, args []string) error {
	fs := flag.NewFlagSet("callback listen", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	listen := fs.String("listen", envOr("PAAS_LISTEN", ":8089"), "local bind address (HTTP; TLS terminates at the reverse proxy)")
	publicURL := fs.String("public-url", envOr("PAAS_CALLBACK_URL", ""), "public https URL registered with PaaS")
	eventsRaw := fs.String("events", "1,2,3,4", "comma-separated event types to register")
	timeout := fs.Duration("timeout", 0, "exit after this duration (0 = until Ctrl-C)")
	expect := fs.Int("expect", 0, "exit 0 after receiving this many callbacks (0 = ignore)")
	unregister := fs.Bool("unregister-on-exit", true, "DELETE the registered URL on exit")
	skipRegister := fs.Bool("skip-register", false, "only listen; do not call /callback")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := mustHTTPS(*publicURL); err != nil {
		return err
	}
	pub, err := url.Parse(strings.TrimSpace(*publicURL))
	if err != nil {
		return err
	}
	hookPath := pub.EscapedPath()
	if hookPath == "" {
		hookPath = "/"
	}

	events, err := parseEvents(*eventsRaw)
	if err != nil {
		return err
	}

	var received uint64
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})

	handleHook := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		n := atomic.AddUint64(&received, 1)
		eventID := r.Header.Get("X-Event-Id")
		attempt := r.Header.Get("X-Event-Attempt")
		fmt.Printf("\n── callback #%d  %s  eventId=%s attempt=%s\n",
			n, time.Now().Format(time.RFC3339), eventID, attempt)
		var pretty bytes.Buffer
		if json.Indent(&pretty, body, "", "  ") == nil {
			fmt.Println(pretty.String())
		} else {
			fmt.Println(string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}
	mux.HandleFunc(hookPath, handleHook)
	if hookPath != "/" {
		alt := strings.TrimSuffix(hookPath, "/")
		if alt == "" {
			alt = "/"
		}
		if alt != hookPath {
			mux.HandleFunc(alt, handleHook)
		} else {
			mux.HandleFunc(hookPath+"/", handleHook)
		}
	}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		return fmt.Errorf("listen %s: %w", *listen, err)
	}
	srv := &http.Server{Handler: loggingHandler(mux), ReadHeaderTimeout: 10 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("listening on http://%s%s  (healthz: /healthz)\n", ln.Addr(), hookPath)
		fmt.Printf("register this public URL with PaaS: %s\n", *publicURL)
		errCh <- srv.Serve(ln)
	}()

	var client *Client
	registered := false
	if !*skipRegister {
		if err := cfg.requireAuth(); err != nil {
			_ = srv.Close()
			return err
		}
		client = newClient(cfg)
		for _, ev := range events {
			env, status, err := client.post("/ebike/v1/callback", map[string]interface{}{
				"event": ev,
				"url":   *publicURL,
			})
			if err != nil {
				_ = srv.Close()
				return fmt.Errorf("register event %d: %w", ev, err)
			}
			if !env.Success {
				_ = srv.Close()
				return fmt.Errorf("register event %d failed http=%d: %s (%s)",
					ev, status, env.errorType(), env.promot())
			}
			fmt.Printf("registered event=%d url=%s data=%s\n", ev, *publicURL, string(env.Data))
		}
		registered = true
	}

	var deadline <-chan time.Time
	if *timeout > 0 {
		t := time.NewTimer(*timeout)
		defer t.Stop()
		deadline = t.C
		fmt.Printf("will stop after %s\n", *timeout)
	}

	var exitErr error
loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nshutting down…")
			break loop
		case <-deadline:
			fmt.Println("\ntimeout reached")
			if *expect > 0 && atomic.LoadUint64(&received) < uint64(*expect) {
				exitErr = fmt.Errorf("received %d callbacks, want at least %d",
					atomic.LoadUint64(&received), *expect)
			}
			break loop
		case err := <-errCh:
			if err != nil && err != http.ErrServerClosed {
				exitErr = err
			}
			break loop
		case <-time.After(200 * time.Millisecond):
			if *expect > 0 && atomic.LoadUint64(&received) >= uint64(*expect) {
				fmt.Printf("\nreceived %d callbacks, done\n", atomic.LoadUint64(&received))
				break loop
			}
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	if registered && *unregister && client != nil {
		for _, ev := range events {
			env, status, err := client.delete("/ebike/v1/callback", url.Values{
				"event": {strconv.Itoa(ev)},
				"url":   {*publicURL},
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "unregister event=%d: %v\n", ev, err)
				continue
			}
			if !env.Success {
				fmt.Fprintf(os.Stderr, "unregister event=%d http=%d: %s (%s)\n",
					ev, status, env.errorType(), env.promot())
				continue
			}
			fmt.Printf("unregistered event=%d\n", ev)
		}
	}

	fmt.Printf("total callbacks received: %d\n", atomic.LoadUint64(&received))
	return exitErr
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			fmt.Printf("← %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		}
		next.ServeHTTP(w, r)
	})
}
