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
- [ ] Suporte a múltiplos profiles (api1.yml, production.yml, etc)
- [ ] Descoberta automática de arquivos .yml no diretório
- [ ] Troca de profile em runtime com Ctrl+P
- [ ] Salvar edições no arquivo YAML correto

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
│   ├── manager.go          # [NOVO] Gerenciamento em runtime
│   └── profile.go          # [NOVO] Gerenciamento de múltiplos profiles
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

### 1.5. Profile Manager - Múltiplas Configurações

#### Conceito
Permitir que o usuário:
- Descubra automaticamente todos os arquivos `.yml` no diretório de execução
- Selecione qual profile quer usar (api1.yml, api2.yml, production.yml, etc)
- Alterne entre profiles durante a execução
- Edite o profile ativo na Settings View
- Salve as alterações no arquivo YAML correto

#### Estrutura de Profile

```go
// config/profile.go
type Profile struct {
    Name     string   // Nome do profile (ex: "api1", "production")
    FilePath string   // Caminho do arquivo (ex: "./api1.yml")
    Config   *Config  // Configuração carregada
}

type ProfileManager interface {
    // Descobre todos os arquivos .yml no diretório
    Discover(dir string) ([]Profile, error)

    // Lista todos os profiles disponíveis
    List() []Profile

    // Obtém o profile ativo
    GetActive() *Profile

    // Troca o profile ativo
    SetActive(name string) error

    // Salva o profile ativo no arquivo YAML
    Save() error

    // Atualiza a configuração do profile ativo
    UpdateActive(cfg *Config) error
}

type profileManagerImpl struct {
    profiles     []Profile
    activeIndex  int
    configLoader *Loader
    mu           sync.RWMutex
}
```

#### Implementação - Descoberta de Profiles

```go
// config/profile.go
func (pm *profileManagerImpl) Discover(dir string) ([]Profile, error) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    // Busca todos os arquivos .yml no diretório
    files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
    if err != nil {
        return nil, fmt.Errorf("failed to glob yml files: %w", err)
    }

    // Também busca .yaml
    yamlFiles, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
    if err != nil {
        return nil, fmt.Errorf("failed to glob yaml files: %w", err)
    }

    files = append(files, yamlFiles...)

    if len(files) == 0 {
        return nil, fmt.Errorf("no .yml or .yaml files found in %s", dir)
    }

    profiles := make([]Profile, 0, len(files))

    for _, filePath := range files {
        // Carrega o arquivo
        cfg, err := pm.configLoader.Load(filePath)
        if err != nil {
            // Ignora arquivos inválidos mas loga o erro
            log.Printf("Skipping invalid config file %s: %v", filePath, err)
            continue
        }

        // Extrai nome do arquivo (sem extensão)
        baseName := filepath.Base(filePath)
        name := strings.TrimSuffix(baseName, filepath.Ext(baseName))

        profiles = append(profiles, Profile{
            Name:     name,
            FilePath: filePath,
            Config:   cfg,
        })
    }

    if len(profiles) == 0 {
        return nil, fmt.Errorf("no valid config files found")
    }

    pm.profiles = profiles
    pm.activeIndex = 0  // Primeiro profile por padrão

    return profiles, nil
}

func (pm *profileManagerImpl) List() []Profile {
    pm.mu.RLock()
    defer pm.mu.RUnlock()

    return pm.profiles
}

func (pm *profileManagerImpl) GetActive() *Profile {
    pm.mu.RLock()
    defer pm.mu.RUnlock()

    if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
        return nil
    }

    return &pm.profiles[pm.activeIndex]
}

func (pm *profileManagerImpl) SetActive(name string) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    for i, profile := range pm.profiles {
        if profile.Name == name {
            pm.activeIndex = i
            return nil
        }
    }

    return fmt.Errorf("profile %s not found", name)
}

func (pm *profileManagerImpl) UpdateActive(cfg *Config) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
        return fmt.Errorf("no active profile")
    }

    pm.profiles[pm.activeIndex].Config = cfg
    return nil
}

func (pm *profileManagerImpl) Save() error {
    pm.mu.RLock()
    defer pm.mu.RUnlock()

    if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
        return fmt.Errorf("no active profile")
    }

    active := &pm.profiles[pm.activeIndex]

    // Serializa para YAML
    data, err := yaml.Marshal(active.Config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    // Salva no arquivo original
    err = os.WriteFile(active.FilePath, data, 0644)
    if err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }

    return nil
}
```

#### Integração com Config Manager

