/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package statebased

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric/common/errors"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/version"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/stretchr/testify/assert"
)

type mockPolicyEvaluator struct {
	EvaluateRV error
}

func (m *mockPolicyEvaluator) Evaluate(policyBytes []byte, signatureSet []*common.SignedData) error {
	return m.EvaluateRV
}

func buildBlockWithTxs(txs ...[]byte) *common.Block {
	return &common.Block{
		Header: &common.BlockHeader{
			Number: 1,
		},
		Data: &common.BlockData{
			Data: txs,
		},
	}
}

func buildTXWithRwset(rws []byte) []byte {
	return utils.MarshalOrPanic(&common.Envelope{
		Payload: utils.MarshalOrPanic(
			&common.Payload{
				Data: utils.MarshalOrPanic(
					&pb.Transaction{
						Actions: []*pb.TransactionAction{
							{
								Payload: utils.MarshalOrPanic(&pb.ChaincodeActionPayload{
									Action: &pb.ChaincodeEndorsedAction{
										ProposalResponsePayload: utils.MarshalOrPanic(
											&pb.ProposalResponsePayload{
												Extension: utils.MarshalOrPanic(&pb.ChaincodeAction{Results: rws}),
											},
										),
									},
								}),
							},
						},
					},
				),
			},
		),
	})
}

func rwsetBytes(t *testing.T, cc string) []byte {
	rwsb := rwsetutil.NewRWSetBuilder()
	rwsb.AddToWriteSet(cc, "key", []byte("value"))
	rws := rwsb.GetTxReadWriteSet()
	rwsetbytes, err := rws.ToProtoBytes()
	assert.NoError(t, err)

	return rwsetbytes
}

func TestKeylevelValidation(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that writes
	// to a key that contains key-level validation params.
	// We simulate policy check success and failure

	vpMetadataKey := pb.MetaDataKeys_VALIDATION_PARAMETER.String()
	mr := &mockState{map[string][]byte{vpMetadataKey: []byte("EP")}, nil, map[string][]byte{vpMetadataKey: []byte("EP")}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	pe := &mockPolicyEvaluator{}
	validator := NewKeyLevelValidator(pe, pm)

	rwsb := rwsetBytes(t, "cc")
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	endorsements := []*pb.Endorsement{
		{
			Signature: []byte("signature"),
			Endorser:  []byte("endorser"),
		},
	}

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err := validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), endorsements)
	assert.NoError(t, err)

	pe.EvaluateRV = fmt.Errorf("policy evaluation error")

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), endorsements)
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}

