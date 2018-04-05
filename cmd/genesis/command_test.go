package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/kowala-tech/kcoin/core"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"math/big"
	"github.com/kowala-tech/kcoin/common"
	"github.com/kowala-tech/kcoin/params"
)

func TestItFailsWhenRunningHandlerWithInvalidCommandValues(t *testing.T) {
	baseValidCommand := GenerateGenesisCommand{
		network: "test",
		maxNumValidators: "1",
		unbondingPeriod: "1",
		walletAddressGenesisValidator: "0xe2ac86cbae1bbbb47d157516d334e70859a1bee4",
		prefundedAccounts: []PrefundedAccount{
			{
				walletAddress: "0xe2ac86cbae1bbbb47d157516d334e70859a1bee4",
				balance:       15,
			},
		},
	}

	tests := []struct {
		TestName                string
		InvalidCommandFromValid func(command GenerateGenesisCommand) GenerateGenesisCommand
		ExpectedError           error
	}{
		{
			TestName: "Invalid Network",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.network = "fakeNetwork"
				return command
			},
			ExpectedError: ErrInvalidNetwork,
		},
		{
			TestName: "Empty max number of validators",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.maxNumValidators = ""
				return command
			},
			ExpectedError: ErrEmptyMaxNumValidators,
		},
		{
			TestName: "Empty unbonding period of days",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.unbondingPeriod = ""
				return command
			},
			ExpectedError: ErrEmptyUnbondingPeriod,
		},
		{
			TestName: "Empty wallet address of genesis validator",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.walletAddressGenesisValidator = ""
				return command
			},
			ExpectedError: ErrEmptyWalletAddressValidator,
		},
		{
			TestName: "Invalid wallet address less than 20 bytes with Hex prefix",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.walletAddressGenesisValidator = "0xe2ac86cbae1bbbb47d157516d334e70859a1be"
				return command
			},
			ExpectedError: ErrInvalidWalletAddressValidator,
		},
		{
			TestName: "Invalid wallet address less than 20 bytes without Hex prefix",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.walletAddressGenesisValidator = "e2ac86cbae1bbbb47d157516d334e70859a1be"
				return command
			},
			ExpectedError: ErrInvalidWalletAddressValidator,
		},
		{
			TestName: "Empty prefunded accounts",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.prefundedAccounts = []PrefundedAccount{}
				return command
			},
			ExpectedError: ErrEmptyPrefundedAccounts,
		},
		{
			TestName: "Prefunded accounts does not include validator address",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.prefundedAccounts = []PrefundedAccount{
					{
						walletAddress: "0xaaaaaacbae1bbbb47d157516d334e70859a1bee4",
						balance:       15,
					},
				}
				return command
			},
			ExpectedError: ErrWalletAddressValidatorNotInPrefundedAccounts,
		},
		{
			TestName: "Prefunded accounts has invalid account.",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.prefundedAccounts = []PrefundedAccount{
					{
						walletAddress: "0xe2ac86cbae1bbbb47d157516d334e70859a1bee4",
						balance:       15,
					},
					{
						walletAddress: "0xe286cbae1bbbb47d157516d334e70859a1bee4",
						balance:       15,
					},
				}
				return command
			},
			ExpectedError: ErrInvalidAddressInPrefundedAccounts,
		},
		{
			TestName: "Invalid consensus engine.",
			InvalidCommandFromValid: func(command GenerateGenesisCommand) GenerateGenesisCommand {
				command.consensusEngine = "fakeConsensus"
				return command
			},
			ExpectedError: ErrInvalidConsensusEngine,
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			handler := GenerateGenesisCommandHandler{}
			err := handler.Handle(test.InvalidCommandFromValid(baseValidCommand))
			if err != test.ExpectedError {
				t.Fatalf(
					"Invalid options did not return error. Expected error: %s, received error: %s",
					test.ExpectedError.Error(),
					err.Error(),
				)
			}
		})
	}
}

func TestItWritesTheGeneratedFileToAWriter(t *testing.T) {
	cmd := GenerateGenesisCommand{
		network:                       "test",
		maxNumValidators:              "5",
		unbondingPeriod:               "5",
		walletAddressGenesisValidator: "0xe2ac86cbae1bbbb47d157516d334e70859a1bee4",
		prefundedAccounts: []PrefundedAccount{
			{
				walletAddress: "0xe2ac86cbae1bbbb47d157516d334e70859a1bee4",
				balance:       15,
			},
		},
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	handler := GenerateGenesisCommandHandler{w: writer}

	err := handler.Handle(cmd)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	fileName := "testfiles/testnet_default.json"
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Failed to read file %s", fileName)
	}

	var expectedGenesis = new(core.Genesis)
	err = json.Unmarshal(contents, expectedGenesis)
	if err != nil {
		t.Fatalf("Error unmarshalling json genesis with error: %s", err.Error())
	}

	var generatedGenesis = new(core.Genesis)
	err = json.Unmarshal(b.Bytes(), generatedGenesis)

	assertEqualGenesis(t, expectedGenesis, generatedGenesis)
}