```go
// config/manager.go (ATUALIZADO)
type Manager interface {
    Get() *Config
    Update(cfg *Config) error
    Save() error
    OnChange(callback func(*Config))

    // ✨ NOVO: Suporte a profiles
    GetProfileManager() ProfileManager
}

type managerImpl struct {
    profileMgr ProfileManager
    listeners  []func(*Config)
    mu         sync.RWMutex
}

func NewManager(dir string) (Manager, error) {
    loader := NewLoader()
    profileMgr := NewProfileManager(loader)

    // Descobre profiles no diretório
    _, err := profileMgr.Discover(dir)
    if err != nil {
        return nil, err
    }

    return &managerImpl{
        profileMgr: profileMgr,
        listeners:  make([]func(*Config), 0),
    }, nil
}

func (m *managerImpl) Get() *Config {
    active := m.profileMgr.GetActive()
    if active == nil {
        return nil
    }
    return active.Config
}

func (m *managerImpl) Update(cfg *Config) error {
    if err := m.profileMgr.UpdateActive(cfg); err != nil {
        return err
    }

    // Notifica listeners
    m.mu.RLock()
    defer m.mu.RUnlock()

    for _, listener := range m.listeners {
        listener(cfg)
    }

    return nil
}

func (m *managerImpl) Save() error {
    return m.profileMgr.Save()
}

func (m *managerImpl) GetProfileManager() ProfileManager {
    return m.profileMgr
}
```

#### UI - Profile Selector na Settings View

```go
// ui/views/settings.go (ATUALIZADO)
type SettingsView struct {
    configMgr config.Manager

    // Profile selector
    profileSelector  list.Model     // ✨ NOVO: Lista de profiles
    showingProfiles  bool           // Modal de seleção aberto?

    // Form inputs
    methodInput      textinput.Model
    urlInput         textinput.Model
    bodyInput        textarea.Model
    headersEditor    *HeadersEditor

    focusIndex int
}

func NewSettingsView(configMgr config.Manager) *SettingsView {
    // Cria lista de profiles
    profileMgr := configMgr.GetProfileManager()
    profiles := profileMgr.List()

    items := make([]list.Item, len(profiles))
    for i, p := range profiles {
        items[i] = profileItem{
            name:     p.Name,
            filePath: p.FilePath,
            active:   i == 0, // Marca o ativo
        }
    }

    profileList := list.New(items, list.NewDefaultDelegate(), 0, 0)
    profileList.Title = "Select Profile"

    return &SettingsView{
        configMgr:       configMgr,
        profileSelector: profileList,
        showingProfiles: false,
        // ... resto da inicialização
    }
}

func (s *SettingsView) Update(msg tea.Msg) tea.Cmd {
    // Se modal de profiles está aberto
    if s.showingProfiles {
        switch msg := msg.(type) {
        case tea.KeyMsg:
            switch msg.String() {
            case "enter":
                // Troca de profile
                selected := s.profileSelector.SelectedItem().(profileItem)
                return s.switchProfile(selected.name)

            case "esc":
                s.showingProfiles = false
                return nil
            }
        }

        // Delega para a lista
        var cmd tea.Cmd
        s.profileSelector, cmd = s.profileSelector.Update(msg)
        return cmd
    }

    // Navegação normal
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+p":
            // Abre modal de profiles
            s.showingProfiles = true
            return nil

        case "ctrl+s":
            // Salva configuração no arquivo ativo
            return s.saveConfig()

        case "tab":
            s.focusIndex = (s.focusIndex + 1) % 5  // Agora são 5 campos
            s.updateFocus()
        }
    }

    return nil
}

func (s *SettingsView) switchProfile(name string) tea.Cmd {
    return func() tea.Msg {
        profileMgr := s.configMgr.GetProfileManager()

        if err := profileMgr.SetActive(name); err != nil {
            return ProfileErrorMsg{err}
        }

        // Recarrega inputs com nova config
        cfg := s.configMgr.Get()
        s.loadConfigToInputs(cfg)

        return ProfileSwitchedMsg{name: name}
    }
}

func (s *SettingsView) saveConfig() tea.Cmd {
    return func() tea.Msg {
        cfg := s.configMgr.Get()

        // Atualiza com valores dos inputs
        cfg.Request.Method = s.methodInput.Value()
        cfg.Request.URLTemplate = s.urlInput.Value()
        cfg.Request.BodyTemplate = s.bodyInput.Value()
        cfg.Request.Headers = s.headersEditor.headers

        // Atualiza em memória
        if err := s.configMgr.Update(cfg); err != nil {
            return ConfigErrorMsg{err}
        }

        // Salva no arquivo YAML ativo
        if err := s.configMgr.Save(); err != nil {
            return ConfigErrorMsg{err}
        }

        profileMgr := s.configMgr.GetProfileManager()
        active := profileMgr.GetActive()

        return ConfigSavedMsg{
            profileName: active.Name,
            filePath:    active.FilePath,
        }
    }
}

func (s *SettingsView) Render(width, height int) string {
    // Se modal de profiles está aberto, renderiza por cima
    if s.showingProfiles {
        return s.renderProfileSelector(width, height)
    }

    var b strings.Builder

    // Header com profile ativo
    profileMgr := s.configMgr.GetProfileManager()
    active := profileMgr.GetActive()

    title := fmt.Sprintf("⚙️  Settings - Profile: %s", active.Name)
    b.WriteString(styles.TitleStyle.Render(title))
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

    b.WriteString("Ctrl+P: Switch Profile | Ctrl+S: Save | Esc: Back")

    return styles.AppStyle.Render(b.String())
}

func (s *SettingsView) renderProfileSelector(width, height int) string {
    // Modal centralizado com lista de profiles
    profileMgr := s.configMgr.GetProfileManager()
    profiles := profileMgr.List()

    var b strings.Builder
    b.WriteString("╔═══════════════════════════════════════╗\n")
    b.WriteString("║        Select Configuration Profile        ║\n")
    b.WriteString("╠═══════════════════════════════════════╣\n")

    for i, p := range profiles {
        active := profileMgr.GetActive()
        marker := " "
        if p.Name == active.Name {
            marker = "●"  // Bullet para o ativo
        }

        line := fmt.Sprintf("║ %s %s", marker, p.Name)
        padding := 40 - len(line) - 1
        line += strings.Repeat(" ", padding) + "║\n"
        b.WriteString(line)
    }

    b.WriteString("╚═══════════════════════════════════════╝\n")
    b.WriteString("Enter: Select | Esc: Cancel")

    return lipgloss.Place(
        width, height,
        lipgloss.Center, lipgloss.Center,
        lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("#d6acff")).
            Padding(1, 2).
            Render(b.String()),
    )
}

// Profile list item
type profileItem struct {
    name     string
    filePath string
    active   bool
}

func (i profileItem) FilterValue() string { return i.name }
func (i profileItem) Title() string {
    if i.active {
        return "● " + i.name  // Marcador de ativo
    }
    return "  " + i.name
}
func (i profileItem) Description() string { return i.filePath }
```

