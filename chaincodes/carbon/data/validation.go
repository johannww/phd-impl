package data

// ValidationMethod defines the method used for validation.
type ValidationMethod uint

const (
	ValidationMethodSattelite ValidationMethod = iota
	ValidationMethodGroundTruth
	ValidationMethodSelfDeclaration
)

func (vm ValidationMethod) IsValid() bool {
	switch vm {
	case ValidationMethodSattelite, ValidationMethodGroundTruth, ValidationMethodSelfDeclaration:
		return true
	default:
		return false
	}
}


type ValidationProps struct {
	Methods []ValidationMethod
}
