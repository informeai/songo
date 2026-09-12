package synth

// Reverb aplica uma reverberação simples estilo Schroeder: vários filtros
// comb em paralelo (dão a densidade de reflexões) seguidos de um filtro
// allpass (difunde no tempo sem colorir o timbre). mix em [0,1] controla
// a proporção entre o som seco e o reverberado.
func Reverb(samples []float64, mix float64) []float64 {
	if mix < 0 {
		mix = 0
	} else if mix > 1 {
		mix = 1
	}

	combDelaysMs := []float64{29.7, 37.1, 41.1, 43.7}
	var wet []float64
	for _, ms := range combDelaysMs {
		wet = addBuffers(wet, combFilter(samples, ms, 0.77))
	}
	for i := range wet {
		wet[i] /= float64(len(combDelaysMs))
	}
	wet = allpass(wet, 5.0, 0.5)

	return addBuffers(scaleBuffer(samples, 1-mix), scaleBuffer(wet, mix))
}

// combFilter é um filtro comb feedback: soma cada amostra a uma cópia
// atrasada e atenuada dela mesma, criando reflexões repetidas que decaem.
func combFilter(samples []float64, delayMs, gain float64) []float64 {
	delay := max(int(delayMs/1000*SampleRate), 1)
	const tailLoops = 20 // ciclos extras pra deixar a cauda decair
	n := len(samples) + delay*tailLoops
	buf := make([]float64, n)
	copy(buf, samples)
	for i := delay; i < n; i++ {
		buf[i] += buf[i-delay] * gain
	}
	return buf
}

// allpass é um filtro passa-tudo (Schroeder): difunde as reflexões no
// tempo sem alterar o espectro de frequência do sinal.
func allpass(samples []float64, delayMs, gain float64) []float64 {
	delay := max(int(delayMs/1000*SampleRate), 1)
	out := make([]float64, len(samples))
	buf := make([]float64, delay)
	for i, s := range samples {
		bufOut := buf[i%delay]
		v := -gain*s + bufOut
		out[i] = v
		buf[i%delay] = s + gain*v
	}
	return out
}

// addBuffers soma duas amostras de tamanhos possivelmente diferentes,
// preenchendo com zero o que faltar na mais curta.
func addBuffers(a, b []float64) []float64 {
	n := max(len(a), len(b))
	out := make([]float64, n)
	copy(out, a)
	for i, v := range b {
		out[i] += v
	}
	return out
}

// scaleBuffer multiplica todas as amostras por um ganho fixo.
func scaleBuffer(a []float64, gain float64) []float64 {
	out := make([]float64, len(a))
	for i, v := range a {
		out[i] = v * gain
	}
	return out
}