#### Exemplo de Estrutura de Diretório

```
/home/user/my-project/
├── rapper                    # Binário
├── api1.yml                  # Profile 1 - API interna
├── api2.yml                  # Profile 2 - API externa
├── production.yml            # Profile 3 - Produção
├── staging.yml               # Profile 4 - Staging
├── data/
│   ├── users.csv
│   └── orders.csv
└── output/
    └── results.json
```

**api1.yml:**
```yaml
request:
  method: POST
  url_template: "http://localhost:8080/users/{{.id}}"
  body_template: |
    {"name": "{{.name}}"}
  headers:
    Authorization: "Bearer dev-token-123"
    Content-Type: "application/json"

csv:
  fields: [id, name]
  separator: ","

workers: 2
```

**production.yml:**
```yaml
request:
  method: POST
  url_template: "https://api.production.com/users/{{.id}}"
  body_template: |
    {"name": "{{.name}}", "email": "{{.email}}"}
  headers:
    Authorization: "Bearer prod-token-xyz"
    X-API-Key: "production-key"
    Cookie: "session=abc123"

csv:
  fields: [id, name, email]
  separator: ","

workers: 8
```

#### Fluxo de Uso

```
1. Usuário inicia o app no diretório com múltiplos .yml
   ↓
2. ProfileManager.Discover() encontra: api1.yml, api2.yml, production.yml
   ↓
3. Primeiro profile (api1) é carregado automaticamente
   ↓
4. Usuário aperta F3 → View Settings
   ↓
5. Header mostra: "⚙️ Settings - Profile: api1"
   ↓
6. Usuário aperta Ctrl+P → Modal de profiles abre
   ↓
7. Lista mostra:
      ● api1        (./api1.yml)
        api2        (./api2.yml)
        production  (./production.yml)
   ↓
8. Usuário seleciona "production" e aperta Enter
   ↓
9. ProfileManager.SetActive("production")
   ↓
10. Inputs são atualizados com config de production.yml
   ↓
11. Usuário edita URL template, adiciona header
   ↓
12. Usuário aperta Ctrl+S
   ↓
13. ConfigManager.Save() salva no production.yml
   ↓
14. Gateway recebe hot-reload com novos headers
   ↓
15. Próximas requisições usam nova configuração
```

#### Mensagens Tea (Tea.Msg)

```go
// ui/views/settings.go
type ProfileSwitchedMsg struct {
    name string
}

type ProfileErrorMsg struct {
    err error
}

type ConfigSavedMsg struct {
    profileName string
    filePath    string
}

type ConfigErrorMsg struct {
    err error
}
```

#### Visual do Profile Selector

