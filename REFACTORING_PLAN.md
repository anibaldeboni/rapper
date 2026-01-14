# Plano de Refatoração - Rapper TUI

## 📋 Análise da Estrutura Atual

### Pontos Fortes
- ✅ Boa separação de responsabilidades (config, processor, web, ui)
- ✅ Uso de interfaces para mockability
- ✅ Padrões modernos (generics, atomic operations, context)
- ✅ Testes com mocks e cobertura adequada
- ✅ Bom uso de Bubbletea components (list, viewport, spinner)

### Pontos de Melhoria Identificados

#### 1. **Configuração Estática e Rígida**
**Problema Atual:**
```yaml
# config.yml - Lido uma única vez na inicialização
token: "JWT_TOKEN"              # Apenas 1 token fixo
path:
  method: PUT                    # Método fixo
  template: "https://..."        # URL fixa
payload:
  template: |                    # Payload fixo
    {"field": "{{.fieldname}}"}
```

**Limitações:**
- Token único no header `Authorization: Bearer <token>`
- Impossível trocar configuração sem reiniciar
- Sem suporte para múltiplos headers (Cookie, X-API-Key, etc)
- Workers definidos por flag CLI

---

#### 2. **Estrutura TUI Simples Demais**
**Problema Atual:**
```go
type Model struct {
    viewport  viewport.Model   // Apenas logs
    filesList list.Model       // Apenas seleção CSV
    help      help.Model       // Ajuda
    spinner   spinner.Model    // Loading
}
```

**Limitações:**
- Apenas 3 estados: `SelectFile`, `Running`, `Stale`
- Nenhuma view de configuração
- Sem formulários para editar settings
- Modelo monolítico (um único arquivo `ui.go` com 300+ linhas)

---

#### 3. **Headers Hardcoded**
**Problema Atual:**
```go
// web/gateway.go - Exec()
header := map[string]string{
    "Authorization": fmt.Sprintf("Bearer %s", g.token),
}
```

**Limitações:**
- Sempre envia `Authorization: Bearer <token>`
- Sem flexibilidade para Cookie, X-API-Key, Custom-Header, etc
- Não suporta múltiplos headers simultâneos

---

#### 4. **Workers Inflexíveis**
**Problema Atual:**
```go
// main.go
workers := flag.Int("workers", 1, "number of workers")

// Fixado na criação do Processor
csvProcessor := processor.NewProcessor(cfg, gateway, logger, *workers)
```

**Limitações:**
- Workers definidos no startup
- Não pode aumentar/diminuir durante execução
- Sem feedback visual na UI do número atual

---

## 🎯 Objetivos da Refatoração

### 1. Simplificação da Estrutura
- [ ] Consolidar arquivos pequenos relacionados
- [ ] Reduzir duplicação (estilos, mensagens)
- [ ] Melhorar organização do módulo `ui/`

### 2. Melhores Práticas Bubbletea
- [ ] Separar views em componentes reutilizáveis
- [ ] Implementar padrão "multi-view" com navegação
- [ ] Usar tea.Cmd para operações assíncronas
- [ ] Aplicar padrão Elm Architecture corretamente

### 3. Configuração Dinâmica de Requests
- [ ] Adicionar view de "Settings" na TUI
- [ ] Formulário para editar URL template, body template, headers
- [ ] Hot-reload de configuração sem restart
- [ ] Salvar/carregar múltiplos profiles

### 4. Workers Dinâmicos
- [ ] Slider/input para ajustar workers em runtime
- [ ] Feedback visual de workers ativos
- [ ] Métrica de throughput (req/s)

### 5. Headers Flexíveis
- [ ] Substituir `token: string` por `headers: map[string]string`
- [ ] UI para adicionar/remover headers customizados
- [ ] Templates suportam variáveis em headers

---

## 🏗️ Arquitetura Proposta

### Nova Estrutura de Diretórios

