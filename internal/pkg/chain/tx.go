package chain

import (
	"sync"
)

type TxPublish struct {
	Action   GameAction
	HopLimit uint32
}

func (tx TxPublish) DecrementHopLimit() {
	tx.HopLimit--
}

func (tx TxPublish) GetHopLimit() uint32 {
	return tx.HopLimit
}

const txHopLimit = 10
const initialBalance = 100

var pendingTransactions = struct {
	sync.RWMutex
	m map[int][]TxPublish
}{
	m: map[int][]TxPublish{
		Spawn:   {},
		Attack:  {},
		Defence: {},
		Reveal:  {},
		Cancel:  {},
	},
}

var TxsChan = make(chan TxPublish)

func publishAction(action GameAction) {
	TxsChan <- TxPublish{
		Action: action,
		HopLimit: txHopLimit,
	}
}

func getNewSpawns(tmpBalances map[string]int64) []TxPublish {
	pendingTransactions.RLock()
	var validSpawns = []TxPublish(nil)
	defer pendingTransactions.RUnlock()
	for _, spawn := range pendingTransactions.m[Spawn] {
		tmpBalances[spawn.Action.Attacker] = initialBalance
		validSpawns = append(validSpawns, spawn)
	}
	return validSpawns
}

func getNewAttacks(tmpBalances map[string]int64) []TxPublish {
	var validAttacks = []TxPublish(nil)
	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()
	for _, attack := range pendingTransactions.m[Attack] {
		if isValidAttack(attack, tmpBalances) {
			tmpBalances[attack.Action.Attacker] -= int64(attack.Action.Bet)
			tmpBalances[attack.Action.Defender] -= int64(attack.Action.Bet)
			validAttacks = append(validAttacks, attack)
		}
	}
	return validAttacks
}

func getNewDefences(attacks []TxPublish) []TxPublish {
	var validDefences = []TxPublish(nil)
	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()
	for _, defence := range pendingTransactions.m[Defence] {
		if isValidDefence(defence, attacks) {
			validDefences = append(validDefences, defence)
		}
	}
	return validDefences
}

func getNewReveals(defences []TxPublish) []TxPublish {
	var validReveals = []TxPublish(nil)
	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()
	for _, reveal := range pendingTransactions.m[Reveal] {
		if isValidReveal(reveal, defences) {
			validReveals = append(validReveals, reveal)
		}
	}
	return validReveals
}

func getNewCancels(defences []TxPublish) []TxPublish {
	var validCancels = []TxPublish(nil)
	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()
	for _, cancel := range pendingTransactions.m[Cancel] {
		if isValidCancel(cancel, defences) {
			validCancels = append(validCancels, cancel)
		}
	}
	return validCancels
}

func hasNoPendingTransactions() bool {
	pendingTransactions.RLock()
	defer pendingTransactions.RUnlock()

	return len(pendingTransactions.m[Spawn]) < 1 &&
		len(pendingTransactions.m[Attack]) < 1 &&
		len(pendingTransactions.m[Defence]) < 1 &&
		len(pendingTransactions.m[Reveal]) < 1 &&
		len(pendingTransactions.m[Cancel]) < 1

}

