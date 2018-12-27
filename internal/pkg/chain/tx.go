package chain

import (
	"fmt"
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
	tx := TxPublish{
		Action:   action,
		HopLimit: txHopLimit,
	}
	TxsChan <- tx
	go ReceiveTransaction(tx)
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
			if match, exists := getState(defence.Action.Identifier); exists {
				fmt.Println(match)
				if match.AttackMove != nil{
					if action, err :=
						createReveal(match, gossiperKey, defence.Action); err != nil {
						fmt.Println("error creating Reveal for defence",
							defence.Action.Identifier, err.Error())
					} else {
						publishAction(action)
					}
				}

			}
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
			if match, exists := getState(reveal.Action.Identifier); exists{
				raspState.Lock()
				match.AttackMove = &reveal.Action.Move
				match.Nonce = &reveal.Action.Nonce
				raspState.Unlock()
			}
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

	switch action.Type {
	case Spawn:
		if isSpawnClaimedUnsafe(action.Attacker) {
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
		if isAttackClaimedUnsafe(action.Identifier) {
			return true
		}
		player := blockchain.heads[blockchain.longest].players[action.Attacker]
		ok, err := VerifyAttack(
			&player.Key,
			action.Identifier,
			action.Bet,
			action.SignedSpecial,
		)
		if err != nil || !ok {
			return false
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingAttack := range pendingTransactions.m[Attack] {
			if action.Identifier == pendingAttack.Action.Identifier {
				return true
			}
		}
	case Defence:
		if isDefenceClaimedUnsafe(action.Identifier) {
			return true
		}
		player := blockchain.heads[blockchain.longest].players[action.Defender]
		ok, err := VerifyDefence(
			&player.Key,
			action.Identifier,
			action.Move,
			action.SignedSpecial,
		)
		if err != nil || !ok {
			return false
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingDefence := range pendingTransactions.m[Defence] {
			if action.Identifier == pendingDefence.Action.Identifier {
				return true
			}
		}
	case Reveal:
		if isRevealClaimedUnsafe(action.Identifier) {
			return true
		}
		player := blockchain.heads[blockchain.longest].players[action.Attacker]
		var hiddenMove Signature
		match, exists := blockchain.heads[blockchain.longest].matches[action.Identifier]
		if exists{
			hiddenMove = match.HiddenMove
		}else{
			pendingTransactions.RLock()
			for _, tx := range pendingTransactions.m[2]{
				if tx.Action.Identifier == action.Identifier{
					hiddenMove = tx.Action.HiddenMove
				}
			}
			pendingTransactions.RUnlock()
		}
		ok, err := VerifyReveal(
			&player.Key,
			action.Identifier,
			action.Move,
			action.Nonce,
			action.HiddenMove,
			action.SignedSpecial,
		)
		if err != nil || !ok {
			return false
		}
		ok, err = VerifyHiddenMove(
			&player.Key,
			action.Identifier,
			action.Move,
			action.Nonce,
			hiddenMove,
		)
		if err != nil || !ok {
			return false
		}
		pendingTransactions.RLock()
		defer pendingTransactions.RUnlock()
		for _, pendingReveal := range pendingTransactions.m[Reveal] {
			if action.Identifier == pendingReveal.Action.Identifier {
				return true
			}
		}
	case Cancel:
		if isCancelClaimedUnsafe(action.Identifier) {
			return true
		}
		player := blockchain.heads[blockchain.longest].players[action.Attacker]
		ok, err :=
			VerifyCancel(&player.Key, action.Identifier, action.SignedSpecial)
		if err != nil || !ok {
			return false
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
		player :=
			blockchain.heads[blockchain.longest].players[tx.Action.Attacker]
		ok, err := VerifyAttack(
			&player.Key,
			tx.Action.Identifier,
			tx.Action.Bet,
			tx.Action.SignedSpecial,
		)
		if err != nil || !ok {
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
			match, matchExists := ledger.matches[tx.Action.Identifier]
			if !matchExists || match.Stage != Attack {
				return false
			}
		}
		player :=
			blockchain.heads[blockchain.longest].players[tx.Action.Defender]
		ok, err := VerifyDefence(
			&player.Key,
			tx.Action.Identifier,
			tx.Action.Move,
			tx.Action.SignedSpecial,
		)
		if err != nil || !ok {
			return false
		}
		defences[tx.Action.Identifier] = struct{}{}
		return true
	case Reveal:
		if _, exists := reveals[tx.Action.Identifier]; exists {
			return false
		}
		_, defenceExists := defences[tx.Action.Identifier]
		if !defenceExists {
			match, matchExists := ledger.matches[tx.Action.Identifier]
			if !matchExists || match.Stage != Defence {
				return false
			}
		}
		player := blockchain.heads[blockchain.longest].players[tx.Action.Attacker]
		hiddenMove := blockchain.heads[blockchain.longest].
			matches[tx.Action.Identifier].HiddenMove
		ok, err := VerifyReveal(
			&player.Key,
			tx.Action.Identifier,
			tx.Action.Move,
			tx.Action.Nonce,
			tx.Action.HiddenMove,
			tx.Action.SignedSpecial,
		)
		if err != nil || !ok {
			return false
		}
		ok, err = VerifyHiddenMove(
			&player.Key,
			tx.Action.Identifier,
			tx.Action.Move,
			tx.Action.Nonce,
			hiddenMove,
		)
		if err != nil || !ok {
			return false
		}
		reveals[tx.Action.Identifier] = struct{}{}
		return true
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
		if match, exists := ledger.matches[tx.Action.Identifier]; !exists || match.Stage != Attack {
			return false
		}
		player :=
			blockchain.heads[blockchain.longest].players[tx.Action.Attacker]
		ok, err := VerifyCancel(
			&player.Key,
			tx.Action.Identifier,
			tx.Action.SignedSpecial,
		)
		if err != nil || !ok {
			return false
		}
		cancels[tx.Action.Identifier] = struct{}{}
	}
	return true
}

func removeClaimedPendingTransactionsUnsafe() {
	pendingTransactions.Lock()
	newSpawns := make([]TxPublish, 0)
	for _, spawnTx := range pendingTransactions.m[Spawn] {
		if !isSpawnClaimedUnsafe(spawnTx.Action.Attacker) {
			newSpawns = append(newSpawns, spawnTx)
		}

	}
	newAttacks := make([]TxPublish, 0)
	for _, attackTx := range pendingTransactions.m[Attack] {
		if !isAttackClaimedUnsafe(attackTx.Action.Identifier) {
			newAttacks = append(newAttacks, attackTx)
		}
	}
	newDefences := make([]TxPublish, 0)
	for _, defenceTx := range pendingTransactions.m[Defence] {
		if !isDefenceClaimedUnsafe(defenceTx.Action.Identifier) {
			newDefences = append(newDefences, defenceTx)
		}
	}
	newReveals := make([]TxPublish, 0)
	for _, revealTx := range pendingTransactions.m[Reveal] {
		if !isRevealClaimedUnsafe(revealTx.Action.Identifier) {
			newReveals = append(newReveals, revealTx)
		}
	}
	newCancels := make([]TxPublish, 0)
	for _, cancelTx := range pendingTransactions.m[Cancel] {
		if !isCancelClaimedUnsafe(cancelTx.Action.Identifier) {
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
