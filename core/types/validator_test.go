package types

import (
	"github.com/kowala-tech/kUSD/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"fmt"
)

func TestValidator_Properties(t *testing.T) {
	address := common.Address{}
	deposit := uint64(100)
	weight := &big.Int{}
	validator := NewValidator(address, deposit, weight)

	assert.Equal(t, address, validator.Address())
	assert.Equal(t, deposit, validator.Deposit())
	assert.Equal(t, weight, validator.Weight())
}

func TestValidatorSet_EmptyReturnsError(t *testing.T) {
	validatorSet, err := NewValidatorSet(nil)

	assert.Error(t, err)
	assert.Nil(t, validatorSet)
}

func TestValidatorSet_One(t *testing.T) {
	address := common.HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	deposit := uint64(100)
	weight := &big.Int{}
	validator := NewValidator(address, deposit, weight)

	validatorSet, err := NewValidatorSet([]*Validator{validator})

	assert.NoError(t, err)
	assert.Equal(t, 1, validatorSet.Size())
	assert.Equal(t, validator, validatorSet.AtIndex(0))
	assert.Equal(t, validator, validatorSet.Get(address))
	assert.Equal(t, true, validatorSet.Contains(address))
	assert.Equal(t, validator, validatorSet.Proposer())
}

func TestValidatorSet_UpdateWeightChangesProposer(t *testing.T) {
	validator := makeValidator("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", 100, 100)
	validator2 := makeValidator("0x6aaeb6053f3e94c9b9a09f33669435e7ef1beaed", 101, 101)
	validator3 := makeValidator("0x7aaeb6053f3e94c9b9a09f33669435e7ef1beaed", 99, 99)

	validatorSet, err := NewValidatorSet([]*Validator{validator, validator2, validator3})
	assert.NoError(t, err)

	validatorSet.UpdateWeight()
	assert.Equal(t, validator2, validatorSet.Proposer())
	assert.Equal(t, big.NewInt(101), validatorSet.Proposer().weight)
	assert.Equal(t, big.NewInt(200), validatorSet.AtIndex(0).weight)
	assert.Equal(t, big.NewInt(101), validatorSet.AtIndex(1).weight)
	assert.Equal(t, big.NewInt(198), validatorSet.AtIndex(2).weight)
	assert.Equal(t, 3, validatorSet.Size())
}

func TestValidatorSet_UpdateWeightChangesProposerElections(t *testing.T) {
	validator := makeValidator("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", 100, 100)
	validator2 := makeValidator("0x6aaeb6053f3e94c9b9a09f33669435e7ef1beaed", 101, 101)
	validator3 := makeValidator("0x7aaeb6053f3e94c9b9a09f33669435e7ef1beaed", 99, 99)

	validatorSet, err := NewValidatorSet([]*Validator{validator, validator2, validator3})
	assert.NoError(t, err)
	assert.Equal(t, 3, validatorSet.Size())

	elections := []struct {
		proposerWeight   *big.Int
		validator1weight *big.Int
		validator2weight *big.Int
		validator3weight *big.Int
	}{
		{big.NewInt(101), big.NewInt(200), big.NewInt(101), big.NewInt(198)},
		{big.NewInt(200), big.NewInt(200), big.NewInt(202), big.NewInt(297)},
		{big.NewInt(297), big.NewInt(300), big.NewInt(303), big.NewInt(297)},
		{big.NewInt(303), big.NewInt(400), big.NewInt(303), big.NewInt(396)},
	}

	for round, tc := range elections {
		t.Run(fmt.Sprintf("round %d", round), func(t *testing.T) {
			validatorSet.UpdateWeight()
			assert.Equal(t, tc.proposerWeight, validatorSet.Proposer().weight)
			assert.Equal(t, tc.validator1weight, validatorSet.AtIndex(0).weight)
			assert.Equal(t, tc.validator2weight, validatorSet.AtIndex(1).weight)
			assert.Equal(t, tc.validator3weight, validatorSet.AtIndex(2).weight)
		})
	}
}

func makeValidator(hexAddress string, deposit int, weight int64) *Validator {
	address := common.HexToAddress(hexAddress)
	return NewValidator(address, uint64(deposit), big.NewInt(weight))
}
