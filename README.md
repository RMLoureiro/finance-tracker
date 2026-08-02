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

Requisitos: Go 1.23+ e Node 20+. Para a base de dados usa uma das opções:

- `make dev-db` — Postgres embebido em `localhost:5433`, sem Docker nem instalação;
- ou `docker compose up -d db` — Postgres em `localhost:5432` (ajusta `DATABASE_URL`).

```sh
make dev-db       # terminal 1: Postgres de desenvolvimento (localhost:5433)
make dev-server   # terminal 2: API Go em :8080 (migra e faz seed ao arrancar)
make dev-web      # terminal 3: Next.js em :3000 (proxy de /api para :8080)
make test         # testes do backend (usam Postgres embebido)
```

Abre http://localhost:3000 — no primeiro arranque a app pede-te para criar o acesso.

## Deploy (Dokploy)

O `Dockerfile` produz uma imagem única: o binário Go serve a API e o export
estático do Next. O `docker-compose.prod.yml` junta a app e o Postgres:

1. No Dokploy cria um serviço **Compose** apontado a este repositório com o
   ficheiro `docker-compose.prod.yml`;
2. Define a variável `POSTGRES_PASSWORD` no ambiente do projeto;
3. Associa o domínio ao serviço `app`, porta `8080` — o Traefik do Dokploy
   trata do HTTPS (os cookies já vão com `Secure`).

Os dados ficam no volume `pgdata`; faz backup com `pg_dump` regularmente.

Variáveis de ambiente do servidor (ver `.env.example`):

- `DATABASE_URL` — ligação Postgres
- `PORT` — porta da API (por omissão 8080)
- `COOKIE_SECURE=true` — em produção, atrás de HTTPS

## Estrutura

```
server/   API Go (auth, orçamentos, despesas, recorrentes, relatórios)
web/      Next.js (UI «Bolso», PWA)
```
