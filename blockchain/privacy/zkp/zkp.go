package zkp

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/constraint"

	"blockchain/storage/txpool"
)

type ZKProof struct {
	Proof        []byte
	PublicInputs []byte
}

type PrivacyManager struct {
	provingKey       groth16.ProvingKey
	verifyingKey     groth16.VerifyingKey
	constraintSystem constraint.ConstraintSystem
}

// NewPrivacyManager creates a new PrivacyManager with setup keys
func NewPrivacyManager() (*PrivacyManager, error) {
	// Compile the circuit
	var circuit TransactionCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %w", err)
	}

	// Generate the proving and verifying keys
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Groth16: %w", err)
	}

	return &PrivacyManager{
		provingKey:       pk,
		verifyingKey:     vk,
		constraintSystem: ccs,
	}, nil
}

// GenerateProof creates a zero-knowledge proof for a transaction
func (p *PrivacyManager) GenerateProof(transaction *txpool.Transaction) (*ZKProof, error) {
	// Create assignment with private and public inputs
	assignment := TransactionCircuit{
		SenderPrivateKey: 12345, // In practice, this would be the actual private key
		SenderNonce:      100,   // In practice, this would be the sender's nonce/balance
		SenderAddress:    hashString(transaction.From),
		ReceiverAddress:  hashString(transaction.To),
		Amount:           int(transaction.Amount),
		Fee:              int(transaction.Fee),
		TransactionHash:  hashString(transaction.ID),
	}

	// Create a witness from the assignment
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %w", err)
	}

	// Generate the proof
	proof, err := groth16.Prove(p.constraintSystem, p.provingKey, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	// Create public witness
	publicWitness, err := witness.Public()
	if err != nil {
		return nil, fmt.Errorf("failed to create public witness: %w", err)
	}

	// Serialize the proof
	var proofBuf bytes.Buffer
	encoder := gob.NewEncoder(&proofBuf)
	err = encoder.Encode(proof)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize proof: %w", err)
	}

	// Serialize public inputs
	var pubInputsBuf bytes.Buffer
	encoder = gob.NewEncoder(&pubInputsBuf)
	err = encoder.Encode(publicWitness)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize public inputs: %w", err)
	}

	return &ZKProof{
		Proof:        proofBuf.Bytes(),
		PublicInputs: pubInputsBuf.Bytes(),
	}, nil
}

// VerifyProof verifies a zero-knowledge proof
func (p *PrivacyManager) VerifyProof(zkProof *ZKProof) (bool, error) {
	// Deserialize the proof
	buf := bytes.NewBuffer(zkProof.Proof)
	decoder := gob.NewDecoder(buf)
	var proof groth16.Proof
	err := decoder.Decode(&proof)
	if err != nil {
		return false, fmt.Errorf("failed to deserialize proof: %w", err)
	}

	// Deserialize public inputs
	buf = bytes.NewBuffer(zkProof.PublicInputs)
	decoder = gob.NewDecoder(buf)
	publicWitness, err := witness.New(ecc.BN254.ScalarField())
	if err != nil {
		return false, fmt.Errorf("failed to create public witness: %w", err)
	}
	err = decoder.Decode(publicWitness)
	if err != nil {
		return false, fmt.Errorf("failed to deserialize public inputs: %w", err)
	}

	// Verify the proof
	err = groth16.Verify(proof, p.verifyingKey, publicWitness)
	if err != nil {
		return false, nil // Invalid proof, but not an error
	}

	return true, nil
}

// hashString converts a string to a frontend.Variable for use in circuits
func hashString(s string) frontend.Variable {
	// Simple hash function for demonstration purposes
	// In practice, you would use a proper cryptographic hash function
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	return hash
}
