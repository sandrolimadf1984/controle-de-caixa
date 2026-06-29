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
	mu        sync.Mutex
	dataPath  string
	active    int       // janelas abertas conectadas
	zeroSince = time.Now()
	lastReq   = time.Now()
	started   = time.Now()
)

func readData() string {
	b, err := os.ReadFile(dataPath)
	if err != nil || len(b) == 0 {
		return "{}"
	}
	return string(b)
}

// Gravação durável: escreve no .tmp, faz fsync (garante que foi pro disco) e troca atomicamente.
func writeData(s string) {
	mu.Lock()
	defer mu.Unlock()
	tmp := dataPath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		_, werr := f.Write([]byte(s))
		f.Sync()
		f.Close()
		if werr == nil {
			os.Remove(dataPath)
			if rerr := os.Rename(tmp, dataPath); rerr == nil {
				// também faz uma cópia de segurança simples
				os.WriteFile(dataPath+".bak", []byte(s), 0644)
				return
			}
		}
	}
	// fallback direto
	os.WriteFile(dataPath, []byte(s), 0644)
	os.WriteFile(dataPath+".bak", []byte(s), 0644)
}

func touch() { mu.Lock(); lastReq = time.Now(); mu.Unlock() }

func main() {
	exe, err := os.Executable()
	dir := "."
	if err == nil {
		dir = filepath.Dir(exe)
	}
	dataPath = filepath.Join(dir, "ControleDeCaixa-dados.json")

	addr := "127.0.0.1:0"
	if p := os.Getenv("CAIXA_PORT"); p != "" {
		addr = "127.0.0.1:" + p
	}
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
		touch()
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
		touch()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
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
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(200)
		fmt.Fprint(w, "ok")
	})
	// Uma janela conectou.
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		active++
		lastReq = time.Now()
		mu.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(200)
		fmt.Fprint(w, "ok")
	})
	// Uma janela foi fechada (enviado no pagehide).
	mux.HandleFunc("/bye", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if active > 0 {
			active--
		}
		if active <= 0 {
			zeroSince = time.Now()
		}
		lastReq = time.Now()
		mu.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(200)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		touch()
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(200)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		go func() { time.Sleep(200 * time.Millisecond); os.Exit(0) }()
	})

	go http.Serve(ln, mux)
	go openApp(url)

	// Encerramento: só quando a janela é realmente fechada (contagem de conexões),
	// nunca durante o uso (minimizar/trocar de janela/suspender não desligam).
	for {
		time.Sleep(3 * time.Second)
		mu.Lock()
		a := active
		zs := zeroSince
		lr := lastReq
		up := time.Since(started)
		mu.Unlock()
		if up < 15*time.Second {
			continue // tempo de carregar a primeira janela
		}
		// Janela fechada (sem conexões ativas) por mais de 6s -> encerra.
		if a <= 0 && time.Since(zs) > 6*time.Second {
			os.Exit(0)
		}
		// Rede de segurança: processo abandonado (sem nenhuma requisição) por 6h.
		if time.Since(lr) > 6*time.Hour {
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
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
