# auth-service

Um serviço de autenticação, pronto para ser embutido em uma aplicação web. Ele oferece cadastro com confirmação por email, login, sessão por cookie seguro, JWT, rate limiting e uma interface web em SolidJS.

> **Nota importante:** usuários são persistidos no banco de dados. Cadastros ainda não confirmados e limitadores de requisições ficam em memória; por isso, um reinício do processo apaga confirmações pendentes e os contadores de rate limit.

## Fluxo de autenticação

~~~text
Cadastro ──► código por email ──► verificação ──► usuário criado
                                                    │
                                                    ▼
                                             login + cookie JWT
                                                    │
                                                    ▼
                                              endpoints protegidos
~~~

## Requisitos

- Go **1.26.4** ou compatível com a versão declarada em <code>backend/go.mod</code>;
- Node.js 24 e Yarn para compilar o frontend;
- uma conta do [Resend](https://resend.com/) com domínio remetente configurado;
- Docker e Docker Compose, caso opte pela execução conteinerizada.

## Execução local

### 1. Configure o backend

Na raiz do projeto:

~~~bash
cp backend/.env.example backend/.env
./scripts/generate-jwt-secret.sh
~~~

Copie o valor exibido pelo script para <code>JWT_SECRET</code> em <code>backend/.env</code>. O segredo precisa ter pelo menos 32 bytes e não pode ser um valor de exemplo.

Depois, informe suas credenciais do Resend:

~~~dotenv
RESEND_API_KEY=re_...
RESEND_FROM=noreply@seu-dominio.com
~~~

O endereço usado em <code>RESEND_FROM</code> deve ser aceito pelo domínio verificado no Resend.

### 2. Instale as dependências e compile

~~~bash
cd frontend
yarn install
cd ..
make build
~~~

O build gera o frontend e o copia para <code>backend/cmd/dist</code>, onde ele será incorporado ao binário Go.

### 3. Inicie o serviço

~~~bash
make start
~~~

Abra <http://localhost:8888>. A API estará disponível em <code>http://localhost:8888/api</code>.

Para desenvolvimento visual do frontend, mantenha o backend rodando e use outro terminal:

~~~bash
cd frontend
yarn dev
~~~

O Vite disponibiliza a aplicação normalmente em <http://localhost:5173> e encaminha <code>/api</code> para o backend em <code>localhost:8888</code>.

## Execução com Docker

### 1. Crie o arquivo de ambiente

~~~bash
cp .env.example .env
./scripts/generate-jwt-secret.sh
~~~

Cole o segredo gerado em <code>JWT_SECRET</code> e preencha <code>RESEND_API_KEY</code> e <code>RESEND_FROM</code> no <code>.env</code>.

### 2. Suba o serviço

~~~bash
docker compose up --build -d
docker compose logs -f auth
~~~

A aplicação ficará em <http://localhost:8123>. O banco SQLite é mantido no volume Docker <code>auth_data</code>.

Para parar o serviço:

~~~bash
docker compose down
~~~

O volume não é removido pelo comando acima. Para removê-lo deliberadamente, use <code>docker compose down -v</code> — isso apaga os usuários armazenados no volume.

## Configuração

Para execução local, use <code>backend/.env</code>. No Docker, use o <code>.env</code> da raiz.

| Variável | Obrigatória | Padrão | Uso |
| --- | ---: | --- | --- |
| <code>DB_URL</code> | Sim | — | Caminho do SQLite. Localmente, <code>file:auth.sqlite</code>. |
| <code>JWT_SECRET</code> | Sim | — | Segredo HS256 com pelo menos 32 bytes. |
| <code>JWT_SECRET_PREVIOUS</code> | Não | vazio | Segredo anterior durante uma rotação de chaves. |
| <code>RESEND_API_KEY</code> | Sim | — | Chave da API de envio de email. |
| <code>RESEND_FROM</code> | Sim | — | Remetente válido e verificado. |
| <code>RESEND_API_URL</code> | Não | API do Resend | Endpoint alternativo, útil em testes ou proxies. |
| <code>SERVER_PORT</code> | Não | <code>8888</code> | Porta HTTP do serviço. |
| <code>FRAME_ANCESTORS</code> | Não | <code>'self'</code> | Origens permitidas pelo header CSP <code>frame-ancestors</code>. |
| <code>COOKIE_SECURE</code> | Não | automático | Use <code>true</code> em HTTPS; em desenvolvimento, <code>false</code>. |
| <code>COOKIE_SAMESITE</code> | Não | <code>lax</code> | Política SameSite do cookie: <code>lax</code>, <code>strict</code> ou <code>none</code>. |
| <code>VITE_AUTH_BRIDGE_ORIGINS</code> | Não | vazio | Lista, separada por vírgulas, de origens autorizadas para o bridge. É aplicada no build do frontend. |

## API

Todos os endpoints JSON usam <code>Content-Type: application/json</code>. Erros são retornados no formato:

~~~json
{
  "error": "mensagem do erro"
}
~~~

| Método | Endpoint | Autenticação | Descrição |
| --- | --- | --- | --- |
| <code>GET</code> | <code>/healthz</code> | Pública | Verifica se o processo está ativo. |
| <code>GET</code> | <code>/readyz</code> | Pública | Verifica se o SQLite está disponível. |
| <code>GET</code> | <code>/metrics</code> | Pública | Expõe métricas no formato Prometheus. |
| <code>POST</code> | <code>/api/register</code> | Pública | Cria um cadastro pendente e envia o código. |
| <code>POST</code> | <code>/api/verify</code> | Pública | Confirma o código e cria o usuário. |
| <code>POST</code> | <code>/api/resend</code> | Pública | Reenvia o código de um cadastro pendente. |
| <code>POST</code> | <code>/api/login</code> | Pública | Valida credenciais e cria a sessão. |
| <code>POST</code> | <code>/api/logout</code> | Cookie/origem | Encerra a sessão no navegador. |
| <code>GET</code> | <code>/api/me</code> | Cookie ou Bearer | Retorna o usuário autenticado. |
| <code>POST</code> | <code>/api/bridge/token</code> | Cookie ou Bearer | Gera um token de integração válido por 5 minutos. |

### Cadastro e confirmação

~~~bash
export AUTH_URL=http://localhost:8888
curl -i -X POST "$AUTH_URL/api/register" -H 'Content-Type: application/json' --data '{"email":"ana@example.com","password":"uma-senha-segura"}'
~~~

O código chega no email informado. Confirme a conta com:

~~~bash
curl -i -X POST "$AUTH_URL/api/verify" -H 'Content-Type: application/json' --data '{"email":"ana@example.com","code":"123456"}'
~~~

Reenvio de código:

~~~bash
curl -i -X POST "$AUTH_URL/api/resend" -H 'Content-Type: application/json' --data '{"email":"ana@example.com"}'
~~~

O código expira em 15 minutos e a conta é removida após 5 tentativas inválidas.

### Login, sessão e usuário atual

Guarde o cookie retornado pelo login:

~~~bash
curl -i -c cookies.txt -X POST "$AUTH_URL/api/login" -H 'Content-Type: application/json' --data '{"email":"ana@example.com","password":"uma-senha-segura"}'
~~~

Consulte o usuário autenticado:

~~~bash
curl -i -b cookies.txt "$AUTH_URL/api/me"
~~~

Também é possível usar o JWT explicitamente:

~~~bash
curl -i "$AUTH_URL/api/me" -H "Authorization: Bearer SEU_TOKEN"
~~~

Encerre a sessão:

~~~bash
curl -i -b cookies.txt -X POST "$AUTH_URL/api/logout"
~~~

## Regras e proteções

- emails são normalizados para letras minúsculas e limitados a 254 caracteres;
- senhas têm entre 8 e 72 bytes e nunca são armazenadas em texto puro;
- códigos de confirmação têm exatamente 6 dígitos;
- tokens de sessão duram 24 horas;
- tokens de bridge duram 5 minutos e não renovam a sessão principal;
- o corpo de cada requisição é limitado a 1 MiB;
- JSON com campos desconhecidos ou conteúdo extra é rejeitado;
- endpoints sensíveis têm limites por IP, email e combinação IP/email;
- cookies usam <code>HttpOnly</code> e <code>SameSite</code>, com <code>Secure</code> configurável;
- respostas de login e endpoints de sessão usam <code>Cache-Control: no-store</code>.

## Bridge entre aplicações

O frontend pode entregar um token curto a uma aplicação pai dentro de um <code>iframe</code>. Configure as origens exatas antes do build:

~~~dotenv
VITE_AUTH_BRIDGE_ORIGINS=https://app.example.com,https://admin.example.com
~~~

A aplicação pai envia uma mensagem para o iframe:

~~~js
iframe.contentWindow.postMessage(
  {
    type: 'AUTH_SERVICE/TOKEN_REQUEST',
    requestId: 'dashboard-001',
  },
  'https://auth.example.com',
)
~~~

A resposta contém <code>type: 'AUTH_SERVICE/TOKEN_RESPONSE'</code>, o mesmo <code>requestId</code> e um <code>token</code> JWT válido por 5 minutos. Valide sempre <code>event.origin</code> no lado da aplicação consumidora.

## Desenvolvimento e testes

Backend:

~~~bash
cd backend
go test ./...
go vet ./...
~~~

Frontend:

~~~bash
cd frontend
yarn lint
yarn test:frontend
yarn test:components
~~~

Comandos úteis na raiz:

~~~bash
make build   # compila frontend e backend
make start   # compila e inicia o serviço
make clean   # remove artefatos de build
~~~