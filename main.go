package main

import (
	"flag"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/lpernett/godotenv"

	"github.com/eugene-ivanov-hash/mitm-proxy/proxy"
	"github.com/eugene-ivanov-hash/mitm-proxy/rule"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9999", "proxy address")
	caCertFile := flag.String("cacertfile", "", "certificate .pem file for trusted CA")
	caKeyFile := flag.String("cakeyfile", "", "key .pem file for trusted CA")
	debug := flag.Bool("debug", false, "enable debug logging")
	rulesDir := flag.String("rulesdir", "proxy_rules", "directory for rules")
	envFile := flag.String("env", "", "environment file")
	testRules := flag.Bool("test", false, "test rules")
	flag.Parse()

	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var err error
	if *envFile != "" {
		err = godotenv.Load(*envFile)
	} else {
		err = godotenv.Load()
	}
	if err != nil {
		slog.Debug("Error loading environment file", slog.String("err", err.Error()))
	}

	envs, err := godotenv.Parse(strings.NewReader(strings.Join(os.Environ(), "\n")))
	if err != nil {
		slog.Debug("Error parsing environment variables", slog.String("err", err.Error()))
	}

	_ = envs

	requestRules, responseRules, err := rule.CompileRules(*rulesDir, envs)
	if err != nil {
		slog.Error("Error compiling rules", slog.String("err", err.Error()))
		return
	}

	if *testRules {
		request := &http.Request{
			Method: "POST",
			URL:    &url.URL{Scheme: "https", Host: "example.com"},
			Host:   "example.com",
			Header: http.Header{
				"User-Agent": []string{"curl/7.64.1"},
			},
			Body: io.NopCloser(strings.NewReader("test body")),
		}

		response := &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/html; charset=UTF-8"},
			},
			Body: io.NopCloser(strings.NewReader("test body")),
		}

		for _, r := range requestRules {
			_, err := r.Check(request, nil)
			if err != nil {
				slog.Error("Error checking request rule", slog.String("rule", r.Name), slog.String("err", err.Error()))
				continue
			}

			err = r.Apply(request, nil)
			if err != nil {
				slog.Error("Error applying request rule", slog.String("rule", r.Name), slog.String("err", err.Error()))
				continue
			}
		}

		for _, r := range responseRules {
			_, err := r.Check(request, response)
			if err != nil {
				slog.Error("Error checking request rule", slog.String("rule", r.Name), slog.String("err", err.Error()))
				continue
			}

			err = r.Apply(request, response)
			if err != nil {
				slog.Error("Error applying response rule", slog.String("rule", r.Name), slog.String("err", err.Error()))
				continue
			}
		}

		return
	}

	proxySSl := proxy.NewProxySslServer(*caCertFile, *caKeyFile, requestRules, responseRules)

	slog.Info("Starting proxy server on", slog.String("addr", *addr))

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		slog.Error("Error listening on", slog.String("addr", *addr), slog.String("err", err.Error()))
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Error accepting connection", slog.String("err", err.Error()))
			continue
		}

		go proxySSl.HandleTLS(conn)
	}
}
