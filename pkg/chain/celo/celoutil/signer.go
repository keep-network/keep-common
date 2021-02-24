package celoutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/celo-org/celo-blockchain/crypto"
)

// SignatureSize is a byte size of a signature calculated by Celo with
// recovery-id, V, included. The signature consists of three values (R,S,V)
// in the following order:
// R = [0:31]
// S = [32:63]
// V = [64]
const SignatureSize = 65

// CeloSigner provides functions to sign a message and verify a signature
// using the Celo-specific signature format. It also provides functions for
// conversion of a public key to an address.
type CeloSigner struct {
	privateKey *ecdsa.PrivateKey
}

// NewSigner creates a new CeloSigner instance for the provided ECDSA
// private key.
func NewSigner(privateKey *ecdsa.PrivateKey) *CeloSigner {
	return &CeloSigner{privateKey}
}

// PublicKey returns byte representation of a public key for the private key
// signer was created with.
func (cs *CeloSigner) PublicKey() []byte {
	publicKey := cs.privateKey.PublicKey
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
}

// Sign signs the provided message using Celo-specific format.
func (cs *CeloSigner) Sign(message []byte) ([]byte, error) {
	signature, err := crypto.Sign(celoPrefixedHash(message), cs.privateKey)
	if err != nil {
		return nil, err
	}

	if len(signature) == SignatureSize {
		// celo-blockchain/crypto produces signature with v={0, 1} and we need
		// to add 27 to v-part (signature[64]) to conform with the on-chain
		// signature validation code that accepts v={27, 28} as specified in the
		// Appendix F of the Ethereum Yellow Paper
		// https://ethereum.github.io/yellowpaper/paper.pdf
		// As a fork of Ethereum, Celo respects this paper and didn't changed
		// anything relevant in their implementation.
		signature[len(signature)-1] = signature[len(signature)-1] + 27
	}

	return signature, nil
}

// Verify verifies the provided message against a signature using the key
// CeloSigner was created with. The signature has to be provided in
// Celo-specific format.
func (cs *CeloSigner) Verify(message []byte, signature []byte) (bool, error) {
	return verifySignature(message, signature, &cs.privateKey.PublicKey)
}

// VerifyWithPublicKey verifies the provided message against a signature and
// public key. The signature has to be provided in Celo-specific format.
func (cs *CeloSigner) VerifyWithPublicKey(
	message []byte,
	signature []byte,
	publicKey []byte,
) (bool, error) {
	unmarshalledPubKey, err := unmarshalPublicKey(
		publicKey,
		cs.privateKey.Curve,
	)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: [%v]", err)
	}

	return verifySignature(message, signature, unmarshalledPubKey)
}

func verifySignature(
	message []byte,
	signature []byte,
	publicKey *ecdsa.PublicKey,
) (bool, error) {
	// Convert the operator's static key into an uncompressed public key
	// which should be 65 bytes in length.
	uncompressedPubKey := crypto.FromECDSAPub(publicKey)
	// If our signature is in the [R || S || V] format, ensure we strip out
	// the Celo-specific recovery-id, V, if it already hasn't been done.
	if len(signature) == SignatureSize {
		signature = signature[:len(signature)-1]
	}

	// The signature should be now 64 bytes long.
	if len(signature) != 64 {
		return false, fmt.Errorf(
			"signature should have 64 bytes; has: [%d]",
			len(signature),
		)
	}

	return crypto.VerifySignature(
		uncompressedPubKey,
		celoPrefixedHash(message),
		signature,
	), nil
}

func celoPrefixedHash(message []byte) []byte {
	// Celo uses the same prefix as Ethereum.
	return crypto.Keccak256(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(message))),
		message,
	)
}

func unmarshalPublicKey(
	bytes []byte,
	curve elliptic.Curve,
) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(curve, bytes)
	if x == nil {
		return nil, fmt.Errorf(
			"invalid public key bytes",
		)
	}
	ecdsaPublicKey := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
	return (*ecdsa.PublicKey)(ecdsaPublicKey), nil
}

// PublicKeyToAddress transforms the provided ECDSA public key into Celo
// address represented in bytes.
func (cs *CeloSigner) PublicKeyToAddress(publicKey ecdsa.PublicKey) []byte {
	return crypto.PubkeyToAddress(publicKey).Bytes()
}

// PublicKeyBytesToAddress transforms the provided ECDSA public key in a bytes
// format into Celo address represented in bytes.
func (cs *CeloSigner) PublicKeyBytesToAddress(publicKey []byte) []byte {
	// Does the same as crypto.PubkeyToAddress but directly on public key bytes.
	return crypto.Keccak256(publicKey[1:])[12:]
}