```
┌─────────────────────────────────────────┐
│                                         │
│   ╔═══════════════════════════════════╗ │
│   ║  Select Configuration Profile     ║ │
│   ╠═══════════════════════════════════╣ │
│   ║ ● api1           (./api1.yml)     ║ │
│   ║   api2           (./api2.yml)     ║ │
│   ║   production     (./production.yml)║ │
│   ║   staging        (./staging.yml)  ║ │
│   ╚═══════════════════════════════════╝ │
│                                         │
│   Enter: Select | Esc: Cancel          │
│                                         │
└─────────────────────────────────────────┘
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
│  Profile: production.yml                            │
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
│  │   - Settings: Formulário de config + Ctrl+P    │ │
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

Ctrl+P      → Switch Profile (abre modal de seleção)
Ctrl+S      → Save Config (salva no arquivo YAML ativo)

Tab         → Próximo campo (dentro da view)
Shift+Tab   → Campo anterior
Enter       → Confirmar/Selecionar
Esc         → Voltar para view anterior / Fechar modal
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
| **Configuração** | Estática (config.yml único) | Dinâmica (editável na UI) |
| **Profiles** | Não suportado | Múltiplos .yml (api1, prod, staging) |
| **Troca de Config** | Restart obrigatório | Hot-swap com Ctrl+P |
| **Headers** | `Authorization: Bearer <token>` | `map[string]string` flexível |
| **Workers** | Flag CLI (`--workers=4`) | Slider interativo (runtime) |
| **Views** | 1 view (logs + lista) | 4 views (files/logs/settings/workers) |
| **Navegação** | Apenas seleção de arquivo | F1-F4 + Tab + Esc |
| **Hot-reload** | Não suportado | `config.Manager` com listeners |
| **Métricas** | Apenas contadores | req/s, workers ativos, gráficos |
| **Persistência** | Manual (editar arquivo) | Salvar na UI (Ctrl+S) |
| **UX** | Limitada | Rica e interativa |

---

## ✅ Checklist de Implementação

### Fase 1: Fundação (Refatoração Base)
- [ ] Criar `config.Manager` com suporte a hot-reload
- [ ] Criar `config.ProfileManager` para múltiplos YAMLs
- [ ] Implementar descoberta automática de arquivos .yml
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

### Fase 3: Settings View com Profile Management
- [ ] Criar `SettingsView` com formulário
- [ ] Implementar `HeadersEditor` component
- [ ] Adicionar Profile Selector modal (Ctrl+P)
- [ ] Implementar troca de profile em runtime
- [ ] Exibir profile ativo no header da Settings View
- [ ] Adicionar validação de templates
- [ ] Conectar `SettingsView` com `config.Manager`
- [ ] Implementar persistência no arquivo YAML correto
- [ ] Testar hot-reload após troca de profile

### Fase 4: Workers Dinâmicos
- [ ] Criar `WorkerPool` com `SetWorkers()`
- [ ] Implementar `Metrics` com req/s, contadores
- [ ] Criar `WorkersView` com slider
- [ ] Adicionar tick para atualização de métricas
- [ ] Testar aumento/diminuição de workers em runtime

### Fase 5: Polimento
- [ ] Adicionar animações de transição entre views
- [ ] Melhorar feedback visual (spinners, progress bars)
- [ ] Adicionar toast notifications (config saved, profile switched)
- [ ] Melhorar visual do profile selector modal
- [ ] Documentar novos recursos no README
- [ ] Atualizar testes com novos mocks
- [ ] Adicionar testes para ProfileManager

---

## 🧪 Testes a Adicionar

```go
// config/manager_test.go
func TestManager_Update(t *testing.T)
func TestManager_OnChange_NotifiesListeners(t *testing.T)
func TestManager_Save_PersistsToYAML(t *testing.T)
func TestManager_GetProfileManager(t *testing.T)

// config/profile_test.go
func TestProfileManager_Discover(t *testing.T)
func TestProfileManager_Discover_NoFiles(t *testing.T)
func TestProfileManager_Discover_InvalidYAML(t *testing.T)
func TestProfileManager_SetActive(t *testing.T)
func TestProfileManager_GetActive(t *testing.T)
func TestProfileManager_UpdateActive(t *testing.T)
func TestProfileManager_Save(t *testing.T)
func TestProfileManager_List(t *testing.T)

// ui/navigation_test.go
func TestNavigation_Push(t *testing.T)
func TestNavigation_Back(t *testing.T)

// ui/views/settings_test.go
func TestSettingsView_SaveConfig(t *testing.T)
func TestSettingsView_SwitchProfile(t *testing.T)
func TestSettingsView_ProfileSelectorModal(t *testing.T)
func TestSettingsView_SaveToDifferentProfile(t *testing.T)
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
**Versão:** 2.0.0 - Profile Management Edition
