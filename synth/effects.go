package synth

import "math"

// LowPass é um filtro passa-baixa de 1 polo (RC). cutoffHz baixo deixa o
// som mais abafado/grave; alto quase não tem efeito.
func LowPass(samples []float64, cutoffHz float64) []float64 {
	out := make([]float64, len(samples))
	rc := 1 / (2 * math.Pi * cutoffHz)
	dt := 1.0 / SampleRate
	alpha := dt / (rc + dt)
	prev := 0.0
	for i, s := range samples {
		prev = prev + alpha*(s-prev)
		out[i] = prev
	}
	return out
}

// SweepFilter aplica um passa-baixa cujo cutoff varia ao longo do tempo,
// dando o "wobble" clássico de sintetizador analógico/retro.
func SweepFilter(samples []float64, startHz, endHz float64) []float64 {
	out := make([]float64, len(samples))
	dt := 1.0 / SampleRate
	prev := 0.0
	n := len(samples)
	for i, s := range samples {
		t := float64(i) / float64(n)
		cutoff := startHz + t*(endHz-startHz)
		rc := 1 / (2 * math.Pi * cutoff)
		alpha := dt / (rc + dt)
		prev = prev + alpha*(s-prev)
		out[i] = prev
	}
	return out
}

// Delay adiciona ecos repetidos e decrescentes ao sinal original.
func Delay(samples []float64, delaySec, feedback float64, repeats int) []float64 {
	delaySamples := int(delaySec * SampleRate)
	total := len(samples) + delaySamples*repeats
	out := make([]float64, total)
	copy(out, samples)
	gain := 1.0
	for r := 1; r <= repeats; r++ {
		gain *= feedback
		offset := delaySamples * r
		for i, s := range samples {
			if offset+i >= total {
				break
			}
			out[offset+i] += s * gain
		}
	}
	return out
}

// Bitcrush reduz a resolução das amostras (menos bits e menos taxa de
// amostragem efetiva) pra dar o timbre "lo-fi"/quebrado.
func Bitcrush(samples []float64, bits int, downsample int) []float64 {
	out := make([]float64, len(samples))
	levels := math.Pow(2, float64(bits))
	held := 0.0
	for i, s := range samples {
		if i%downsample == 0 {
			held = math.Round(s*levels) / levels
		}
		out[i] = held
	}
	return out
}
