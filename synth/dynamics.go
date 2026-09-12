package synth

import "math"

// Distortion aplica saturação (waveshaping via tanh). drive > 1 satura mais
// forte, dando aquele timbre "quebrado"/overdrive.
func Distortion(samples []float64, drive float64) []float64 {
	out := make([]float64, len(samples))
	for i, s := range samples {
		out[i] = math.Tanh(s * drive)
	}
	return out
}

// Gain multiplica a amplitude de todas as amostras por factor.
func Gain(samples []float64, factor float64) []float64 {
	out := make([]float64, len(samples))
	for i, s := range samples {
		out[i] = s * factor
	}
	return out
}

// Normalize escala as amostras para que o pico absoluto chegue a target
// (ex.: 1.0 usa toda a faixa dinâmica, 0.9 deixa uma folga).
func Normalize(samples []float64, target float64) []float64 {
	peak := 0.0
	for _, s := range samples {
		if a := math.Abs(s); a > peak {
			peak = a
		}
	}
	if peak == 0 {
		return samples
	}
	return Gain(samples, target/peak)
}

// Reverse inverte a ordem das amostras (som tocando de trás pra frente).
func Reverse(samples []float64) []float64 {
	n := len(samples)
	out := make([]float64, n)
	for i, s := range samples {
		out[n-1-i] = s
	}
	return out
}
