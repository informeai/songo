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

# gera uma composição com múltiplas vozes (veja a seção "Composição")
songo song examples/racer_theme.song -p

# transcreve uma melodia monofônica de um WAV pra um .song (veja abaixo)
songo transcribe melodia.wav -tempo 120
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

## Composição: músicas com múltiplas vozes (arquivos `.song`)

Enquanto `.sfx` descreve um efeito sonoro único, `.song` descreve uma
composição inteira: várias vozes (baixo, lead, bateria...) tocando em
paralelo, cada uma com seu instrumento e sua sequência de notas. As notas
usam notação científica (`A4`, `C#5`, `Eb3`) ou `.`/`x` pra pausa/pancada
(em vozes de `noise`, sem altura definida), com duração em `w` (semibreve),
`h` (mínima), `q` (semínima), `e` (colcheia) ou `s` (semicolcheia) — some
um `.` no final pra ponteada (ex.: `q.`).

```sh
songo song examples/racer_theme.song -p
```

Exemplo simplificado (veja o arquivo completo em
[`examples/racer_theme.song`](examples/racer_theme.song): um tema de
corrida 8-bit original com estrutura completa — intro construindo
tensão, tema principal repetido 2x, ponte com progressão diferente,
reprise do tema e outro final, 4 vozes, 32 segundos):

```
tempo 150

voice bass triangle amp=0.6
A2 e
E3 e

voice lead square duty=0.25 amp=0.3
A4 e
C5 e

voice kick noise amp=0.6
x q
. q
```

Instrumentos disponíveis pra `voice`: `square` (flag `duty`), `triangle`,
`sine`, `sawtooth`, `pluck`, `fm` (flags `mod_ratio`, `mod_index`) e
`noise` (percussão, sem altura). Todas as vozes são mixadas e o resultado
é normalizado automaticamente (headroom de 0.9) pra não estourar quando
várias vozes tocam ao mesmo tempo.

### Repetindo trechos com `repeat`/`end`

Pra trilhas mais longas, em vez de reescrever o mesmo trecho de notas
várias vezes, envolva-o num bloco `repeat <vezes>` / `end` dentro da
`voice`:

```
voice hihat noise amp=0.15
repeat 16
. e
x e
end
```

Isso equivale a escrever `. e` / `x e` sozinhos 16 vezes seguidas. Blocos
`repeat` não podem ser aninhados nem conter `voice`/`tempo`/outro `repeat`
— só linhas de nota.

> Por que uma DSL de composição em vez de recriar trilhas de jogos
> existentes: transcrever nota por nota a trilha de um jogo específico
> reproduziria uma composição protegida por direitos autorais. O songo
> te dá as ferramentas (sequenciador + instrumentos 8-bit) pra compor
> algo original no mesmo estilo.

### Transcrevendo uma melodia de um WAV com `songo transcribe`

Além de compor manualmente, dá pra gerar um `.song` automaticamente a
partir de áudio: `songo transcribe` detecta a altura (pitch) por
autocorrelação e converte pra uma sequência de notas.

```sh
songo transcribe entrada.wav -o melodia.song -tempo 120 -instrument square -voice lead
```

| Flag           | Padrão    | Descrição                                                |
|----------------|-----------|-----------------------------------------------------------|
| `-o`           | `<entrada>.song` | Caminho do `.song` gerado                          |
| `-tempo`       | `120`     | BPM assumido pra quantizar as durações detectadas em w/h/q/e/s |
| `-instrument`  | `square`  | Instrumento usado na voice gerada                          |
| `-voice`       | `melody`  | Nome da voice gerada                                       |

**Limitações importantes:**
- Só aceita **WAV mono, 16-bit, 44100Hz** (sem decodificador de MP3
  embutido, pra manter o projeto sem dependências externas). Converta
  com `ffmpeg -i entrada.mp3 -ac 1 -ar 44100 -sample_fmt s16 entrada.wav`.
- Detecta **uma melodia monofônica** (uma nota de cada vez — um
  instrumento solo, um assobio, um canto). Não separa uma faixa com
  vários instrumentos tocando ao mesmo tempo (baixo+bateria+lead juntos)
  em vozes diferentes — isso é um problema de pesquisa (separação de
  fontes polifônica) que a implementação atual não resolve.
- As durações são quantizadas pro BPM informado; se o BPM real do áudio
  for diferente do passado em `-tempo`, os valores de w/h/q/e/s podem
  sair um pouco "torcidos" (funciona, mas talvez precise de ajuste manual
  no `.song` gerado).

Validado com um roundtrip: uma escala sintetizada com `songo song` foi
transcrita de volta e recuperou as notas exatamente.

## Estrutura do projeto

```
songo/
├── synth/      # motor de síntese reutilizável (ondas, envelope ADSR,
│               # filtros, delay, bitcrush, FM, leitura/escrita de WAV)
├── presets/    # receitas prontas de efeitos sonoros em cima do synth
├── script/     # interpretador da DSL declarativa de efeitos (.sfx)
├── song/       # interpretador da DSL de composição multi-voz (.song)
├── transcribe/ # detecção de pitch (WAV -> .song de uma voz)
├── examples/   # scripts .sfx e .song de exemplo
└── cmd/songo/  # CLI
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
