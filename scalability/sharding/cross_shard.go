package sharding

import (
	"../../privacy/private_tx"
)

func (m *ShardManager) CrossShardTransfer(fromShard, toShard string, tx *private_tx.PrivateTransaction) error {
	from := m.Shards[fromShard]
	to := m.Shards[toShard]

	// 1. Lock
	if err := from.LockTransaction(tx); err != nil {
		return err
	}

	// 2. Commit
	if err := to.CommitTransaction(tx); err != nil {
		return err
	}

	// 3. Finalize
	from.FinalizeBlock()
	to.FinalizeBlock()

	return nil
}
