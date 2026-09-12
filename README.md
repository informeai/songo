# songo

Gerador de efeitos sonoros estilo 8-bit em Go puro — sem cgo, sem
dependências externas. Sintetiza os sons na hora (onda quadrada,
triangular, ruído, FM synthesis) e exporta `.wav`.

## Requisitos

- Go 1.21+

## Instalação

```sh
go build -o songo ./cmd/songo
```

Ou rode direto sem buildar, usando `go run ./cmd/songo ...` (assim como
nos exemplos abaixo).

## Uso

```sh
# lista os presets disponíveis
songo list

# gera um preset específico (salva como <preset>.wav no diretório atual)
songo generate coin

# gera um preset com nome de arquivo customizado
songo generate laser -o meu_laser.wav

# gera e já toca o som (afplay no macOS, paplay/aplay/ffplay no Linux,
# PowerShell no Windows)
songo generate powerup -p

# gera todos os presets de uma vez (salva em ./sounds/ por padrão)
songo all

# gera todos os presets numa pasta customizada
songo all -o assets/sfx

# gera todos e toca cada um em sequência, útil pra navegar pelos sons
songo all -p

# gera a partir de um script da DSL própria do songo (veja a seção abaixo)
songo run examples/coin.sfx -p
```

## Presets disponíveis

| Nome         | Descrição                                                             |
|--------------|------------------------------------------------------------------------|
| `coin`       | Arpejo rápido subindo, tipo moeda coletada                              |
| `jump`       | Sweep de frequência subindo, tipo pulo                                  |
| `laser`      | Sweep descendente em onda quadrada fina, tipo tiro                      |
| `laser_lofi` | O `laser` com bitcrush, timbre mais quebrado/lo-fi                      |
| `explosion`  | Ruído com decay longo                                                   |
| `cannon`     | Ruído + filtro passa-baixa dinâmico + eco, mais denso que a explosão    |
| `bass`       | Onda triangular grave, tipo canal de baixo do NES                       |
| `powerup`    | Arpejo ascendente terminando num tom sustentado                         |
| `gameover`   | Sequência descendente lenta terminando num grave sustentado             |
| `fm_bell`    | Sino metálico via FM synthesis                                          |

## DSL: gerando sons com scripts `.sfx`

Além dos presets prontos, o songo interpreta uma linguagem declarativa
própria: um arquivo de texto com uma instrução por linha. **Geradores**
(`square`, `triangle`, `noise`, `sweep`, `fm`) concatenam trechos de áudio
ao buffer; **processadores** (`envelope`, `lowpass`, `sweep_filter`,
`delay`, `bitcrush`) transformam o buffer inteiro já gerado até ali.
Linhas em branco e começando com `#` são ignoradas.

```sh
# gera a partir de um script .sfx (salva como <nome-do-script>.wav)
songo run examples/coin.sfx

# com saída customizada e tocando o som
songo run examples/cannon.sfx -o cannon.wav -p
```

Exemplo (`examples/coin.sfx`):

```
# moeda coletada: arpejo rápido subindo (B5 -> E6)
square 988 0.05 duty=0.5 amp=0.4
square 1319 0.15 duty=0.5 amp=0.4
envelope 0.001 0.02 0.6 0.05
```

Exemplo com processadores (`examples/cannon.sfx`):

```
# canhão: ruído + filtro fechando ao longo do tempo + eco
noise 0.5 amp=0.8
sweep_filter 4000 100
envelope 0.001 0.05 0.4 0.3
delay 0.08 0.4 3
```

### Comandos da DSL

| Comando        | Argumentos posicionais                  | Flags (com padrão)     | Tipo       |
|----------------|------------------------------------------|-------------------------|------------|
| `square`       | `freq dur`                                | `duty=0.5` `amp=1.0`    | gerador    |
| `triangle`     | `freq dur`                                | `amp=1.0`               | gerador    |
| `noise`        | `dur`                                     | `amp=1.0`               | gerador    |
| `sweep`        | `freqInicial freqFinal dur`               | `duty=0.5` `amp=1.0`    | gerador    |
| `fm`           | `freqCarrier freqModulador indiceMod dur` | `amp=1.0`               | gerador    |
| `envelope`     | `attack decay sustain release`            | —                       | processador |
| `lowpass`      | `cutoffHz`                                | —                       | processador |
| `sweep_filter` | `cutoffInicial cutoffFinal`               | —                       | processador |
| `delay`        | `delaySegundos feedback repeticoes`       | —                       | processador |
| `bitcrush`     | `bits downsample`                         | —                       | processador |

Mais exemplos em [`examples/`](examples/).

## Estrutura do projeto

```
songo/
├── synth/     # motor de síntese reutilizável (ondas, envelope ADSR,
│              # filtros, delay, bitcrush, FM, escrita de WAV)
├── presets/   # receitas prontas de efeitos sonoros em cima do synth
├── script/    # interpretador da DSL declarativa (.sfx)
├── examples/  # scripts .sfx de exemplo
└── cmd/songo/ # CLI
```

## Usando como biblioteca

Os pacotes `synth` e `presets` podem ser importados diretamente em outro
projeto Go para gerar sons em tempo de build ou em runtime:

```go
import (
    "github.com/informeai/songo/presets"
    "github.com/informeai/songo/synth"
)

samples := presets.Coin()
synth.WriteWAV("coin.wav", samples)
```

## Licença

MIT — veja [LICENSE](LICENSE).
