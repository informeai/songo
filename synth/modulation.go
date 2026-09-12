package synth

import "math"

// Vibrato modula a altura do som lendo o buffer através de uma linha de
// atraso curta cujo tamanho oscila (LFO senoidal), gerando o efeito de
// pitch "balançando" sem precisar re-sintetizar a onda original.
// rateHz é a velocidade do balanço; depthMs é o quanto o pitch varia.
func Vibrato(samples []float64, rateHz, depthMs float64) []float64 {
	out := make([]float64, len(samples))
	depthSamples := depthMs / 1000 * SampleRate
	for i := range samples {
		t := float64(i) / SampleRate
		mod := depthSamples * (0.5 + 0.5*math.Sin(2*math.Pi*rateHz*t))
		out[i] = interpolate(samples, float64(i)-mod)
	}
	return out
}

// Tremolo modula o volume do som com um LFO senoidal. depth em [0,1]:
// 0 não tem efeito, 1 faz o volume oscilar entre zero e o volume original.
func Tremolo(samples []float64, rateHz, depth float64) []float64 {
	out := make([]float64, len(samples))
	for i, s := range samples {
		t := float64(i) / SampleRate
		lfo := 1 - depth*(0.5+0.5*math.Sin(2*math.Pi*rateHz*t))
		out[i] = s * lfo
	}
	return out
}

// interpolate lê o buffer numa posição fracionária usando interpolação
// linear entre as duas amostras vizinhas.
func interpolate(samples []float64, pos float64) float64 {
	if pos < 0 {
		return 0
	}
	i0 := int(math.Floor(pos))
	if i0 >= len(samples)-1 {
		if i0 >= len(samples) {
			return 0
		}
		return samples[i0]
	}
	frac := pos - float64(i0)
	return samples[i0]*(1-frac) + samples[i0+1]*frac
}
