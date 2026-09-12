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
(`square`, `triangle`, `noise`, `sweep`, `fm`, `sine`, `sawtooth`, `pluck`)
concatenam trechos de áudio ao buffer; **processadores** (`envelope`,
`lowpass`, `sweep_filter`, `delay`, `bitcrush`, `highpass`, `resonant`,
`distortion`, `reverb`, `vibrato`, `tremolo`, `gain`, `normalize`,
`reverse`) transformam o buffer inteiro já gerado até ali. Linhas em
branco e começando com `#` são ignoradas.

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

**Geradores** (concatenam áudio ao buffer):

| Comando    | Argumentos posicionais                    | Flags (com padrão) | Descrição                                    |
|------------|--------------------------------------------|----------------------|-----------------------------------------------|
| `square`   | `freq dur`                                  | `duty=0.5` `amp=1.0` | Onda quadrada clássica de console 8-bit        |
| `triangle` | `freq dur`                                  | `amp=1.0`            | Onda triangular (grave, tipo baixo do NES)     |
| `noise`    | `dur`                                       | `amp=1.0`            | Ruído branco                                   |
| `sweep`    | `freqInicial freqFinal dur`                 | `duty=0.5` `amp=1.0` | Onda quadrada com frequência variando no tempo |
| `fm`       | `freqCarrier freqModulador indiceMod dur`   | `amp=1.0`            | FM synthesis (sinos, timbres metálicos)        |
| `sine`     | `freq dur`                                  | `amp=1.0`            | Tom puro (oscilador senoidal)                  |
| `sawtooth` | `freq dur`                                  | `amp=1.0`            | Dente de serra, rica em harmônicos             |
| `pluck`    | `freq dur`                                  | `amp=1.0`            | Corda dedilhada (Karplus-Strong)               |

**Processadores** (transformam o buffer inteiro já gerado):

| Comando        | Argumentos posicionais              | Flags (com padrão) | Descrição                                          |
|----------------|---------------------------------------|----------------------|------------------------------------------------------|
| `envelope`     | `attack decay sustain release`        | —                    | ADSR                                                  |
| `lowpass`      | `cutoffHz`                            | —                    | Filtro passa-baixa                                    |
| `sweep_filter` | `cutoffInicial cutoffFinal`           | —                    | Passa-baixa com cutoff variando no tempo ("wobble")   |
| `delay`        | `delaySegundos feedback repeticoes`   | —                    | Eco                                                   |
| `bitcrush`     | `bits downsample`                     | —                    | Redução de resolução, timbre lo-fi                    |
| `highpass`     | `cutoffHz`                            | —                    | Filtro passa-alta                                     |
| `resonant`     | `cutoffHz q`                          | —                    | Passa-baixa ressonante (state-variable), "uivo" analógico |
| `distortion`   | `drive`                               | —                    | Saturação/overdrive (tanh)                            |
| `reverb`       | `mix` (0 a 1)                         | —                    | Reverberação estilo Schroeder                         |
| `vibrato`      | `rateHz depthMs`                      | —                    | Modulação de pitch (LFO)                              |
| `tremolo`      | `rateHz depth` (depth 0 a 1)          | —                    | Modulação de volume (LFO)                             |
| `gain`         | `factor`                              | —                    | Multiplica a amplitude                                |
| `normalize`    | —                                      | `level=1.0`          | Escala o pico pra `level`                             |
| `reverse`      | —                                      | —                    | Inverte o áudio (toca de trás pra frente)             |

Mais exemplos em [`examples/`](examples/), incluindo `pluck.sfx`,
`synth_lead.sfx` (sawtooth + resonant), `haunted_bell.sfx` (vibrato +
reverb), `power_chord.sfx` (distortion + tremolo + normalize) e
`reverse_cymbal.sfx`.

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