```
internal/
├── config/
│   ├── config.go           # Estruturas de configuração
│   ├── loader.go           # Carregamento de YAML
│   └── manager.go          # [NOVO] Gerenciamento em runtime
│
├── ui/
│   ├── app.go              # [REFATORADO] Model principal
│   ├── commands.go         # [NOVO] tea.Cmd factories
│   ├── navigation.go       # [NOVO] Controle de navegação entre views
│   │
│   ├── views/              # [NOVO] Views separadas
│   │   ├── files.go        # View de seleção de arquivos
│   │   ├── logs.go         # View de logs (processamento)
│   │   ├── settings.go     # View de configuração
│   │   └── workers.go      # View de controle de workers
│   │
│   ├── components/         # [NOVO] Componentes reutilizáveis
│   │   ├── header.go       # Header com título e breadcrumb
│   │   ├── form.go         # Formulário genérico (input, textarea)
│   │   └── metrics.go      # Painel de métricas (workers, req/s)
│   │
│   └── styles/             # [MOVIDO de internal/styles]
│       └── styles.go
│
├── processor/
│   ├── processor.go        # [REFATORADO] Interface + impl
│   ├── worker_pool.go      # [NOVO] Pool de workers ajustável
│   └── metrics.go          # [NOVO] Métricas em tempo real
│
└── web/
    ├── gateway.go          # [REFATORADO] Headers flexíveis
    └── client.go
```

---

## 📐 Detalhamento das Mudanças

### 1. Config Manager - Configuração Dinâmica

#### Nova Estrutura de Config
```go
// config/config.go
type Config struct {
    Request  RequestConfig         `yaml:"request"`
    CSV      CSVConfig            `yaml:"csv"`
    Workers  int                  `yaml:"workers"`  // Valor inicial
}

type RequestConfig struct {
    Method      string            `yaml:"method"`
    URLTemplate string            `yaml:"url_template"`
    BodyTemplate string           `yaml:"body_template"`
    Headers     map[string]string `yaml:"headers"`  // ✨ NOVO: flexível
}

type CSVConfig struct {
    Fields    []string `yaml:"fields"`
    Separator string   `yaml:"separator"`
}
```

#### Exemplo de config.yml atualizado
```yaml
request:
  method: POST
  url_template: "https://api.example.com/users/{{.id}}"
  body_template: |
    {
      "name": "{{.name}}",
      "email": "{{.email}}"
    }
  headers:
    Authorization: "Bearer eyJhbGc..."
    X-API-Key: "my-secret-key"
    Cookie: "session_id=abc123"
    Content-Type: "application/json"

csv:
  fields: [id, name, email]
  separator: ","

workers: 4  # Inicial, ajustável na UI
```

#### Manager para Hot-Reload
```go
// config/manager.go
type Manager interface {
    Get() *Config
    Update(cfg *Config) error
    Save() error
    OnChange(callback func(*Config))
}

type managerImpl struct {
    current  *Config
    filePath string
    mu       sync.RWMutex
    listeners []func(*Config)
}

func (m *managerImpl) Update(cfg *Config) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.current = cfg

    // Notifica listeners (processor, gateway)
    for _, listener := range m.listeners {
        listener(cfg)
    }

    return nil
}
```

---

### 2. TUI Multi-View Architecture

#### Views Enum
```go
// ui/navigation.go
type View int

const (
    ViewFiles View = iota    // Seleção de CSV
    ViewLogs                 // Logs de processamento
    ViewSettings             // Configuração de request
    ViewWorkers              // Controle de workers
)

type Navigation struct {
    current View
    history []View
}

func (n *Navigation) Push(v View) { /* ... */ }
func (n *Navigation) Back() View { /* ... */ }
```

