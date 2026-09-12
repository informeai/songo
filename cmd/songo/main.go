// Command songo gera efeitos sonoros estilo 8-bit em arquivos .wav.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/informeai/songo/presets"
	"github.com/informeai/songo/script"
	"github.com/informeai/songo/synth"
)

func usage() {
	fmt.Fprintln(os.Stderr, `songo - gerador de efeitos sonoros 8-bit

Uso:
  songo list                             lista os presets disponíveis
  songo generate <preset> [-o arq] [-p]  gera um preset (padrão: <preset>.wav)
  songo all [-o dir] [-p]                gera todos os presets (padrão: ./sounds)
  songo run <script.sfx> [-o arq] [-p]   interpreta um script da DSL do songo

Flags:
  -o <caminho>   caminho de saída (arquivo em generate/run, diretório em all)
  -p, --play     toca cada som logo após gerá-lo

DSL (arquivos .sfx): uma instrução por linha. Geradores (square, triangle,
noise, sweep, fm) concatenam trechos de áudio; processadores (envelope,
lowpass, sweep_filter, delay, bitcrush) transformam o áudio já gerado.
Exemplo:

  square 988 0.05 duty=0.5 amp=0.4
  square 1319 0.15 duty=0.5 amp=0.4
  envelope 0.001 0.02 0.6 0.05

Veja mais exemplos em examples/*.sfx.`)
}

// parseFlags extrai "-o <valor>" e "-p"/"--play" de uma lista de argumentos,
// nessa ou em qualquer ordem.
func parseFlags(args []string, defaultOut string) (out string, play bool) {
	out = defaultOut
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o":
			if i+1 < len(args) {
				out = args[i+1]
				i++
			}
		case "-p", "--play":
			play = true
		}
	}
	return
}

func playOrWarn(path string) {
	if err := synth.Play(path); err != nil {
		fmt.Fprintln(os.Stderr, "aviso: não foi possível tocar o som:", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		names := make([]string, 0, len(presets.All))
		for name := range presets.All {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}

	case "generate":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "uso: songo generate <preset> [-o arquivo.wav] [-p]")
			os.Exit(1)
		}
		name := os.Args[2]
		preset, ok := presets.All[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "preset desconhecido: %s (use 'songo list')\n", name)
			os.Exit(1)
		}
		out, play := parseFlags(os.Args[3:], name+".wav")
		if err := synth.WriteWAV(out, preset()); err != nil {
			fmt.Fprintln(os.Stderr, "erro ao gerar som:", err)
			os.Exit(1)
		}
		fmt.Println("gerado:", out)
		if play {
			playOrWarn(out)
		}

	case "all":
		dir, play := parseFlags(os.Args[2:], "sounds")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "erro ao criar diretório:", err)
			os.Exit(1)
		}
		for _, name := range presets.Names() {
			out := filepath.Join(dir, name+".wav")
			if err := synth.WriteWAV(out, presets.All[name]()); err != nil {
				fmt.Fprintln(os.Stderr, "erro ao gerar", name, ":", err)
				os.Exit(1)
			}
			fmt.Println("gerado:", out)
			if play {
				playOrWarn(out)
			}
		}

	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "uso: songo run <script.sfx> [-o arquivo.wav] [-p]")
			os.Exit(1)
		}
		scriptPath := os.Args[2]
		defaultOut := strings.TrimSuffix(filepath.Base(scriptPath), filepath.Ext(scriptPath)) + ".wav"
		out, play := parseFlags(os.Args[3:], defaultOut)

		f, err := os.Open(scriptPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro ao abrir script:", err)
			os.Exit(1)
		}
		samples, err := script.Run(f)
		f.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro no script:", err)
			os.Exit(1)
		}
		if err := synth.WriteWAV(out, samples); err != nil {
			fmt.Fprintln(os.Stderr, "erro ao gerar som:", err)
			os.Exit(1)
		}
		fmt.Println("gerado:", out)
		if play {
			playOrWarn(out)
		}

	default:
		usage()
		os.Exit(1)
	}
}
