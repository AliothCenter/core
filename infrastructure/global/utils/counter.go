package utils

var (
	// GreeceAlphabet is the alphabet of greece
	GreeceAlphabet = []rune{
		'α', 'β', 'γ', 'δ', 'ε', 'ζ', 'η', 'θ', 'ι', 'κ', 'λ', 'μ', 'ν', 'ξ', 'ο', 'π', 'ρ', 'σ',
		'τ', 'υ', 'φ', 'χ', 'ψ', 'ω', 'Α', 'Β', 'Γ', 'Δ', 'Ε', 'Ζ', 'Η', 'Θ', 'Ι', 'Κ', 'Λ', 'Μ', 'Ν', 'Ξ', 'Ο', 'Π',
		'Ρ', 'Σ', 'Τ', 'Υ', 'Φ', 'Χ', 'Ψ', 'Ω',
	}
	GreeceAlphabetString = []string{
		"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta", "iota",
		"kappa", "lambda", "mu", "nu", "xi", "omicron", "pi", "rho", "sigma", "tau", "upsilon", "phi", "chi", "psi",
		"omega", "Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta", "Iota", "Kappa", "Lambda",
		"Mu", "Nu", "Xi", "Omicron", "Pi", "Rho", "Sigma", "Tau", "Upsilon", "Phi", "Chi", "Psi", "Omega",
	}
)

func GenerateGreeceAlphabet(i int) (greeceAlphabet string) {
	return string(GreeceAlphabet[i%len(GreeceAlphabet)])
}

func GenerateGreeceAlphabetString(i int) (greeceAlphabet string) {
	return GreeceAlphabetString[i%len(GreeceAlphabetString)]
}