#### Model Principal Refatorado
```go
// ui/app.go
type Model struct {
    // Estado
    nav       *Navigation
    state     *State  // Running/Stale/etc

    // Dependências
    configMgr config.Manager
    processor processor.Processor
    logger    logs.Logger

    // Sub-models (um para cada view)
    filesView    *FilesView
    logsView     *LogsView
    settingsView *SettingsView
    workersView  *WorkersView

    // Comuns
    help   help.Model
    width  int
    height int
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Navegação global
        switch msg.String() {
        case "ctrl+s":
            m.nav.Push(ViewSettings)
            return m, nil
        case "ctrl+w":
            m.nav.Push(ViewWorkers)
            return m, nil
        case "esc":
            m.nav.Back()
            return m, nil
        }

        // Delega para view atual
        return m.updateCurrentView(msg)
    }

    return m, nil
}

func (m Model) View() string {
    // Renderiza view atual
    switch m.nav.current {
    case ViewFiles:
        return m.filesView.Render(m.width, m.height)
    case ViewLogs:
        return m.logsView.Render(m.width, m.height)
    case ViewSettings:
        return m.settingsView.Render(m.width, m.height)
    case ViewWorkers:
        return m.workersView.Render(m.width, m.height)
    }

    return ""
}
```

---

### 3. Settings View - Edição de Configuração

```go
// ui/views/settings.go
type SettingsView struct {
    configMgr config.Manager

    // Form inputs (usando bubbles/textinput)
    methodInput      textinput.Model
    urlInput         textinput.Model
    bodyInput        textarea.Model
    headersEditor    *HeadersEditor  // [NOVO]

    focusIndex int
    focused    bool
}

type HeadersEditor struct {
    headers map[string]string
    list    list.Model
    keyInput   textinput.Model
    valueInput textinput.Model
    editing    bool
}

func (s *SettingsView) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            s.focusIndex = (s.focusIndex + 1) % 4
            s.updateFocus()

        case "enter":
            if s.focusIndex == 3 {  // Headers editor
                return s.headersEditor.Toggle()
            }

        case "ctrl+s":
            // Salva configuração
            return s.saveConfig()
        }
    }

    return nil
}

func (s *SettingsView) saveConfig() tea.Cmd {
    return func() tea.Msg {
        cfg := s.configMgr.Get()

        cfg.Request.Method = s.methodInput.Value()
        cfg.Request.URLTemplate = s.urlInput.Value()
        cfg.Request.BodyTemplate = s.bodyInput.Value()
        cfg.Request.Headers = s.headersEditor.headers

        if err := s.configMgr.Update(cfg); err != nil {
            return ConfigErrorMsg{err}
        }

        if err := s.configMgr.Save(); err != nil {
            return ConfigErrorMsg{err}
        }

        return ConfigSavedMsg{}
    }
}

func (s *SettingsView) Render(width, height int) string {
    var b strings.Builder

    b.WriteString(styles.TitleStyle.Render("⚙️  Request Settings"))
    b.WriteString("\n\n")

    b.WriteString("Method:\n")
    b.WriteString(s.methodInput.View())
    b.WriteString("\n\n")

    b.WriteString("URL Template:\n")
    b.WriteString(s.urlInput.View())
    b.WriteString("\n\n")

    b.WriteString("Body Template:\n")
    b.WriteString(s.bodyInput.View())
    b.WriteString("\n\n")

    b.WriteString("Headers (press Enter to edit):\n")
    b.WriteString(s.headersEditor.View())
    b.WriteString("\n\n")

    b.WriteString("Ctrl+S: Save | Esc: Back")

    return styles.AppStyle.Render(b.String())
}
```

**Headers Editor Component:**
```go
// ui/components/headers_editor.go
type HeadersEditor struct {
    headers map[string]string
    items   []list.Item
    list    list.Model

    // Modal para adicionar/editar
    editing    bool
    keyInput   textinput.Model
    valueInput textinput.Model
}

func (h *HeadersEditor) View() string {
    if h.editing {
        return h.renderModal()
    }

    return h.list.View()
}

func (h *HeadersEditor) renderModal() string {
    var b strings.Builder

    b.WriteString("╔══ Add Header ══╗\n")
    b.WriteString("║ Key:   " + h.keyInput.View() + " ║\n")
    b.WriteString("║ Value: " + h.valueInput.View() + " ║\n")
    b.WriteString("╚════════════════╝\n")
    b.WriteString("Enter: Save | Esc: Cancel")

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(1).
        Render(b.String())
}

// Lista de headers como items
type headerItem struct {
    key   string
    value string
}

func (i headerItem) FilterValue() string { return i.key }
func (i headerItem) Title() string       { return i.key }
func (i headerItem) Description() string { return i.value }
```

