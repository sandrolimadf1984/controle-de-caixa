# Controle de Caixa

Sistema de controle de caixa que nasceu para substituir uma planilha de fechamento
diário usada numa rede de laboratórios. A ideia era simples: tirar o time do Excel e
dar uma ferramenta rápida de lançar, conferir e acompanhar os números do dia — mas que
fosse fácil de distribuir entre várias unidades, sem instalar nada e funcionando até de
um pen drive.

Acabou virando um app só, em um arquivo, que roda no navegador **e** como programa de
Windows (`.exe`), com os dados salvos num arquivo ao lado do executável.

> Os dados que vêm no projeto são **fictícios**, só para demonstração.

## Funcionalidades

- **Lançamento por dia e por atendente**, separando Pix, Dinheiro, Cartão e Coleta Externa.
  Cada venda entra como um item separado (não só a soma), o que facilita conferência.
- **Cadastro de atendentes** com atalho/autocompletar no lançamento (e dá pra digitar um
  nome avulso quando alguém cobre o caixa, sem precisar cadastrar).
- **Fontes pagadoras particulares (tabelas/convênios)**: cadastro com busca, vínculo da
  tabela em cada venda, desconto opcional e **pagamento combinado** (quando o cliente paga
  parte no cartão e parte no dinheiro, por exemplo — soma como uma venda só e conta a
  tabela uma única vez).
- **Relatórios e rankings** por dia, mês, trimestre, semestre e ano: quem mais vendeu,
  tabelas mais usadas, participação nas vendas, maior venda do dia, descontos concedidos,
  total de cadastros etc.
- **Gráficos interativos** (barras, pizza, ranking) com tooltip ao passar o mouse.
- **Observações por dia**.
- **Backup e restauração** dos dados em um clique.

## Tecnologias

- **HTML, CSS e JavaScript** (sem framework, sem dependências externas — é um arquivo só).
- **Go** para o launcher que transforma o app em um executável de Windows (servidor local
  em `localhost`, dados gravados em arquivo JSON ao lado do `.exe`).
- **SVG** desenhado na mão para os gráficos.
- `localStorage` + arquivo JSON para persistência (dependendo do modo de uso).

## Como rodar

### Versão web (mais rápido de testar)
É só abrir o arquivo `controle-caixa.html` no navegador (Chrome ou Edge). Funciona offline.

### Versão programa (.exe) para Windows
O executável é gerado a partir da pasta `launcher/` (código em Go). Para compilar:

```bash
cd launcher
# precisa do Go instalado (https://go.dev)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui -s -w" -o ControleDeCaixa.exe .
```

Ao abrir o `.exe`, ele sobe um servidor local, abre o app numa janela e salva tudo num
arquivo `ControleDeCaixa-dados.json` na mesma pasta — por isso é portátil em pen drive.

## Estrutura do projeto

```
controle-caixa.html      -> o app completo (interface + lógica), versão web/demo
launcher/
  main.go                -> launcher em Go: serve o app e grava os dados em arquivo
  index.html             -> o app embutido no executável (sem dados, modo programa)
  versioninfo.json       -> metadados/ícone do .exe
docs/                    -> imagens / capturas de tela
```

## Algumas decisões técnicas (e o porquê)

- **Um arquivo só, sem framework.** Como o app precisava rodar em máquinas variadas e às
  vezes de um pen drive, evitei build pesado e dependências. Tudo num HTML carrega rápido
  e nunca "quebra por falta de pacote".
- **Virar `.exe` com um launcher em Go.** Em vez de empacotar um navegador inteiro
  (pesado), o launcher só sobe um servidor local e abre o app na janela do Edge/Chrome em
  modo aplicativo. O `.exe` final fica pequeno (~5 MB) e os dados ficam num arquivo ao lado
  dele — o que resolveu a portabilidade no pen drive.
- **Cada venda como item individual.** Guardar cada valor separadamente (em vez de só a
  soma) permitiu os rankings, o vínculo de tabela por venda e os pagamentos combinados sem
  retrabalho depois.
- **Salvamento "à prova de susto".** Salva sozinho durante o uso, tem botão de salvar com
  confirmação na tela e ainda grava de novo ao fechar a janela, pra não perder lançamento.

## Autor

**Sandro de Lima Pereira**

## Licença

MIT — veja o arquivo [LICENSE](LICENSE).
