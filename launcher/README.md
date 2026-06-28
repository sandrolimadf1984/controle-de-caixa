# Launcher (Go) — gera o executável de Windows

Este launcher embute o app (`index.html`) e o transforma em um único `.exe`.
Ao rodar, ele sobe um servidor local em `127.0.0.1`, abre o app numa janela do
Edge/Chrome em modo aplicativo e grava os lançamentos em
`ControleDeCaixa-dados.json`, na mesma pasta do executável.

## Compilar

```bash
# a partir desta pasta (launcher/), com o Go instalado:
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui -s -w" -o ControleDeCaixa.exe .
```

O ícone do `.exe` é embutido via `versioninfo.json` (ferramenta `goversioninfo`,
opcional). Sem ela, o programa compila normalmente, só com o ícone padrão.
