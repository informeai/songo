// Package synth contém geradores de forma de onda, envelope, filtros e
// efeitos usados para sintetizar sons estilo 8-bit em Go puro (sem cgo,
// sem libs externas).
package synth

import (
	"math"
	"math/rand"
)

// SampleRate é a taxa de amostragem usada por todo o pacote.
const SampleRate = 44100

// Square gera uma onda quadrada clássica de console 8-bit.
// duty controla o timbre (0.5 = quadrada pura, 0.25/0.125 = mais fina, tipo NES).
func Square(freq, duration, duty, amp float64) []float64 {
	n := int(duration * SampleRate)
	out := make([]float64, n)
	period := SampleRate / freq
	for i := range out {
		phase := math.Mod(float64(i), period) / period
		if phase < duty {
			out[i] = amp
		} else {
			out[i] = -amp
		}
	}
	return out
}

// Triangle gera uma onda triangular (grave, tipo canal de baixo do NES).
func Triangle(freq, duration, amp float64) []float64 {
	n := int(duration * SampleRate)
	out := make([]float64, n)
	for i := range out {
		phase := math.Mod(float64(i)*freq/SampleRate, 1)
		out[i] = amp * (4*math.Abs(phase-0.5) - 1)
	}
	return out
}

// Noise gera ruído branco (explosões, tiros, passos).
func Noise(duration, amp float64) []float64 {
	n := int(duration * SampleRate)
	out := make([]float64, n)
	for i := range out {
		out[i] = amp * (rand.Float64()*2 - 1)
	}
	return out
}

// Sweep gera uma onda quadrada cuja frequência varia linearmente de
// startFreq a endFreq ao longo da duração (jump, laser, etc).
func Sweep(startFreq, endFreq, duration, duty, amp float64) []float64 {
	n := int(duration * SampleRate)
	out := make([]float64, n)
	for i := range out {
		t := float64(i) / float64(n)
		freq := startFreq + t*(endFreq-startFreq)
		phase := math.Mod(float64(i)*freq/SampleRate, 1)
		if phase < duty {
			out[i] = amp
		} else {
			out[i] = -amp
		}
	}
	return out
}

// FM gera um tom por FM synthesis: a frequência do carrier é modulada por
// um segundo oscilador. modIndex alto = som mais metálico/inarmônico
// (sinos, gongos); modIndex baixo = timbre mais suave, tipo vibrato.
func FM(carrierFreq, modFreq, modIndex, duration, amp float64) []float64 {
	n := int(duration * SampleRate)
	out := make([]float64, n)
	for i := range out {
		t := float64(i) / SampleRate
		modulator := modIndex * math.Sin(2*math.Pi*modFreq*t)
		out[i] = amp * math.Sin(2*math.Pi*carrierFreq*t+modulator)
	}
	return out
}

// Concat junta vários trechos de amostras em sequência.
func Concat(parts ...[]float64) []float64 {
	var out []float64
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// Mix soma várias vozes na mesma posição no tempo (polifonia). As vozes
// mais curtas simplesmente param de contribuir quando acabam.
func Mix(voices ...[]float64) []float64 {
	maxLen := 0
	for _, v := range voices {
		if len(v) > maxLen {
			maxLen = len(v)
		}
	}
	out := make([]float64, maxLen)
	for _, v := range voices {
		for i, s := range v {
			out[i] += s
		}
	}
	return out
}
