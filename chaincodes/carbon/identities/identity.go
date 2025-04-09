package identities

// TODO: add attributes
type Identity interface {
	String() string
}

type X509Identity struct {
	CertID string
}

func (x509identity *X509Identity) String() string {
	return x509identity.CertID
}

// TODO: read about idemix identities
