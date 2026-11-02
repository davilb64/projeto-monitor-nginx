#!/bin/bash
URLS=(
    "http://localhost"
    "http://localhost/contato.html"
    "http://localhost:8081/dados.json"
    "http://localhost:8081/rota-falsa"
    "http://localhost:8082"
    "http://localhost:8082/config"
)

echo "Gerando trafego. ctrl c para parar"

while true; do
    URL=${URLS[$RANDOM % ${#URLS[@]}]}
    curl -s -o /dev/null $URL
    sleep 0.$((RANDOM % 5 + 1))
done