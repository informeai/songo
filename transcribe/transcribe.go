// Package transcribe detecta a melodia monofônica (uma nota de cada vez)
// de um sinal de áudio e converte pra um arquivo .song de uma única voz.
// Não separa instrumentos tocando simultaneamente (baixo+bateria+lead
// juntos) — é pensado pra transcrever uma linha melódica isolada (um
// assobio, um canto, um solo).
package transcribe

import (
	"fmt"
	"math"
	"strings"

	"github.com/informeai/songo/synth"
)

const (
	frameSize  = 2048 // ~46ms a 44100Hz: resolução de frequência suficiente pra graves
	hopSize    = 512  // ~11.6ms: resolução temporal da detecção
	silenceRMS = 0.02
	// tolerância de agrupamento: ~meio semitom (2^(1/24)) pra frames
	// consecutivos ainda contarem como a mesma nota sustentada
	pitchTolerance = 1.03
	minGroupFrames = 3 // grupos menores que isso são ruído/transiente, mescla no vizinho
)

// minFreq/maxFreq delimitam a faixa de frequência considerada válida pra
// detecção (abaixo/acima disso é tratado como ruído). São var, não const,
// porque viram limites de laço via conversão pra int em tempo de execução.
var (
	minFreq = 70.0 // ~D2
	maxFreq = 1200.0
)

var noteNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// NoteEvent é uma nota sustentada (freq > 0) ou uma pausa (freq == 0)
// com sua duração em segundos.
type NoteEvent struct {
	Freq     float64
	Duration float64
}

// FromSamples detecta a melodia monofônica presente nas amostras.
func FromSamples(samples []float64) []NoteEvent {
	type frame struct {
		freq float64 // 0 = silêncio
	}
	var frames []frame

	window := hannWindow(frameSize)
	for start := 0; start+frameSize <= len(samples); start += hopSize {
		raw := samples[start : start+frameSize]
		if rms(raw) < silenceRMS {
			frames = append(frames, frame{freq: 0})
			continue
		}
		windowed := make([]float64, frameSize)
		for i, s := range raw {
			windowed[i] = s * window[i]
		}
		frames = append(frames, frame{freq: detectPitch(windowed)})
	}

	// agrupa frames consecutivos com o mesmo pitch (dentro da tolerância)
	// ou consecutivos em silêncio numa única nota/pausa
	var groups []NoteEvent
	i := 0
	for i < len(frames) {
		j := i + 1
		for j < len(frames) && samePitch(frames[i].freq, frames[j].freq) {
			j++
		}
		groups = append(groups, NoteEvent{
			Freq:     frames[i].freq,
			Duration: float64(j-i) * hopSize / synth.SampleRate,
		})
		i = j
	}

	return mergeShortGroups(groups)
}

func samePitch(a, b float64) bool {
	if a == 0 || b == 0 {
		return a == b
	}
	ratio := a / b
	return ratio < pitchTolerance && ratio > 1/pitchTolerance
}

// mergeShortGroups funde grupos muito curtos (provavelmente ruído de
// transição entre notas) no grupo anterior, evitando fragmentar a melodia
// detectada em dezenas de micro-notas espúrias.
func mergeShortGroups(groups []NoteEvent) []NoteEvent {
	minDuration := float64(minGroupFrames) * hopSize / synth.SampleRate
	var out []NoteEvent
	for _, g := range groups {
		if g.Duration < minDuration && len(out) > 0 {
			out[len(out)-1].Duration += g.Duration
			continue
		}
		out = append(out, g)
	}
	return out
}

func rms(frame []float64) float64 {
	var sum float64
	for _, s := range frame {
		sum += s * s
	}
	return math.Sqrt(sum / float64(len(frame)))
}

func hannWindow(n int) []float64 {
	w := make([]float64, n)
	for i := range w {
		w[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
	}
	return w
}

// detectPitch estima a frequência fundamental via autocorrelação: soma o
// produto do sinal com uma cópia deslocada de si mesmo pra cada possível
// período, e escolhe o deslocamento (lag) com maior correlação dentro da
// faixa de frequência esperada.
func detectPitch(frame []float64) float64 {
	n := len(frame)
	minLag := int(float64(synth.SampleRate) / maxFreq)
	maxLag := min(int(float64(synth.SampleRate)/minFreq), n-1)

	bestLag, bestCorr := -1, 0.0
	for lag := minLag; lag <= maxLag; lag++ {
		var sum float64
		for i := 0; i < n-lag; i++ {
			sum += frame[i] * frame[i+lag]
		}
		if sum > bestCorr {
			bestCorr = sum
			bestLag = lag
		}
	}
	if bestLag <= 0 {
		return 0
	}
	return float64(synth.SampleRate) / float64(bestLag)
}

// freqToNote converte Hz pra notação científica (A4, C#5...), afinação
// padrão A4 = 440Hz.
func freqToNote(freq float64) string {
	midi := int(math.Round(69 + 12*math.Log2(freq/440)))
	name := noteNames[((midi%12)+12)%12]
	octave := midi/12 - 1
	return fmt.Sprintf("%s%d", name, octave)
}

// quantizeDuration escolhe o código de duração (w/h/q/e/s, com "." pra
// ponteada) mais próximo do tempo medido, dado o BPM assumido.
func quantizeDuration(seconds, bpm float64) string {
	beat := 60 / bpm
	beats := seconds / beat
	options := []struct {
		code  string
		beats float64
	}{
		{"s", 0.25}, {"s.", 0.375},
		{"e", 0.5}, {"e.", 0.75},
		{"q", 1}, {"q.", 1.5},
		{"h", 2}, {"h.", 3},
		{"w", 4},
	}
	best, bestDiff := options[0], math.Abs(beats-options[0].beats)
	for _, o := range options[1:] {
		if d := math.Abs(beats - o.beats); d < bestDiff {
			best, bestDiff = o, d
		}
	}
	return best.code
}

// emitNoteLines gera uma ou mais linhas ".song" pro token (nota ou "."
// de pausa) dado, quebrando durações muito longas em várias semibreves
// consecutivas, já que a DSL não tem um código de duração maior que "w".
func emitNoteLines(token string, seconds, bpm float64) []string {
	wholeSec := 4 * (60 / bpm)
	var lines []string
	for seconds > wholeSec*1.25 {
		lines = append(lines, token+" w")
		seconds -= wholeSec
	}
	lines = append(lines, token+" "+quantizeDuration(seconds, bpm))
	return lines
}

// ToSongText converte os eventos detectados num arquivo .song de uma
// única voz, pronto pra ser tocado com "songo song".
func ToSongText(events []NoteEvent, bpm float64, voiceName, instrument string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "tempo %g\n\n", bpm)
	fmt.Fprintf(&sb, "voice %s %s\n", voiceName, instrument)
	for _, ev := range events {
		token := "."
		if ev.Freq > 0 {
			token = freqToNote(ev.Freq)
		}
		for _, line := range emitNoteLines(token, ev.Duration, bpm) {
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
