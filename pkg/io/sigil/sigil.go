// Package sigil provides the Sigil transformation framework for composable,
// reversible data transformations.
//
// Sigils are the core abstraction - each sigil implements a specific transformation
// (encoding, compression, hashing, encryption) with a uniform interface. Sigils can
// be chained together to create transformation pipelines.
//
// Example usage:
//
//	hexSigil, _ := sigil.NewSigil("hex")
//	base64Sigil, _ := sigil.NewSigil("base64")
//	result, _ := sigil.Transmute(data, []sigil.Sigil{hexSigil, base64Sigil})
package sigil

// Sigil defines the interface for a data transformer.
//
// A Sigil represents a single transformation unit that can be applied to byte data.
// Sigils may be reversible (encoding, compression, encryption) or irreversible (hashing).
//
// For reversible sigils: Out(In(x)) == x for all valid x
// For irreversible sigils: Out returns the input unchanged
// For symmetric sigils: In(x) == Out(x)
//
// Implementations must handle nil input by returning nil without error,
// and empty input by returning an empty slice without error.
type Sigil interface {
	// In applies the forward transformation to the data.
	// For encoding sigils, this encodes the data.
	// For compression sigils, this compresses the data.
	// For hash sigils, this computes the digest.
	In(data []byte) ([]byte, error)

	// Out applies the reverse transformation to the data.
	// For reversible sigils, this recovers the original data.
	// For irreversible sigils (e.g., hashing), this returns the input unchanged.
	Out(data []byte) ([]byte, error)
}

// Transmute applies a series of sigils to data in sequence.
//
// Each sigil's In method is called in order, with the output of one sigil
// becoming the input of the next. If any sigil returns an error, Transmute
// stops immediately and returns nil with that error.
//
// To reverse a transmutation, call each sigil's Out method in reverse order.
func Transmute(data []byte, sigils []Sigil) ([]byte, error) {
	var err error
	for _, s := range sigils {
		data, err = s.In(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Untransmute reverses a transmutation by applying Out in reverse order.
//
// Each sigil's Out method is called in reverse order, with the output of one sigil
// becoming the input of the next. If any sigil returns an error, Untransmute
// stops immediately and returns nil with that error.
func Untransmute(data []byte, sigils []Sigil) ([]byte, error) {
	var err error
	for i := len(sigils) - 1; i >= 0; i-- {
		data, err = sigils[i].Out(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
