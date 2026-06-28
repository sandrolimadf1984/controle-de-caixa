package main

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed index.html
var indexHTML string

//go:embed icon.png
var iconPNG []byte

var (
	mu       sync.Mutex
	dataPath string
	lastPing = time.Now()
	started  = time.Now()
)

func readData() string {
	b, err := os.ReadFile(dataPath)
	if err != nil || len(b) == 0 {
		return "{}"
	}
	return string(b)
}

func writeData(s string) {
	mu.Lock()
	defer mu.Unlock()
	tmp := dataPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(s), 0644); err == nil {
		os.Remove(dataPath)
		os.Rename(tmp, dataPath)
	} else {
		os.WriteFile(dataPath, []byte(s), 0644)
	}
}

func touch() { mu.Lock(); lastPing = time.Now(); mu.Unlock() }

func main() {
	exe, err := os.Executable()
	dir := "."
	if err == nil {
		dir = filepath.Dir(exe)
	}
	dataPath = filepath.Join(dir, "ControleDeCaixa-dados.json")

	addr := "127.0.0.1:0"
	if p := os.Getenv("CAIXA_PORT"); p != "" { addr = "127.0.0.1:" + p }
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		os.WriteFile(filepath.Join(dir, "erro-controle-caixa.txt"),
			[]byte("Nao foi possivel iniciar o programa: "+err.Error()), 0644)
		return
	}
	port := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d/", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data := readData()
		inject := "<script>window.__DADOS__=" + data +
			";window.__DADOS_AT__=" + fmt.Sprintf("%d", time.Now().UnixMilli()) +
			";window.__APPMODE__=true;</script>\n<link rel=\"icon\" href=\"/icon.png\">"
		page := strings.Replace(indexHTML, "<!--INJECT-->", inject, 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprint(w, page)
	})
	mux.HandleFunc("/icon.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(iconPNG)
	})
	mux.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, readData())
	})
	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		touch()
		b := make([]byte, 0, 1<<20)
		buf := make([]byte, 65536)
		for {
			n, e := r.Body.Read(buf)
			if n > 0 {
				b = append(b, buf[:n]...)
			}
			if e != nil {
				break
			}
		}
		if len(b) > 0 {
			writeData(string(b))
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		touch()
		w.WriteHeader(200)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		go func() { time.Sleep(200 * time.Millisecond); os.Exit(0) }()
	})

	go http.Serve(ln, mux)

	go openApp(url)

	// Auto-quit shortly after the window is closed (no pings).
	for {
		time.Sleep(4 * time.Second)
		mu.Lock()
		idle := time.Since(lastPing)
		up := time.Since(started)
		mu.Unlock()
		if up > 20*time.Second && idle > 12*time.Second {
			os.Exit(0)
		}
	}
}

func openApp(url string) {
	pf := os.Getenv("ProgramFiles")
	pfx := os.Getenv("ProgramFiles(x86)")
	lad := os.Getenv("LOCALAPPDATA")
	candidates := []string{
		filepath.Join(pf, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pfx, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(pfx, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(lad, `Google\Chrome\Application\chrome.exe`),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			cmd := exec.Command(c, "--app="+url, "--window-size=1200,860")
			if cmd.Start() == nil {
				return
			}
		}
	}
	// Fallback: default browser (normal tab).
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