---

### 4. Workers View - Controle Dinâmico

```go
// ui/views/workers.go
type WorkersView struct {
    processor processor.Processor

    slider    int  // Valor atual do slider (1-16)
    maxWorkers int  // runtime.NumCPU() * 2

    // Métricas
    activeWorkers int
    requestsPerSec float64
    totalRequests  uint64
}

func (w *WorkersView) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "left", "h":
            if w.slider > 1 {
                w.slider--
                return w.applyWorkers()
            }

        case "right", "l":
            if w.slider < w.maxWorkers {
                w.slider++
                return w.applyWorkers()
            }

        case "enter":
            return w.applyWorkers()
        }

    case MetricsMsg:
        w.requestsPerSec = msg.ReqPerSec
        w.totalRequests = msg.TotalReq
        w.activeWorkers = msg.ActiveWorkers
        return w.tickMetrics()
    }

    return nil
}

func (w *WorkersView) applyWorkers() tea.Cmd {
    return func() tea.Msg {
        if err := w.processor.SetWorkers(w.slider); err != nil {
            return WorkersErrorMsg{err}
        }

        return WorkersUpdatedMsg{workers: w.slider}
    }
}

func (w *WorkersView) tickMetrics() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        metrics := w.processor.GetMetrics()
        return MetricsMsg{
            ReqPerSec:     metrics.ReqPerSec,
            TotalReq:      metrics.TotalReq,
            ActiveWorkers: metrics.ActiveWorkers,
        }
    })
}

func (w *WorkersView) Render(width, height int) string {
    var b strings.Builder

    b.WriteString(styles.TitleStyle.Render("👷 Workers Control"))
    b.WriteString("\n\n")

    // Slider visual
    b.WriteString(fmt.Sprintf("Workers: %d / %d\n", w.slider, w.maxWorkers))
    b.WriteString(w.renderSlider())
    b.WriteString("\n\n")

    // Métricas
    b.WriteString(styles.ItemStyle.Render("📊 Metrics"))
    b.WriteString("\n")
    b.WriteString(fmt.Sprintf("  Active:     %d workers\n", w.activeWorkers))
    b.WriteString(fmt.Sprintf("  Throughput: %.2f req/s\n", w.requestsPerSec))
    b.WriteString(fmt.Sprintf("  Total:      %d requests\n", w.totalRequests))
    b.WriteString("\n\n")

    b.WriteString("◀ Left/Right: Adjust | Enter: Apply | Esc: Back")

    return styles.AppStyle.Render(b.String())
}

func (w *WorkersView) renderSlider() string {
    total := w.maxWorkers
    filled := w.slider

    bar := "["
    for i := 1; i <= total; i++ {
        if i <= filled {
            bar += "█"
        } else {
            bar += "░"
        }
    }
    bar += "]"

    return styles.SelectedItemStyle.Render(bar)
}
```

---

### 5. Dynamic Worker Pool

```go
// processor/worker_pool.go
type WorkerPool struct {
    workers   int
    tasks     chan csvLineMap
    wg        sync.WaitGroup
    mu        sync.RWMutex

    gateway   web.HttpGateway
    logger    logs.Logger

    // Métricas
    metrics   *Metrics
}

type Metrics struct {
    reqCount      atomic.Uint64
    errCount      atomic.Uint64
    startTime     time.Time
    activeWorkers atomic.Int32
}

func (m *Metrics) ReqPerSec() float64 {
    elapsed := time.Since(m.startTime).Seconds()
    if elapsed == 0 {
        return 0
    }
    return float64(m.reqCount.Load()) / elapsed
}

func (p *WorkerPool) SetWorkers(n int) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    if n < 1 || n > MaxWorkers {
        return fmt.Errorf("workers must be between 1 and %d", MaxWorkers)
    }

    delta := n - p.workers

    if delta > 0 {
        // Adicionar workers
        for i := 0; i < delta; i++ {
            p.wg.Add(1)
            go p.worker()
        }
    } else if delta < 0 {
        // Remover workers (fechar canal temporariamente)
        // Implementação complexa - pode usar context cancelation
        // por worker individual
    }

    p.workers = n
    return nil
}

func (p *WorkerPool) GetMetrics() MetricsSnapshot {
    return MetricsSnapshot{
        ReqPerSec:     p.metrics.ReqPerSec(),
        TotalReq:      p.metrics.reqCount.Load(),
        ErrorReq:      p.metrics.errCount.Load(),
        ActiveWorkers: int(p.metrics.activeWorkers.Load()),
    }
}
```

