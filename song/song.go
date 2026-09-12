// Package song interpreta a DSL de composição do songo: arquivos .song
// com múltiplas vozes (baixo, lead, bateria...) descritas como sequências
// de notas musicais, ao contrário do package script que descreve um
// único efeito sonoro. Exemplo:
//
//	tempo 150
//
//	voice bass triangle
//	A2 e
//	E3 e
//
//	voice lead square duty=0.25
//	A4 e
//	C5 e
package song

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/informeai/songo/synth"
)

var noteIndex = map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}

type voice struct {
	gen     string
	params  map[string]string
	samples []float64
}

// Run interpreta um arquivo .song e retorna as amostras de todas as vozes
// já mixadas.
func Run(r io.Reader) ([]float64, error) {
	bpm := 120.0
	var voices []*voice
	var current *voice

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)

		switch fields[0] {
		case "tempo":
			v, err := parseTempo(fields)
			if err != nil {
				return nil, fmt.Errorf("linha %d: %w", lineNo, err)
			}
			bpm = v

		case "voice":
			v, err := parseVoice(fields)
			if err != nil {
				return nil, fmt.Errorf("linha %d: %w", lineNo, err)
			}
			current = v
			voices = append(voices, current)

		default:
			if current == nil {
				return nil, fmt.Errorf("linha %d: nota fora de uma voice (declare 'voice <nome> <instrumento>' antes)", lineNo)
			}
			if err := appendNote(current, fields, bpm); err != nil {
				return nil, fmt.Errorf("linha %d: %w", lineNo, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	all := make([][]float64, len(voices))
	for i, v := range voices {
		all[i] = v.samples
	}
	mixed := synth.Mix(all...)
	// várias vozes somadas facilmente estouram [-1, 1]; normaliza deixando
	// uma pequena folga (headroom) em vez de cortar o pico.
	return synth.Normalize(mixed, 0.9), nil
}

func parseTempo(fields []string) (float64, error) {
	if len(fields) != 2 {
		return 0, fmt.Errorf("uso: tempo <bpm>")
	}
	v, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, fmt.Errorf("bpm inválido: %q", fields[1])
	}
	return v, nil
}

func parseVoice(fields []string) (*voice, error) {
	if len(fields) < 3 {
		return nil, fmt.Errorf("uso: voice <nome> <instrumento> [param=valor ...]")
	}
	params := map[string]string{}
	for _, f := range fields[3:] {
		key, val, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("parâmetro inválido: %q", f)
		}
		params[key] = val
	}
	return &voice{gen: fields[2], params: params}, nil
}

func appendNote(v *voice, fields []string, bpm float64) error {
	if len(fields) < 2 {
		return fmt.Errorf("uso: <nota> <duração> [amp=valor]")
	}
	noteStr, lengthStr := fields[0], fields[1]

	amp := 1.0
	for _, f := range fields[2:] {
		key, val, ok := strings.Cut(f, "=")
		if !ok || key != "amp" {
			return fmt.Errorf("parâmetro inválido: %q", f)
		}
		a, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("amp inválido: %q", val)
		}
		amp = a
	}

	beats, err := noteLength(lengthStr)
	if err != nil {
		return err
	}
	duration := beats * (60 / bpm)

	if noteStr == "." || noteStr == "-" {
		v.samples = append(v.samples, make([]float64, int(duration*synth.SampleRate))...)
		return nil
	}

	var freq float64
	if v.gen != "noise" {
		freq, err = noteToFreq(noteStr)
		if err != nil {
			return err
		}
	}
	samples, err := renderNote(v.gen, v.params, freq, duration, amp)
	if err != nil {
		return err
	}
	v.samples = append(v.samples, samples...)
	return nil
}