func TestKeylevelValidationPvtData(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that writes
	// to a pvt key that contains key-level validation params.
	// We simulate policy check success and failure

	vpMetadataKey := pb.MetaDataKeys_VALIDATION_PARAMETER.String()
	mr := &mockState{map[string][]byte{vpMetadataKey: []byte("EP")}, nil, map[string][]byte{vpMetadataKey: []byte("EP")}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	pe := &mockPolicyEvaluator{}
	validator := NewKeyLevelValidator(pe, pm)

	rwsbu := rwsetutil.NewRWSetBuilder()
	rwsbu.AddToPvtAndHashedWriteSet("cc", "coll", "key", []byte("value"))
	rws := rwsbu.GetTxReadWriteSet()
	rwsb, err := rws.ToProtoBytes()
	assert.NoError(t, err)
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.NoError(t, err)

	pe.EvaluateRV = fmt.Errorf("policy evaluation error")

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}

func TestKeylevelValidationMetaUpdate(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that updates
	// the key-level validation parameters for a key.
	// We simulate policy check success and failure

	vpMetadataKey := pb.MetaDataKeys_VALIDATION_PARAMETER.String()
	mr := &mockState{map[string][]byte{vpMetadataKey: []byte("EP")}, nil, map[string][]byte{vpMetadataKey: []byte("EP")}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	pe := &mockPolicyEvaluator{}
	validator := NewKeyLevelValidator(pe, pm)

	rwsbu := rwsetutil.NewRWSetBuilder()
	rwsbu.AddToMetadataWriteSet("cc", "key", map[string][]byte{})
	rws := rwsbu.GetTxReadWriteSet()
	rwsb, err := rws.ToProtoBytes()
	assert.NoError(t, err)
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.NoError(t, err)

	pe.EvaluateRV = fmt.Errorf("policy evaluation error")

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}

func TestKeylevelValidationPvtMetaUpdate(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that updates
	// the key-level validation parameters for a pvt key.
	// We simulate policy check success and failure

	vpMetadataKey := pb.MetaDataKeys_VALIDATION_PARAMETER.String()
	mr := &mockState{map[string][]byte{vpMetadataKey: []byte("EP")}, nil, map[string][]byte{vpMetadataKey: []byte("EP")}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	pe := &mockPolicyEvaluator{}
	validator := NewKeyLevelValidator(pe, pm)

	rwsbu := rwsetutil.NewRWSetBuilder()
	rwsbu.AddToHashedMetadataWriteSet("cc", "coll", "key", map[string][]byte{})
	rws := rwsbu.GetTxReadWriteSet()
	rwsb, err := rws.ToProtoBytes()
	assert.NoError(t, err)
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.NoError(t, err)

	pe.EvaluateRV = fmt.Errorf("policy evaluation error")

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}

func TestKeylevelValidationPolicyRetrievalFailure(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that updates
	// the key-level validation parameters for a key.
	// we simulate the case where we fail to retrieve
	// the validation parameters from the ledger.

	mr := &mockState{nil, fmt.Errorf("metadata retrieval failure"), nil, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	validator := NewKeyLevelValidator(&mockPolicyEvaluator{}, pm)

	rwsb := rwsetBytes(t, "cc")
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err := validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCExecutionFailureError{}, err)
}

func TestCCEPValidation(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that doesn't
	// touch any key with a state-based endorsement policy;
	// we expect to check the normal cc-endorsement policy.

	mr := &mockState{map[string][]byte{}, nil, map[string][]byte{}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	pe := &mockPolicyEvaluator{}
	validator := NewKeyLevelValidator(pe, pm)

	rwsbu := rwsetutil.NewRWSetBuilder()
	rwsbu.AddToWriteSet("cc", "key", []byte("value"))
	rwsbu.AddToWriteSet("cc", "key1", []byte("value"))
	rwsbu.AddToReadSet("cc", "readkey", &version.Height{})
	rwsbu.AddToHashedReadSet("cc", "coll", "readpvtkey", &version.Height{})
	rws := rwsbu.GetTxReadWriteSet()
	rwsb, err := rws.ToProtoBytes()
	assert.NoError(t, err)
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.NoError(t, err)

	pe.EvaluateRV = fmt.Errorf("policy evaluation error")

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}

func TestCCEPValidationReads(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that doesn't
	// touch any key with a state-based endorsement policy;
	// we expect to check the normal cc-endorsement policy.

	mr := &mockState{map[string][]byte{}, nil, map[string][]byte{}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	pe := &mockPolicyEvaluator{}
	validator := NewKeyLevelValidator(pe, pm)

	rwsbu := rwsetutil.NewRWSetBuilder()
	rwsbu.AddToReadSet("cc", "readkey", &version.Height{})
	rws := rwsbu.GetTxReadWriteSet()
	rwsb, err := rws.ToProtoBytes()
	assert.NoError(t, err)
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.NoError(t, err)

	pe.EvaluateRV = fmt.Errorf("policy evaluation error")

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}

func TestCCEPValidationPvtReads(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that doesn't
	// touch any key with a state-based endorsement policy;
	// we expect to check the normal cc-endorsement policy.

	mr := &mockState{map[string][]byte{}, nil, map[string][]byte{}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	pe := &mockPolicyEvaluator{}
	validator := NewKeyLevelValidator(pe, pm)

	rwsbu := rwsetutil.NewRWSetBuilder()
	rwsbu.AddToHashedReadSet("cc", "coll", "readpvtkey", &version.Height{})
	rws := rwsbu.GetTxReadWriteSet()
	rwsb, err := rws.ToProtoBytes()
	assert.NoError(t, err)
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, fmt.Errorf(""))
	}()

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.NoError(t, err)

	pe.EvaluateRV = fmt.Errorf("policy evaluation error")

	err = validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}

func TestKeylevelValidationFailure(t *testing.T) {
	t.Parallel()

	// Scenario: we validate a transaction that writes
	// to a key that contains key-level validation params.
	// Validation fails because the block contains a previous
	// transaction that updates the key-level validation params
	// for that very same key.

	vpMetadataKey := pb.MetaDataKeys_VALIDATION_PARAMETER.String()
	mr := &mockState{map[string][]byte{vpMetadataKey: []byte("EP")}, nil, map[string][]byte{vpMetadataKey: []byte("EP")}, nil, false}
	ms := &mockStateFetcher{mr, nil}
	pm := &KeyLevelValidationParameterManagerImpl{Support: ms}
	validator := NewKeyLevelValidator(&mockPolicyEvaluator{}, pm)

	rwsb := rwsetBytes(t, "cc")
	prp := []byte("barf")
	block := buildBlockWithTxs(buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")), buildTXWithRwset(rwsetUpdatingMetadataFor("cc", "key")))

	validator.PreValidate(1, block)

	go func() {
		validator.PostValidate("cc", 1, 0, nil)
	}()

	err := validator.Validate("cc", 1, 1, rwsb, prp, []byte("CCEP"), []*pb.Endorsement{})
	assert.Error(t, err)
	assert.IsType(t, &errors.VSCCEndorsementPolicyError{}, err)
}
