package rpc

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmtypes "github.com/okex/exchain/libs/tendermint/types"

	"github.com/okex/exchain/libs/cosmos-sdk/client/context"
	"github.com/okex/exchain/libs/cosmos-sdk/client/flags"
	"github.com/okex/exchain/libs/cosmos-sdk/codec"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/types/rest"
)

// TODO these next two functions feel kinda hacky based on their placement

//ValidatorCommand returns the validator set for a given height
func ValidatorCommand(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tendermint-validator-set [height]",
		Short: "Get the full tendermint validator set at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var height *int64

			// optional height
			if len(args) > 0 {
				h, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				if h > 0 {
					tmp := int64(h)
					height = &tmp
				}
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			result, err := GetValidators(cliCtx, height, viper.GetInt(flags.FlagPage), viper.GetInt(flags.FlagLimit))
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(result)
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	cmd.Flags().Bool(flags.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(flags.FlagTrustNode, cmd.Flags().Lookup(flags.FlagTrustNode))
	cmd.Flags().Bool(flags.FlagIndentResponse, false, "indent JSON response")
	viper.BindPFlag(flags.FlagIndentResponse, cmd.Flags().Lookup(flags.FlagIndentResponse))
	cmd.Flags().Int(flags.FlagPage, 0, "Query a specific page of paginated results")
	viper.BindPFlag(flags.FlagPage, cmd.Flags().Lookup(flags.FlagPage))
	cmd.Flags().Int(flags.FlagLimit, 100, "Query number of results returned per page")

	return cmd
}

// Validator output in bech32 format
type ValidatorOutput struct {
	Address          sdk.ConsAddress `json:"address"`
	PubKey           string          `json:"pub_key"`
	ProposerPriority int64           `json:"proposer_priority"`
	VotingPower      int64           `json:"voting_power"`
}

// Validators at a certain height output in bech32 format
type ResultValidatorsOutput struct {
	BlockHeight int64             `json:"block_height"`
	Validators  []ValidatorOutput `json:"validators"`
}

func (rvo ResultValidatorsOutput) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("block height: %d\n", rvo.BlockHeight))

	for _, val := range rvo.Validators {
		b.WriteString(
			fmt.Sprintf(`
  Address:          %s
  Pubkey:           %s
  ProposerPriority: %d
  VotingPower:      %d
		`,
				val.Address, val.PubKey, val.ProposerPriority, val.VotingPower,
			),
		)
	}

	return b.String()
}

func bech32ValidatorOutput(validator *tmtypes.Validator) (ValidatorOutput, error) {
	bechValPubkey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, validator.PubKey)
	if err != nil {
		return ValidatorOutput{}, err
	}

	return ValidatorOutput{
		Address:          sdk.ConsAddress(validator.Address),
		PubKey:           bechValPubkey,
		ProposerPriority: validator.ProposerPriority,
		VotingPower:      validator.VotingPower,
	}, nil
}

// GetValidators from client
func GetValidators(cliCtx context.CLIContext, height *int64, page, limit int) (ResultValidatorsOutput, error) {
	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return ResultValidatorsOutput{}, err
	}

	validatorsRes, err := node.Validators(height, page, limit)
	if err != nil {
		return ResultValidatorsOutput{}, err
	}

	if !cliCtx.TrustNode {
		check, err := cliCtx.Verify(validatorsRes.BlockHeight)
		if err != nil {
			return ResultValidatorsOutput{}, err
		}

		if !bytes.Equal(check.ValidatorsHash, tmtypes.NewValidatorSet(validatorsRes.Validators).Hash()) {
			return ResultValidatorsOutput{}, fmt.Errorf("received invalid validatorset")
		}
	}

	outputValidatorsRes := ResultValidatorsOutput{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]ValidatorOutput, len(validatorsRes.Validators)),
	}

	for i := 0; i < len(validatorsRes.Validators); i++ {
		outputValidatorsRes.Validators[i], err = bech32ValidatorOutput(validatorsRes.Validators[i])
		if err != nil {
			return ResultValidatorsOutput{}, err
		}
	}

	return outputValidatorsRes, nil
}

// REST

// Validator Set at a height REST handler
func ValidatorSetRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 100)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse pagination parameters")
			return
		}

		vars := mux.Vars(r)
		height, err := strconv.ParseInt(vars["height"], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse block height")
			return
		}

		chainHeight, err := GetChainHeight(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, "failed to parse chain height")
			return
		}
		if height > chainHeight {
			rest.WriteErrorResponse(w, http.StatusNotFound, "requested block height is bigger then the chain length")
			return
		}

		output, err := GetValidators(cliCtx, &height, page, limit)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, output)
	}
}

// Latest Validator Set REST handler
func LatestValidatorSetRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 100)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse pagination parameters")
			return
		}

		output, err := GetValidators(cliCtx, nil, page, limit)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, output)
	}
}
