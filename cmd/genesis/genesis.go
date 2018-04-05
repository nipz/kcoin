package main

import (
	"fmt"
	"github.com/kowala-tech/kcoin/kcoin/genesis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
)

var (
	cmd *cobra.Command
)

func init() {
	cmd = &cobra.Command{
		Use:   "genesis",
		Short: "Generator of a genesis file.",
		Long:  `Generate a genesis.json file based on a config file or parameters.`,
		Run: func(cmd *cobra.Command, args []string) {
			loadFromFileConfigIfAvailable()

			command := genesis.GenesisOptions{
				Network:                       viper.GetString("genesis.network"),
				MaxNumValidators:              viper.GetString("genesis.maxNumValidators"),
				UnbondingPeriod:               viper.GetString("genesis.unbondingPeriod"),
				WalletAddressGenesisValidator: viper.GetString("genesis.walletAddressGenesisValidator"),
				PrefundedAccounts:             parsePrefundedAccounts(viper.Get("prefundedAccounts")),
				ConsensusEngine:               viper.GetString("genesis.consensusEngine"),
				SmartContractsOwner:           viper.GetString("genesis.smartContractsOwner"),
				ExtraData:                     viper.GetString("genesis.extraData"),
			}

			fileName := "genesis.json"
			if viper.GetString("genesis.fileName") != "" {
				fileName = viper.GetString("genesis.fileName")
			}

			file, err := os.Create(fileName)
			if err != nil {
				fmt.Printf("Error generating file: %s", err)
				os.Exit(1)
			}

			handler := generateGenesisFileCommandHandler{w: file}
			err = handler.handle(command)
			if err != nil {
				fmt.Printf("Error generating file: %s", err)
				os.Exit(1)
			}

			fmt.Println("Genesis file generated.")
		},
	}

	cmd.Flags().StringP("config", "c", "", "Use to load configuration from config file.")
	cmd.Flags().StringP("network", "n", "", "The network to use, test or main")
	viper.BindPFlag("genesis.network", cmd.Flags().Lookup("network"))
	cmd.Flags().StringP("maxNumValidators", "v", "", "The maximum num of validators.")
	viper.BindPFlag("genesis.maxNumValidators", cmd.Flags().Lookup("maxNumValidators"))
	cmd.Flags().StringP("unbondingPeriod", "p", "", "The unbonding period in days.")
	viper.BindPFlag("genesis.unbondingPeriod", cmd.Flags().Lookup("unbondingPeriod"))
	cmd.Flags().StringP("walletAddressGenesisValidator", "g", "", "The wallet address of the genesis validator.")
	viper.BindPFlag("genesis.walletAddressGenesisValidator", cmd.Flags().Lookup("walletAddressGenesisValidator"))
	cmd.Flags().StringP("consensusEngine", "e", "", "The consensus engine, right now only supports tendermint")
	viper.BindPFlag("genesis.consensusEngine", cmd.Flags().Lookup("consensusEngine"))
	cmd.Flags().StringP("smartContractsOwner", "s", "", "The address of the smart contracts owner.")
	viper.BindPFlag("genesis.smartContractsOwner", cmd.Flags().Lookup("smartContractsOwner"))
	cmd.Flags().StringP("extraData", "d", "", "Extra data")
	viper.BindPFlag("genesis.extraData", cmd.Flags().Lookup("extraData"))
	cmd.Flags().StringP("prefundedAccounts", "a", "", "The prefunded accounts in format 0x212121:12,0x212121:14")
	viper.BindPFlag("prefundedAccounts", cmd.Flags().Lookup("prefundedAccounts"))
	cmd.Flags().StringP("fileName", "o", "", "The output filename (default:genesis.json).")
	viper.BindPFlag("genesis.fileName", cmd.Flags().Lookup("fileName"))
}

func loadFromFileConfigIfAvailable() {
	fileConfig, _ := cmd.Flags().GetString("config")
	if fileConfig != "" {
		viper.SetConfigFile(fileConfig)

		err := viper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parsePrefundedAccounts(accounts interface{}) []genesis.PrefundedAccount {
	prefundedAccounts := make([]genesis.PrefundedAccount, 0)

	switch accounts.(type) {
	case []interface{}:
		accountArray := accounts.([]interface{})
		for _, v := range accountArray {
			val := v.(map[string]interface{})

			prefundedAccount := genesis.PrefundedAccount{
				WalletAddress: val["walletAddress"].(string),
				Balance:       val["balance"].(int64),
			}

			prefundedAccounts = append(prefundedAccounts, prefundedAccount)
		}
	case string:
		accountsString := accounts.(string)
		if accountsString == "" {
			break
		}

		a := strings.Split(accountsString, ",")

		for _, v := range a {
			values := strings.Split(v, ":")
			balance, err := strconv.Atoi(values[1])
			if err != nil {
				balance = 0
			}

			prefundedAccount := genesis.PrefundedAccount{
				WalletAddress: values[0],
				Balance:       int64(balance),
			}

			prefundedAccounts = append(prefundedAccounts, prefundedAccount)
		}
	}

	return prefundedAccounts
}
