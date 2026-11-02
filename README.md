# Monitor NGINX

Projeto para monitorar requisições HTTP em tempo real a partir do arquivo `access.log` do NGINX, com uma interface web em Go + Server-Sent Events (SSE).

## Visão geral

Este projeto simula três serviços em um único NGINX:

- Porta 80: site principal
- Porta 8081: API / JSON
- Porta 8082: área administrativa

Além disso, possui um painel em Go que lê o log do NGINX e exibe cada requisição em tempo real em uma página web.
Futuramente incluirei um dashboard completo com filtros de logs e informações importantes sobre a leitura do log em tempo real.

## Estrutura do projeto

```text
.
├── docker-compose.yaml
├── logs/
│   └── access.log
├── monitor/
│   ├── main.go
│   └── public/
│       └── index.html
├── nginx/
│   └── nginx.conf
├── scripts/
│   └── simula_trafego.sh
├── sites/
│   ├── admin/
│   ├── api/
│   └── frontend/
└── README.md
```

## Funcionalidades

- Proxy / páginas estáticas em NGINX
- Geração de tráfego por script
- Leitura contínua do `access.log`
- Exibição em tempo real via SSE no navegador
- Visualização de IP, rota, método, status HTTP, tamanho da resposta e User-Agent

## Pré-requisitos

- Docker
- Docker Compose
- Go
- curl

## Subir o ambiente

Na raiz do projeto:

```bash
docker compose up -d
```

Isso sobe o container NGINX com os três sites e monta a pasta `logs/` no host.

## Gerar tráfego

O projeto inclui um script para simular requisições em vários endpoints:

```bash
chmod +x scripts/simula_trafego.sh
./scripts/simula_trafego.sh
```

Ele fica fazendo requisições aleatórias para:

- `http://localhost`
- `http://localhost/contato.html`
- `http://localhost:8081/dados.json`
- `http://localhost:8081/rota-falsa`
- `http://localhost:8082`
- `http://localhost:8082/config`

## Iniciar o monitor

No diretório `monitor`:

```bash
cd monitor
go run main.go
```

A aplicação ficará disponível em:

```text
http://localhost:9090
```

## Como o monitor funciona

O servidor Go expõe o endpoint:

```text
/stream
```

Ele lê o arquivo `logs/access.log` em tempo real e envia cada linha já parseada para o navegador usando o protocolo SSE (`text/event-stream`).

O frontend em `monitor/public/index.html` recebe esses eventos e atualiza a tela dinamicamente sem recarregar a página.

## Configuração do NGINX

O arquivo `nginx/nginx.conf` define três `server blocks` com as portas:

```nginx
server {
    listen 80;
    root /usr/share/nginx/html/frontend;
    index index.html;
}

server {
    listen 8081;
    default_type application/json;
    root /usr/share/nginx/html/api;
}

server {
    listen 8082;
    root /usr/share/nginx/html/admin;
    index index.html;
}
```

## Arquivos importantes

- `docker-compose.yaml`: define o container NGINX
- `nginx/nginx.conf`: configuração dos sites
- `logs/access.log`: log de acesso gerado pelo NGINX
- `monitor/main.go`: servidor Go do monitor
- `monitor/public/index.html`: interface do painel
- `scripts/simula_trafego.sh`: gerador de tráfego

## Troubleshooting

### O monitor não mostra logs

Verifique:

1. Se o NGINX está rodando:
   ```bash
   docker compose ps
   ```

2. Se o arquivo de log existe:
   ```bash
   ls -l logs
   ```

3. Se o servidor do monitor está rodando:
   ```bash
   cd monitor && go run main.go
   ```

4. Se a URL do stream responde:
   ```bash
   curl -N http://localhost:9090/stream
   ```

### O container NGINX não responde

Tente reiniciar o ambiente:

```bash
docker compose down
docker compose up -d
```

## Observações

- O projeto foi pensado para ambiente local de desenvolvimento e monitoramento.
- O monitor lê o `access.log` diretamente do host, então as permissões e o caminho do arquivo devem estar corretos.
- O painel é útil para visualizar tráfego real de aplicações HTTP em tempo real.

## Licença

Este projeto é um protótipo para um projeto da faculdade e pode ser adaptado livremente.