// noteLength converte um código de duração (w=semibreve, h=mínima,
// q=semínima, e=colcheia, s=semicolcheia, com "." opcional pra ponteada)
// em número de tempos (beats), assumindo semínima = 1 tempo.
func noteLength(code string) (float64, error) {
	dotted := strings.HasSuffix(code, ".")
	base := strings.TrimSuffix(code, ".")
	var beats float64
	switch base {
	case "w":
		beats = 4
	case "h":
		beats = 2
	case "q":
		beats = 1
	case "e":
		beats = 0.5
	case "s":
		beats = 0.25
	default:
		return 0, fmt.Errorf("duração desconhecida: %q (use w, h, q, e, s, com \".\" opcional pra ponteada)", code)
	}
	if dotted {
		beats *= 1.5
	}
	return beats, nil
}

// noteToFreq converte notação científica (C4, F#3, Eb5...) em Hz, usando
// afinação padrão A4 = 440Hz.
func noteToFreq(note string) (float64, error) {
	if len(note) < 2 {
		return 0, fmt.Errorf("nota inválida: %q", note)
	}
	letter := strings.ToUpper(note[:1])[0]
	base, ok := noteIndex[letter]
	if !ok {
		return 0, fmt.Errorf("nota inválida: %q", note)
	}
	i := 1
	for i < len(note) && (note[i] == '#' || note[i] == 'b') {
		if note[i] == '#' {
			base++
		} else {
			base--
		}
		i++
	}
	if i >= len(note) {
		return 0, fmt.Errorf("nota sem oitava: %q", note)
	}
	octave, err := strconv.Atoi(note[i:])
	if err != nil {
		return 0, fmt.Errorf("oitava inválida em %q", note)
	}
	midi := (octave+1)*12 + base
	return 440 * math.Pow(2, float64(midi-69)/12), nil
}

func renderNote(gen string, params map[string]string, freq, duration, amp float64) ([]float64, error) {
	switch gen {
	case "square":
		duty, err := paramFloat(params, "duty", 0.5)
		if err != nil {
			return nil, err
		}
		samples := synth.Square(freq, duration, duty, amp)
		applyMelodicEnvelope(samples, duration)
		return samples, nil

	case "triangle":
		samples := synth.Triangle(freq, duration, amp)
		applyMelodicEnvelope(samples, duration)
		return samples, nil

	case "sine":
		samples := synth.Sine(freq, duration, amp)
		applyMelodicEnvelope(samples, duration)
		return samples, nil

	case "sawtooth":
		samples := synth.Sawtooth(freq, duration, amp)
		applyMelodicEnvelope(samples, duration)
		return samples, nil

	case "pluck":
		samples := synth.Pluck(freq, duration, amp)
		synth.Envelope(samples, 0.001, 0, 1.0, min(duration*0.1, 0.01))
		return samples, nil

	case "fm":
		modRatio, err := paramFloat(params, "mod_ratio", 2.0)
		if err != nil {
			return nil, err
		}
		modIndex, err := paramFloat(params, "mod_index", 4.0)
		if err != nil {
			return nil, err
		}
		samples := synth.FM(freq, freq*modRatio, modIndex, duration, amp)
		applyMelodicEnvelope(samples, duration)
		return samples, nil

	case "noise":
		samples := synth.Noise(duration, amp)
		applyPercussiveEnvelope(samples, duration)
		return samples, nil

	default:
		return nil, fmt.Errorf("instrumento desconhecido: %s", gen)
	}
}

// applyMelodicEnvelope evita cliques entre notas coladas uma na outra.
func applyMelodicEnvelope(samples []float64, duration float64) {
	release := min(duration*0.3, 0.03)
	synth.Envelope(samples, 0.002, 0.01, 0.85, release)
}

// applyPercussiveEnvelope dá o decaimento rápido típico de bateria 8-bit.
func applyPercussiveEnvelope(samples []float64, duration float64) {
	synth.Envelope(samples, 0.001, duration*0.7, 0.05, duration*0.2)
}

func paramFloat(params map[string]string, key string, def float64) (float64, error) {
	s, ok := params[key]
	if !ok {
		return def, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("valor inválido para %s: %q", key, s)
	}
	return v, nil
}
