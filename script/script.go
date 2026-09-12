// Package script interpreta a DSL declarativa do songo: um arquivo de
// texto com uma instrução por linha, onde geradores (square, triangle,
// noise, sweep, fm) concatenam trechos de áudio e processadores
// (envelope, lowpass, sweep_filter, delay, bitcrush) transformam o buffer
// acumulado até ali. Exemplo:
//
//	square 988 0.05 duty=0.5 amp=0.4
//	square 1319 0.15 duty=0.5 amp=0.4
//	envelope 0.001 0.02 0.6 0.05
package script

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/informeai/songo/synth"
)

// Run interpreta um script songo e retorna as amostras de áudio resultantes.
func Run(r io.Reader) ([]float64, error) {
	var buf []float64
	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var err error
		buf, err = execLine(buf, line)
		if err != nil {
			return nil, fmt.Errorf("linha %d: %w", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return buf, nil
}

func execLine(buf []float64, line string) ([]float64, error) {
	fields := strings.Fields(line)
	cmd := fields[0]
	pos, kv, err := splitArgs(fields[1:])
	if err != nil {
		return nil, err
	}

	switch cmd {
	case "square":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		duty, err := kvFloat(kv, "duty", 0.5)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.Square(p[0], p[1], duty, amp)...), nil

	case "triangle":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.Triangle(p[0], p[1], amp)...), nil

	case "noise":
		p, err := floats(pos, 1, cmd)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.Noise(p[0], amp)...), nil

	case "sweep":
		p, err := floats(pos, 3, cmd)
		if err != nil {
			return nil, err
		}
		duty, err := kvFloat(kv, "duty", 0.5)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.Sweep(p[0], p[1], p[2], duty, amp)...), nil

	case "fm":
		p, err := floats(pos, 4, cmd)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.FM(p[0], p[1], p[2], p[3], amp)...), nil

	case "sine":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.Sine(p[0], p[1], amp)...), nil

	case "sawtooth":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.Sawtooth(p[0], p[1], amp)...), nil

	case "pluck":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		amp, err := kvFloat(kv, "amp", 1.0)
		if err != nil {
			return nil, err
		}
		return append(buf, synth.Pluck(p[0], p[1], amp)...), nil

	case "envelope":
		p, err := floats(pos, 4, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		synth.Envelope(buf, p[0], p[1], p[2], p[3])
		return buf, nil

	case "lowpass":
		p, err := floats(pos, 1, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.LowPass(buf, p[0]), nil

	case "sweep_filter":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.SweepFilter(buf, p[0], p[1]), nil

	case "delay":
		p, err := floats(pos, 3, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Delay(buf, p[0], p[1], int(p[2])), nil

	case "bitcrush":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Bitcrush(buf, int(p[0]), int(p[1])), nil

	case "highpass":
		p, err := floats(pos, 1, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.HighPass(buf, p[0]), nil

	case "resonant":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Resonant(buf, p[0], p[1]), nil

	case "distortion":
		p, err := floats(pos, 1, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Distortion(buf, p[0]), nil

	case "reverb":
		p, err := floats(pos, 1, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Reverb(buf, p[0]), nil

	case "vibrato":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Vibrato(buf, p[0], p[1]), nil

	case "tremolo":
		p, err := floats(pos, 2, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Tremolo(buf, p[0], p[1]), nil

	case "gain":
		p, err := floats(pos, 1, cmd)
		if err != nil {
			return nil, err
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Gain(buf, p[0]), nil

	case "normalize":
		if len(pos) != 0 {
			return nil, fmt.Errorf("%s: não aceita argumentos posicionais (use level=0.9)", cmd)
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		level, err := kvFloat(kv, "level", 1.0)
		if err != nil {
			return nil, err
		}
		return synth.Normalize(buf, level), nil

	case "reverse":
		if len(pos) != 0 {
			return nil, fmt.Errorf("%s: não aceita argumentos", cmd)
		}
		if err := requireBuffer(buf, cmd); err != nil {
			return nil, err
		}
		return synth.Reverse(buf), nil

	default:
		return nil, fmt.Errorf("comando desconhecido: %s", cmd)
	}
}

func requireBuffer(buf []float64, cmd string) error {
	if len(buf) == 0 {
		return fmt.Errorf("%s: nenhum áudio gerado ainda (use um gerador antes, como square/triangle/noise/sweep/fm)", cmd)
	}
	return nil
}

// splitArgs separa os tokens de uma linha em argumentos posicionais
// (sem "=") e flags nomeadas ("chave=valor").
func splitArgs(fields []string) (pos []string, kv map[string]string, err error) {
	kv = map[string]string{}
	for _, f := range fields {
		key, val, ok := strings.Cut(f, "=")
		if !ok {
			pos = append(pos, f)
			continue
		}
		if key == "" {
			return nil, nil, fmt.Errorf("flag inválida: %q", f)
		}
		kv[key] = val
	}
	return pos, kv, nil
}

// floats valida a quantidade de argumentos posicionais e os converte
// para float64, na ordem em que aparecem.
func floats(pos []string, want int, cmd string) ([]float64, error) {
	if len(pos) != want {
		return nil, fmt.Errorf("%s: esperado %d argumento(s) posicional(is), recebido %d", cmd, want, len(pos))
	}
	out := make([]float64, want)
	for i, s := range pos {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: argumento %q inválido: %w", cmd, s, err)
		}
		out[i] = v
	}
	return out, nil
}

// kvFloat lê uma flag nomeada como float64, com valor padrão se ausente.
func kvFloat(kv map[string]string, key string, def float64) (float64, error) {
	s, ok := kv[key]
	if !ok {
		return def, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("valor inválido para %s: %q", key, s)
	}
	return v, nil
}
