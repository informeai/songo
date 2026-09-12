package synth

// Envelope aplica um envelope ADSR (attack, decay, sustain, release) sobre
// as amostras, em segundos, evitando cliques e dando "corpo" ao som em vez
// de um tom cru. Modifica samples in-place.
func Envelope(samples []float64, attack, decay, sustainLevel, release float64) {
	n := len(samples)
	a := int(attack * SampleRate)
	d := int(decay * SampleRate)
	r := int(release * SampleRate)
	if a+d+r > n {
		total := a + d + r
		a = a * n / total
		d = d * n / total
		r = r * n / total
	}
	for i := range n {
		var g float64
		switch {
		case i < a:
			g = float64(i) / float64(maxInt(a, 1))
		case i < a+d:
			t := float64(i-a) / float64(maxInt(d, 1))
			g = 1 - t*(1-sustainLevel)
		case i < n-r:
			g = sustainLevel
		default:
			t := float64(i-(n-r)) / float64(maxInt(r, 1))
			g = sustainLevel * (1 - t)
		}
		samples[i] *= g
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
