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

# gera todos os presets de uma vez (salva em ./sounds/ por padrão)
songo all

# gera todos os presets numa pasta customizada
songo all -o assets/sfx
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

## Estrutura do projeto

```
songo/
├── synth/     # motor de síntese reutilizável (ondas, envelope ADSR,
│              # filtros, delay, bitcrush, FM, escrita de WAV)
├── presets/   # receitas prontas de efeitos sonoros em cima do synth
└── cmd/songo/ # CLI
```

## Usando como biblioteca

Os pacotes `synth` e `presets` podem ser importados diretamente em outro
projeto Go para gerar sons em tempo de build ou em runtime:

```go
import (
    "songo/presets"
    "songo/synth"
)

samples := presets.Coin()
synth.WriteWAV("coin.wav", samples)
```

## Licença

MIT — veja [LICENSE](LICENSE).