//assertEqualGenesis checks if two genesis are the same, it ignores some fields as the timestamp that
//will be always different when it is generated.
func assertEqualGenesis(t *testing.T, expectedGenesis *core.Genesis, generatedGenesis *core.Genesis) {
	assert.Equal(t, expectedGenesis.ExtraData, generatedGenesis.ExtraData)
	assert.Equal(t, expectedGenesis.Config, generatedGenesis.Config)
	assert.Equal(t, expectedGenesis.GasLimit, generatedGenesis.GasLimit)
	assert.Equal(t, expectedGenesis.GasUsed, generatedGenesis.GasUsed)
	assert.Equal(t, expectedGenesis.Coinbase, generatedGenesis.Coinbase)
	assert.Equal(t, expectedGenesis.ParentHash, generatedGenesis.ParentHash)

	assert.Len(t, expectedGenesis.Alloc, len(generatedGenesis.Alloc))

	bigaddr, _ := new(big.Int).SetString(DefaultSmartContractsOwner, 0)
	address := common.BigToAddress(bigaddr)
	expectedAlloc := core.GenesisAccount{Balance: new(big.Int).Mul(common.Big1, big.NewInt(params.Ether))}
	assert.Equal(t, generatedGenesis.Alloc[address], expectedAlloc)
}

func TestOptionalValues(t *testing.T) {
	baseCommand := GenerateGenesisCommand{
		network:                       "test",
		maxNumValidators:              "5",
		unbondingPeriod:               "5",
		walletAddressGenesisValidator: "0xe2ac86cbae1bbbb47d157516d334e70859a1bee4",
		prefundedAccounts: []PrefundedAccount{
			{
				walletAddress: "0xe2ac86cbae1bbbb47d157516d334e70859a1bee4",
				balance:       15,
			},
		},
	}

	t.Run("Consensus engine value", func(t *testing.T) {
		baseCommand.consensusEngine = "tendermint"

		var b bytes.Buffer
		handler := GenerateGenesisCommandHandler{w: &b}

		err := handler.Handle(baseCommand)
		if err != nil {
			t.Fatalf("Error: %s", err.Error())
		}

		generatedGenesis := unmarshalGenesisFromBuffer(t, b)

		assert.NotNil(t, generatedGenesis.Config.Tendermint)
	})

	t.Run("Smart contracts owner", func(t *testing.T) {
		customSmartContractOwner := "0xe2ac86cbae1bbbb47d157516d334e70859a1aaaa"
		baseCommand.smartContractsOwner = customSmartContractOwner

		var b bytes.Buffer
		handler := GenerateGenesisCommandHandler{w: &b}

		err := handler.Handle(baseCommand)
		if err != nil {
			t.Fatalf("Error: %s", err.Error())
		}

		generatedGenesis := unmarshalGenesisFromBuffer(t, b)

		bigaddr, _ := new(big.Int).SetString(customSmartContractOwner, 0)
		address := common.BigToAddress(bigaddr)
		expectedAlloc := core.GenesisAccount{Balance: new(big.Int).Mul(common.Big1, big.NewInt(params.Ether))}

		assert.Equal(t, generatedGenesis.Alloc[address], expectedAlloc)
	})

	t.Run("Extra data", func(t *testing.T) {
		extraDataStr := "TheExtradata"
		baseCommand.extraData = extraDataStr

		var b bytes.Buffer
		handler := GenerateGenesisCommandHandler{w: &b}

		err := handler.Handle(baseCommand)
		if err != nil {
			t.Fatalf("Error: %s", err.Error())
		}

		generatedGenesis := unmarshalGenesisFromBuffer(t, b)
		expectedExtradata := make([]byte, 32)
		expectedExtradata = append([]byte(extraDataStr), expectedExtradata[len(extraDataStr):]...)

		assert.Equal(t, expectedExtradata, generatedGenesis.ExtraData)
	})
}

//unmarshalGenesisFromBuffer unmarshals a genesis struct from buffer in json.
func unmarshalGenesisFromBuffer(t *testing.T, b bytes.Buffer) *core.Genesis {
	var generatedGenesis = new(core.Genesis)

	err := json.Unmarshal(b.Bytes(), generatedGenesis)
	if err != nil {
		t.Fatal("Error unmarshaling genesis.")
	}

	return generatedGenesis
}