---

### 6. Gateway com Headers Flexíveis

```go
// web/gateway.go
type HttpGateway interface {
    Exec(ctx context.Context, variables map[string]string) (Response, error)
    UpdateConfig(method, urlTemplate, bodyTemplate string, headers map[string]string) error
}

type gatewayImpl struct {
    method       string
    urlTemplate  *template.Template
    bodyTemplate *template.Template
    headers      map[string]string  // ✨ NOVO: flexível
    client       HttpClient

    mu sync.RWMutex  // Proteção para hot-reload
}

func (g *gatewayImpl) UpdateConfig(method, urlTpl, bodyTpl string, headers map[string]string) error {
    g.mu.Lock()
    defer g.mu.Unlock()

    urlTemplate, err := NewTemplate("url", urlTpl)
    if err != nil {
        return err
    }

    bodyTemplate, err := NewTemplate("body", bodyTpl)
    if err != nil {
        return err
    }

    g.method = method
    g.urlTemplate = urlTemplate
    g.bodyTemplate = bodyTemplate
    g.headers = headers

    return nil
}

func (g *gatewayImpl) Exec(ctx context.Context, variables map[string]string) (Response, error) {
    g.mu.RLock()
    defer g.mu.RUnlock()

    uri, err := RenderTemplate(g.urlTemplate, variables)
    if err != nil {
        return Response{}, err
    }

    body, err := RenderTemplate(g.bodyTemplate, variables)
    if err != nil {
        return Response{}, err
    }

    // ✨ Headers flexíveis - pode ter Authorization, Cookie, etc
    headers := make(map[string]string)
    for k, v := range g.headers {
        // Suporta templates em headers também!
        rendered, err := RenderString(v, variables)
        if err != nil {
            headers[k] = v  // Fallback para valor literal
        } else {
            headers[k] = rendered
        }
    }

    return g.client.req(ctx, g.method, uri, headers, []byte(body))
}
```

---

## 🎨 Fluxo de Navegação Proposto

```
┌─────────────────────────────────────────────────────┐
│  🎵 RAPPER - HTTP Load Testing                      │
│  ─────────────────────────────────────────────────  │
│                                                      │
│  [F1] Files   [F2] Logs   [F3] Settings   [F4] Workers
│                                                      │
│  ┌────────────────────────────────────────────────┐ │
│  │                                                 │ │
│  │   VIEW ATUAL (depende de F1-F4)                │ │
│  │                                                 │ │
│  │   - Files:    Lista de CSVs + preview          │ │
│  │   - Logs:     Viewport + métricas              │ │
│  │   - Settings: Formulário de config             │ │
│  │   - Workers:  Slider + métricas em tempo real  │ │
│  │                                                 │ │
│  └────────────────────────────────────────────────┘ │
│                                                      │
│  [Help] Ctrl+C: Quit | Tab: Navigate | Enter: Select
└─────────────────────────────────────────────────────┘
```

### Navegação por Teclas
```
F1 ou Ctrl+F → View Files
F2 ou Ctrl+L → View Logs (durante processamento)
F3 ou Ctrl+S → View Settings
F4 ou Ctrl+W → View Workers

Tab         → Próximo campo (dentro da view)
Shift+Tab   → Campo anterior
Enter       → Confirmar/Selecionar
Esc         → Voltar para view anterior
Ctrl+C      → Sair da aplicação
```

---

## 📦 Consolidação de Arquivos

