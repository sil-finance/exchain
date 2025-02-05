package simulation

import (
	"math/rand"

	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	govtypes "github.com/okex/exchain/libs/cosmos-sdk/x/gov/types"
	"github.com/okex/exchain/x/params/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/simulation"
)

// SimulateParamChangeProposalContent returns random parameter change content.
// It will generate a ParameterChangeProposal object with anywhere between 1 and
// the total amount of defined parameters changes, all of which have random valid values.
func SimulateParamChangeProposalContent(paramChangePool []simulation.ParamChange) simulation.ContentSimulatorFn {
	return func(r *rand.Rand, _ sdk.Context, _ []simulation.Account) govtypes.Content {

		lenParamChange := len(paramChangePool)
		if lenParamChange == 0 {
			panic("param changes array is empty")
		}

		numChanges := simulation.RandIntBetween(r, 1, lenParamChange)
		paramChanges := make([]types.ParamChange, numChanges)

		// map from key to empty struct; used only for look-up of the keys of the
		// parameters that are already in the random set of changes.
		paramChangesKeys := make(map[string]struct{})

		for i := 0; i < numChanges; i++ {
			spc := paramChangePool[r.Intn(len(paramChangePool))]

			// do not include duplicate parameter changes for a given subspace/key
			_, ok := paramChangesKeys[spc.ComposedKey()]
			for ok {
				spc = paramChangePool[r.Intn(len(paramChangePool))]
				_, ok = paramChangesKeys[spc.ComposedKey()]
			}

			// add a new distinct parameter to the set of changes and register the key
			// to avoid further duplicates
			paramChangesKeys[spc.ComposedKey()] = struct{}{}
			paramChanges[i] = types.NewParamChange(spc.Subspace, spc.Key, spc.SimValue(r))
		}

		return types.NewParameterChangeProposal(
			simulation.RandStringOfLength(r, 140),  // title
			simulation.RandStringOfLength(r, 5000), // description
			paramChanges,                           // set of changes
			100,
		)
	}
}
