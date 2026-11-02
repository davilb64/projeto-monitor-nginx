package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ncruces/zenity"
)

type LogEntry struct {
	IP           string
	DataHora     string
	Metodo       string
	Rota         string
	Protocolo    string
	Status       string
	TamanhoBytes string
	Referer      string
	UserAgent    string
}

var (
	padraoNGINXCombined = regexp.MustCompile(`^(\S+) \S+ \S+ \[(.*?)\] "(.*?)" (\d{3}) (\S+) "(.*?)" "(.*?)"$`)
	padraoNGINXCommon   = regexp.MustCompile(`^(\S+) \S+ \S+ \[(.*?)\] "(.*?)" (\d{3}) (\S+)$`)
)

func parseRequestLine(requestLine string) (metodo string, rota string, protocolo string, err error) {
	partes := strings.SplitN(requestLine, " ", 3)
	if len(partes) != 3 {
		return "", "", "", errors.New("request line fora do formato esperado")
	}
	return partes[0], partes[1], partes[2], nil
}

func parseLinhaLog(linha string) (LogEntry, bool) {
	linha = strings.TrimSpace(linha)
	if linha == "" {
		return LogEntry{}, false
	}

	if matches := padraoNGINXCombined.FindStringSubmatch(linha); len(matches) == 8 {
		metodo, rota, protocolo, err := parseRequestLine(matches[3])
		if err != nil {
			return LogEntry{}, false
		}
		return LogEntry{
			IP:           matches[1],
			DataHora:     matches[2],
			Metodo:       metodo,
			Rota:         rota,
			Protocolo:    protocolo,
			Status:       matches[4],
			TamanhoBytes: matches[5],
			Referer:      matches[6],
			UserAgent:    matches[7],
		}, true
	}

	if matches := padraoNGINXCommon.FindStringSubmatch(linha); len(matches) == 6 {
		metodo, rota, protocolo, err := parseRequestLine(matches[3])
		if err != nil {
			return LogEntry{}, false
		}
		return LogEntry{
			IP:           matches[1],
			DataHora:     matches[2],
			Metodo:       metodo,
			Rota:         rota,
			Protocolo:    protocolo,
			Status:       matches[4],
			TamanhoBytes: matches[5],
			Referer:      "-",
			UserAgent:    "-",
		}, true
	}

	return LogEntry{}, false
}

func obterCaminhoLog() string {
	caminhosComuns := []string{
		"/home/davi/Área de trabalho/projeto-monitor-nginx/logs/access.log",
		"../logs/access.log",
		"./logs/access.log",
		"/var/log/nginx/access.log",
		"C:\\nginx\\logs\\access.log",
	}

	for _, caminho := range caminhosComuns {
		if _, err := os.Stat(caminho); err == nil {
			fmt.Printf("[\u2713] NGINX detectado! Lendo log de: %s\n", caminho)
			return caminho
		}
	}

	fmt.Println("[!] Log não encontrado automaticamente. Abrindo seletor...")
	caminhoEscolhido, err := zenity.SelectFile(zenity.Title("Selecione o arquivo access.log"))
	if err != nil {
		fmt.Println("Operação cancelada.")
		os.Exit(1)
	}

	return caminhoEscolhido
}

// A função agora recebe o caminho do arquivo e roda a cada nova aba do navegador aberta
func streamDeDados(w http.ResponseWriter, r *http.Request, caminhoArquivo string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming não suportado", http.StatusInternalServerError)
		return
	}

	// 1. Abre o arquivo toda vez que o navegador conecta
	file, err := os.Open(caminhoArquivo)
	if err != nil {
		fmt.Println("Erro ao ler o log:", err)
		return
	}
	defer file.Close()

	// 2. Lê desde a primeira linha (Removemos o file.Seek que pulava pro final)
	reader := bufio.NewReader(file)
	for {
		linha, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				// Chegou no fim do arquivo. Espera meio segundo e tenta ler novas linhas.
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break // Se for outro erro grave, encerra a conexão
		}

		entrada, ok := parseLinhaLog(linha)
		if ok {
			jsonBytes, err := json.Marshal(entrada)
			if err != nil {
				fmt.Println("Erro ao serializar log:", err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", string(jsonBytes))
			flusher.Flush()
		} else {
			// Printa no terminal caso o NGINX tenha gerado uma linha diferente do esperado
			fmt.Printf("Aviso: Linha fora do padrão ignorada: %s", linha)
		}
	}
}

func main() {
	caminhoDoLog := obterCaminhoLog()

	// Uso de Closure (Função Anônima) para passar o caminho do log para a rota
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		streamDeDados(w, r, caminhoDoLog)
	})

	http.Handle("/", http.FileServer(http.Dir("./public")))

	fmt.Println("Monitor rodando perfeitamente!")
	fmt.Println("Acesse o painel em: http://localhost:9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}
}