### Antes (Fragmentado)
```
internal/ui/
├── ui.go              (320 linhas) - Modelo principal
├── states.go          (40 linhas)  - Estados
├── list.go            (80 linhas)  - Lista genérica
├── logo/
│   └── logo.go        (30 linhas)  - Renderização de logo
└── assets/
    └── fonts/...

internal/styles/
└── styles.go          (100 linhas) - Estilos globais

internal/processor/
└── messages.go        (150 linhas) - Mensagens de log
```

### Depois (Consolidado)
```
internal/ui/
├── app.go             (200 linhas) - Modelo principal + navegação
├── commands.go        (100 linhas) - tea.Cmd factories
├── views/
│   ├── files.go       (150 linhas) - View de seleção
│   ├── logs.go        (150 linhas) - View de logs
│   ├── settings.go    (250 linhas) - View de configuração
│   └── workers.go     (200 linhas) - View de workers
├── components/
│   ├── header.go      (50 linhas)  - Header comum
│   ├── form.go        (100 linhas) - Inputs reutilizáveis
│   └── metrics.go     (80 linhas)  - Painel de métricas
└── styles.go          (150 linhas) - Estilos (movido)

internal/processor/
└── messages.go        [EXCLUIR] - Mover para logs/messages.go
```

**Ganhos:**
- ✅ Responsabilidades claras por arquivo
- ✅ Reutilização de componentes
- ✅ Fácil adicionar novas views
- ✅ Menos arquivos pequenos

---

## 🔄 Migrações de Código

### Migração 1: Token → Headers
```diff
# config.yml
- token: "JWT_TOKEN"
+ request:
+   headers:
+     Authorization: "Bearer JWT_TOKEN"
+     X-API-Key: "secret"
```

```diff
// web/gateway.go
- func NewHttpGateway(token, method, urlTpl, bodyTpl string)
+ func NewHttpGateway(method, urlTpl, bodyTpl string, headers map[string]string)

- header := map[string]string{
-     "Authorization": fmt.Sprintf("Bearer %s", g.token),
- }
+ headers := g.headers  // Direto do config
```

### Migração 2: Workers Flag → Config
```diff
# config.yml
+ workers: 4  # Valor inicial
```

```diff
// main.go
- workers := flag.Int("workers", 1, "...")
- csvProcessor := processor.NewProcessor(cfg, gateway, logger, *workers)
+ csvProcessor := processor.NewProcessor(cfg, gateway, logger, cfg.Workers)
```

### Migração 3: UI Monolítica → Multi-View
```diff
// ui/app.go
- type Model struct {
-     viewport  viewport.Model
-     filesList list.Model
- }

+ type Model struct {
+     nav          *Navigation
+     filesView    *FilesView
+     logsView     *LogsView
+     settingsView *SettingsView
+     workersView  *WorkersView
+ }
```

---

## 📊 Comparação: Antes vs Depois

| Aspecto | Antes | Depois |
|---------|-------|--------|
| **Configuração** | Estática (config.yml) | Dinâmica (editável na UI) |
| **Headers** | `Authorization: Bearer <token>` | `map[string]string` flexível |
| **Workers** | Flag CLI (`--workers=4`) | Slider interativo (runtime) |
| **Views** | 1 view (logs + lista) | 4 views (files/logs/settings/workers) |
| **Navegação** | Apenas seleção de arquivo | F1-F4 + Tab + Esc |
| **Hot-reload** | Não suportado | `config.Manager` com listeners |
| **Métricas** | Apenas contadores | req/s, workers ativos, gráficos |
| **UX** | Limitada | Rica e interativa |

---

## ✅ Checklist de Implementação

### Fase 1: Fundação (Refatoração Base)
- [ ] Criar `config.Manager` com suporte a hot-reload
- [ ] Atualizar `Config` para usar `headers: map[string]string`
- [ ] Refatorar `HttpGateway` para aceitar headers flexíveis
- [ ] Migrar `internal/styles` para `internal/ui/styles.go`
- [ ] Criar estrutura `ui/views/` e `ui/components/`

### Fase 2: Multi-View Architecture
- [ ] Implementar `ui/navigation.go` com histórico
- [ ] Criar `FilesView` (migrar código existente)
- [ ] Criar `LogsView` (migrar código existente)
- [ ] Atualizar `Model` principal para delegar para views
- [ ] Adicionar key bindings para F1-F4

