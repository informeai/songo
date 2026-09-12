// Command songo gera efeitos sonoros estilo 8-bit em arquivos .wav.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/informeai/songo/presets"
	"github.com/informeai/songo/script"
	"github.com/informeai/songo/song"
	"github.com/informeai/songo/synth"
	"github.com/informeai/songo/transcribe"
)

func usage() {
	fmt.Fprintln(os.Stderr, `songo - gerador de efeitos sonoros 8-bit

Uso:
  songo list                             lista os presets disponíveis
  songo generate <preset> [-o arq] [-p]  gera um preset (padrão: <preset>.wav)
  songo all [-o dir] [-p]                gera todos os presets (padrão: ./sounds)
  songo run <script.sfx> [-o arq] [-p]   interpreta um script da DSL do songo
  songo song <musica.song> [-o arq] [-p] interpreta uma composição com múltiplas vozes
  songo transcribe <entrada.wav> [flags] detecta a melodia monofônica e gera um .song

Flags:
  -o <caminho>   caminho de saída (arquivo em generate/run, diretório em all)
  -p, --play     toca cada som logo após gerá-lo

DSL (arquivos .sfx): uma instrução por linha.
Geradores: square, triangle, noise, sweep, fm, sine, sawtooth, pluck.
Processadores: envelope, lowpass, sweep_filter, delay, bitcrush,
highpass, resonant, distortion, reverb, vibrato, tremolo, gain,
normalize, reverse.
Exemplo:

  square 988 0.05 duty=0.5 amp=0.4
  square 1319 0.15 duty=0.5 amp=0.4
  envelope 0.001 0.02 0.6 0.05

Veja mais exemplos em examples/*.sfx.

DSL de composição (arquivos .song): múltiplas vozes, cada uma com um
instrumento (square, triangle, sine, sawtooth, pluck, fm, noise) e uma
sequência de notas em notação científica (A4, C#5, Eb3...) ou "." pra
pausa, com duração em w/h/q/e/s (semibreve/mínima/semínima/colcheia/
semicolcheia, "." opcional pra ponteada). Exemplo:

  tempo 150

  voice bass triangle
  A2 e
  E3 e

  voice lead square duty=0.25
  A4 e
  C5 e

Veja examples/*.song.

transcribe: detecta pitch por autocorrelação numa melodia monofônica (um
instrumento/voz de cada vez — não separa faixas com vários instrumentos
tocando juntos) e gera uma voice .song com as notas encontradas. Aceita
apenas WAV mono 16-bit 44100Hz (converta com ffmpeg -ac 1 -ar 44100
-sample_fmt s16). Flags: -o <saida.song>, -tempo <bpm, padrão 120>,
-instrument <square|triangle|sine|sawtooth|pluck|fm, padrão square>,
-voice <nome, padrão melody>.`)
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

// parseTranscribeFlags extrai -o, -tempo, -instrument e -voice, em
// qualquer ordem, com valores padrão.
func parseTranscribeFlags(args []string, defaultOut string) (out string, bpm float64, instrument, voiceName string) {
	out = defaultOut
	bpm = 120
	instrument = "square"
	voiceName = "melody"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o":
			if i+1 < len(args) {
				out = args[i+1]
				i++
			}
		case "-tempo":
			if i+1 < len(args) {
				if v, err := strconv.ParseFloat(args[i+1], 64); err == nil {
					bpm = v
				}
				i++
			}
		case "-instrument":
			if i+1 < len(args) {
				instrument = args[i+1]
				i++
			}
		case "-voice":
			if i+1 < len(args) {
				voiceName = args[i+1]
				i++
			}
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

	case "song":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "uso: songo song <musica.song> [-o arquivo.wav] [-p]")
			os.Exit(1)
		}
		songPath := os.Args[2]
		defaultOut := strings.TrimSuffix(filepath.Base(songPath), filepath.Ext(songPath)) + ".wav"
		out, play := parseFlags(os.Args[3:], defaultOut)

		f, err := os.Open(songPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro ao abrir música:", err)
			os.Exit(1)
		}
		samples, err := song.Run(f)
		f.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro na música:", err)
			os.Exit(1)
		}
		if err := synth.WriteWAV(out, samples); err != nil {
			fmt.Fprintln(os.Stderr, "erro ao gerar áudio:", err)
			os.Exit(1)
		}
		fmt.Println("gerado:", out)
		if play {
			playOrWarn(out)
		}

	case "transcribe":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "uso: songo transcribe <entrada.wav> [-o saida.song] [-tempo 120] [-instrument square] [-voice melody]")
			os.Exit(1)
		}
		inPath := os.Args[2]
		defaultOut := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath)) + ".song"
		out, bpm, instrument, voiceName := parseTranscribeFlags(os.Args[3:], defaultOut)

		samples, err := synth.ReadWAV(inPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro ao ler wav:", err)
			os.Exit(1)
		}
		events := transcribe.FromSamples(samples)
		text := transcribe.ToSongText(events, bpm, voiceName, instrument)
		if err := os.WriteFile(out, []byte(text), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "erro ao gravar .song:", err)
			os.Exit(1)
		}
		fmt.Printf("gerado: %s (%d notas/pausas detectadas)\n", out, len(events))

	default:
		usage()
		os.Exit(1)
	}
}
