# Finance Tracker

Aplicação pessoal de controlo de despesas e orçamentos. Web app instalável no telemóvel (PWA), com interface em português.

## Stack

- **Frontend** — Next.js (App Router, TypeScript), em `web/`
- **Backend** — Go, API REST em `server/`
- **Base de dados** — PostgreSQL
- **Deploy** — Docker / Dokploy

## Funcionalidades

- Orçamentos mensais e anuais por categoria (água, eletricidade, gás, gasolina, comida dos animais, supermercado, …); os anuais aparecem diluídos a 1/12 no mês
- Registo rápido de despesas com as mesmas categorias
- Despesas recorrentes: a app gera a despesa prevista todos os meses e só confirmas o valor real
- Projeção de gasto até ao fim do mês e rollover opcional do saldo da categoria
- Relatórios: evolução mensal por categoria e comparação com meses anteriores
- Utilizador único com login (sessões em cookie, password com Argon2id)

## Desenvolvimento

Requisitos: Go 1.23+, Node 20+, e um PostgreSQL acessível (por exemplo via `docker compose up db`).

```sh
make dev-server   # API Go em :8080
make dev-web      # Next.js em :3000 (proxy de /api para :8080)
make test         # testes do backend
```

Variáveis de ambiente do servidor (ver `.env.example`):

- `DATABASE_URL` — ligação Postgres
- `PORT` — porta da API (por omissão 8080)
- `COOKIE_SECURE=true` — em produção, atrás de HTTPS

## Estrutura

```
server/   API Go (auth, orçamentos, despesas, recorrentes, relatórios)
web/      Next.js (UI «Bolso», PWA)
```