func (action *GameAction) shouldDiscardTransactionUnsafe() bool {
	// TODO verify hashes
	switch action.Type {
	case Spawn:
		if isSpawnClaimed(action.Attacker) {
			return true
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingSpawn := range pendingTransactions.m[Spawn] {
			if action.Identifier == pendingSpawn.Action.Identifier {
				return true
			}
		}
	case Attack:
		if isAttackClaimed(action.Identifier) {
			return true
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingAttack := range pendingTransactions.m[Attack] {
			if action.Identifier == pendingAttack.Action.Identifier {
				return true
			}
		}
	case Defence:
		if isDefenceClaimed(action.Identifier) {
			return true
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingDefence := range pendingTransactions.m[Defence] {
			if action.Identifier == pendingDefence.Action.Identifier {
				return true
			}
		}
	case Reveal:
		if isRevealClaimed(action.Identifier) {
			return true
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingReveal := range pendingTransactions.m[Reveal] {
			if action.Identifier == pendingReveal.Action.Identifier {
				return true
			}
		}
	case Cancel:
		if isCancelClaimed(action.Identifier) {
			return true
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingCancel := range pendingTransactions.m[Cancel] {
			if action.Identifier == pendingCancel.Action.Identifier {
				return true
			}
		}
	}
	return false
}

//func ReceiveTransaction(tx TxPublish) {
//	if tx.File.shouldDiscardTransaction() {
//		return
//	}
//
//	pendingTransactions.Lock()
//	pendingTransactions.l = append(pendingTransactions.l, tx)
//	pendingTransactions.Unlock()
//}

func ReceiveTransaction(tx TxPublish) {
	blockchain.RLock()
	shouldDiscard := tx.Action.shouldDiscardTransactionUnsafe()
	blockchain.RUnlock()
	if shouldDiscard {
		return
	}
	pendingTransactions.Lock()
	switch tx.Action.Type {
	case Spawn:
		pendingTransactions.m[Spawn] = append(pendingTransactions.m[Spawn], tx)
	case Attack:
		pendingTransactions.m[Attack] = append(pendingTransactions.m[Attack], tx)
	case Defence:
		pendingTransactions.m[Defence] = append(pendingTransactions.m[Defence], tx)
	case Reveal:
		pendingTransactions.m[Reveal] = append(pendingTransactions.m[Reveal], tx)
	case Cancel:
		pendingTransactions.m[Cancel] = append(pendingTransactions.m[Cancel], tx)
	}
	pendingTransactions.Unlock()

}

func (tx *TxPublish) canAddToLedgerUnsafe(
	ledger ledger,
	tmpBalances map[string]int64,
	attacks map[uint64]struct{},
	defences map[uint64]struct{},
	reveals map[uint64]struct{},
	cancels map[uint64]struct{},
) bool {
	// TODO verify hashes
	switch tx.Action.Type {
	case Spawn:
		_, exists := tmpBalances[tx.Action.Attacker]
		if exists {
			return false
		}
		tmpBalances[tx.Action.Attacker] = initialBalance
	case Attack:
		if _, exists := attacks[tx.Action.Identifier]; exists {
			return false
		}
		if balance, exists := tmpBalances[tx.Action.Attacker]; !exists || balance < int64(tx.Action.Bet) {
			return false
		}
		if balance, exists := tmpBalances[tx.Action.Defender]; !exists || balance < int64(tx.Action.Bet) {
			return false
		}
		tmpBalances[tx.Action.Attacker] -= int64(tx.Action.Bet)
		tmpBalances[tx.Action.Defender] -= int64(tx.Action.Bet)
		attacks[tx.Action.Identifier] = struct{}{}
	case Defence:
		if _, exists := defences[tx.Action.Identifier]; exists {
			return false
		}
		_, attackExists := attacks[tx.Action.Identifier]
		if !attackExists {
			match, matchExists := ledger.challenges[tx.Action.Identifier]
			if !matchExists || match.Stage != Attack {
				return false
			}
		}
		return true
		defences[tx.Action.Identifier] = struct{}{}
	case Reveal:
		if _, exists := reveals[tx.Action.Identifier]; exists {
			return false
		}
		_, defenceExists := defences[tx.Action.Identifier]
		if !defenceExists {
			match, matchExists := ledger.challenges[tx.Action.Identifier]
			if !matchExists || match.Stage != Defence {
				return false
			}
		}
		return true
		reveals[tx.Action.Identifier] = struct{}{}
	case Cancel:
		if _, exists := cancels[tx.Action.Identifier]; exists {
			return false
		}
		if _, exists := defences[tx.Action.Identifier]; exists {
			return false
		}
		if _, exists := attacks[tx.Action.Identifier]; exists {
			return true
		}
		if match, exists := ledger.challenges[tx.Action.Identifier]; !exists || match.Stage != Attack {
			return false
		}
		cancels[tx.Action.Identifier] = struct{}{}
	}
	return true
}

//func BuildTransaction(file File) TxPublish {
//	tx := TxPublish{File: file, HopLimit: txHopLimit}
//	go ReceiveTransaction(tx)
//	return tx
//}

func BuildTransaction(action GameAction) TxPublish {
	tx := TxPublish{Action: action, HopLimit: txHopLimit}
	go ReceiveTransaction(tx)
	return tx
}

func removeClaimedPendingTransactionsUnsafe() {
	pendingTransactions.Lock()
	newSpawns := make([]TxPublish, 0)
	for _, spawnTx := range pendingTransactions.m[Spawn] {
		if !isSpawnClaimed(spawnTx.Action.Attacker) {
			newSpawns = append(newSpawns, spawnTx)
		}

	}
	newAttacks := make([]TxPublish, 0)
	for _, attackTx := range pendingTransactions.m[Attack] {
		if !isAttackClaimed(attackTx.Action.Identifier) {
			newAttacks = append(newAttacks, attackTx)
		}
	}
	newDefences := make([]TxPublish, 0)
	for _, defenceTx := range pendingTransactions.m[Defence] {
		if !isDefenceClaimed(defenceTx.Action.Identifier) {
			newDefences = append(newDefences, defenceTx)
		}
	}
	newReveals := make([]TxPublish, 0)
	for _, revealTx := range pendingTransactions.m[Reveal] {
		if !isRevealClaimed(revealTx.Action.Identifier) {
			newReveals = append(newReveals, revealTx)
		}
	}
	newCancels := make([]TxPublish, 0)
	for _, cancelTx := range pendingTransactions.m[Cancel] {
		if !isCancelClaimed(cancelTx.Action.Identifier) {
			newCancels = append(newCancels, cancelTx)
		}
	}

	pendingTransactions.m[Spawn] = newSpawns
	pendingTransactions.m[Attack] = newAttacks
	pendingTransactions.m[Defence] = newDefences
	pendingTransactions.m[Reveal] = newReveals
	pendingTransactions.m[Cancel] = newCancels
	pendingTransactions.Unlock()

}