### Fase 3: Settings View
- [ ] Criar `SettingsView` com formulário
- [ ] Implementar `HeadersEditor` component
- [ ] Adicionar validação de templates
- [ ] Conectar `SettingsView` com `config.Manager`
- [ ] Adicionar persistência de configuração

### Fase 4: Workers Dinâmicos
- [ ] Criar `WorkerPool` com `SetWorkers()`
- [ ] Implementar `Metrics` com req/s, contadores
- [ ] Criar `WorkersView` com slider
- [ ] Adicionar tick para atualização de métricas
- [ ] Testar aumento/diminuição de workers em runtime

### Fase 5: Polimento
- [ ] Adicionar animações de transição entre views
- [ ] Melhorar feedback visual (spinners, progress bars)
- [ ] Adicionar profiles (salvar/carregar múltiplas configs)
- [ ] Documentar novos recursos no README
- [ ] Atualizar testes com novos mocks

---

## 🧪 Testes a Adicionar

```go
// config/manager_test.go
func TestManager_Update(t *testing.T)
func TestManager_OnChange_NotifiesListeners(t *testing.T)
func TestManager_Save_PersistsToYAML(t *testing.T)

// ui/navigation_test.go
func TestNavigation_Push(t *testing.T)
func TestNavigation_Back(t *testing.T)

// ui/views/settings_test.go
func TestSettingsView_SaveConfig(t *testing.T)
func TestHeadersEditor_AddHeader(t *testing.T)

// processor/worker_pool_test.go
func TestWorkerPool_SetWorkers_Increase(t *testing.T)
func TestWorkerPool_SetWorkers_Decrease(t *testing.T)
func TestMetrics_ReqPerSec(t *testing.T)

// web/gateway_test.go
func TestGateway_UpdateConfig_HotReload(t *testing.T)
func TestGateway_Exec_FlexibleHeaders(t *testing.T)
```

---

## 📚 Recursos e Referências

### Padrões Bubbletea
- [Elm Architecture](https://guide.elm-lang.org/architecture/)
- [Bubbletea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)
- [Multi-View Pattern](https://github.com/charmbracelet/bubbletea/tree/master/examples/views)

### Componentes Úteis
- [bubbles/textinput](https://github.com/charmbracelet/bubbles/tree/master/textinput) - Input fields
- [bubbles/textarea](https://github.com/charmbracelet/bubbles/tree/master/textarea) - Multiline input
- [bubbles/paginator](https://github.com/charmbracelet/bubbles/tree/master/paginator) - Paginação
- [bubbles/progress](https://github.com/charmbracelet/bubbles/tree/master/progress) - Progress bars

### Exemplos de TUIs Complexas
- [Glow](https://github.com/charmbracelet/glow) - Markdown reader
- [Soft Serve](https://github.com/charmbracelet/soft-serve) - Git server
- [VHS](https://github.com/charmbracelet/vhs) - Terminal recorder

---

## 🎯 Próximos Passos

1. **Revisar este plano** e ajustar prioridades
2. **Criar branch de refatoração**: `git checkout -b refactor/multi-view-tui`
3. **Implementar Fase 1** (fundação) primeiro
4. **Testar cada fase** antes de avançar
5. **Documentar mudanças** no README conforme implementa

---

## 💡 Melhorias Futuras (Opcional)

- [ ] **Profiles**: Salvar múltiplas configurações (prod, dev, staging)
- [ ] **Graphs**: Gráfico de linha para req/s em tempo real
- [ ] **Themes**: Tema claro/escuro
- [ ] **Export**: Exportar resultados para JSON, CSV, HTML
- [ ] **Retry Logic**: Configurar retry automático de requests
- [ ] **Rate Limiting**: Limitar req/s globalmente
- [ ] **WebSocket**: Suporte para websocket requests
- [ ] **Auth Wizard**: Wizard para OAuth2, API Keys, etc

---

**Autor:** Claude (Anthropic)
**Data:** 2026-01-14
**Versão:** 1.0.0
