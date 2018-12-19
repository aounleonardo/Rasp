package chain

import (
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
)

type ledger struct {
	players    map[string]*Player
	challenges map[uint64]*ChallengeState
	length     int
}

var genesis = [32]byte{}

var blockchain = struct {
	sync.RWMutex
	m       map[[32]byte]Block
	heads   map[[32]byte]ledger
	longest [32]byte
}{
	m: map[[32]byte]Block{genesis: {}},
	heads: map[[32]byte]ledger{genesis: {
		length:     0,
		challenges: make(map[uint64]*ChallengeState),
		players:    make(map[string]*Player)},
	},
	longest: genesis,
}

func getPlayer(name string) (Player, bool) {
	blockchain.RLock()
	defer blockchain.RUnlock()
	player, exists := blockchain.heads[blockchain.longest].players[name]
	return *player, exists
}

func getChallengeState(identifier uint64) (ChallengeState, bool) {
	blockchain.RLock()
	defer blockchain.RUnlock()
	state, exists := blockchain.heads[blockchain.longest].challenges[identifier]
	return *state, exists
}

func createForkLedgerUnsafe(
	ForkTxs map[int]map[uint64]GameAction,
	head [32]byte,
	length int,
) ledger {
	if head == genesis {
		return buildLedger(ForkTxs, length)
	}
	for _, tx := range blockchain.m[head].Transactions {
		ForkTxs[tx.Action.Type][tx.Action.Identifier] = tx.Action
	}

	return createForkLedgerUnsafe(ForkTxs, blockchain.m[head].PrevHash, length+1)
}

func createTxsMap(txs []TxPublish) map[int]map[uint64]GameAction {
	var txsMap = map[int]map[uint64]GameAction{
		Spawn:   make(map[uint64]GameAction),
		Attack:  make(map[uint64]GameAction),
		Defence: make(map[uint64]GameAction),
		Reveal:  make(map[uint64]GameAction),
		Cancel:  make(map[uint64]GameAction),
	}
	for _, tx := range txs {
		txsMap[tx.Action.Type][tx.Action.Identifier] = tx.Action
	}
	return txsMap
}

func applyTxsToLedger(txs map[int]map[uint64]GameAction, ledger *ledger) {
	for _, action := range txs[Spawn] {
		key, _ := x509.ParsePKCS1PublicKey(action.SignedSpecial)
		ledger.players[action.Attacker] = &Player{Balance: initialBalance, Key: key}
	}
	for identifier, action := range txs[Attack] {
		ledger.players[action.Attacker].Balance -= int64(action.Bet)
		ledger.challenges[identifier] = &ChallengeState{
			Identifier: identifier,
			Attacker:   action.Attacker,
			Defender:   &action.Defender,
			Bet:        action.Bet,
			Stage:      Attack,
			HiddenMove: action.HiddenMove,
		}
	}
	for identifier, action := range txs[Defence] {
		ledger.players[action.Defender].Balance += int64(action.Bet)
		match := ledger.challenges[identifier]
		match.Stage = Defence
		match.DefenceMove = &action.Move
	}
	for identifier, action := range txs[Reveal] {
		match := ledger.challenges[identifier]
		match.Stage = Reveal
		match.AttackMove = &action.Move
		match.Nonce = &action.Nonce
		switch whoWon(*match.AttackMove, *match.DefenceMove) {
		case Attacker:
			ledger.players[action.Attacker].Balance += 2 * int64(match.Bet)
			ledger.players[action.Defender].Balance -= 2 * int64(match.Bet)
		case Draw:
			ledger.players[action.Attacker].Balance += int64(match.Bet)
			ledger.players[action.Defender].Balance -= int64(match.Bet)
		}
	}
	for identifier, action := range txs[Cancel] {
		match := ledger.challenges[identifier]
		match.Stage = Cancel
		ledger.players[action.Attacker].Balance += int64(match.Bet)
	}
}

func buildLedger(ForkTxs map[int]map[uint64]GameAction, length int) ledger {
	var newLedger = ledger{
		players:    map[string]*Player{},
		challenges: map[uint64]*ChallengeState{},
		length:     length,
	}
	applyTxsToLedger(ForkTxs, &newLedger)
	return newLedger
}

// lock blockchain before using
func getBalancesUnsafe(ledger ledger) map[string]int64 {
	balances := make(map[string]int64, len(ledger.players))
	for name, player := range ledger.players {
		balances[name] = player.Balance
	}
	return balances
}

func spawnNotClaimed(newPlayer string) bool {
	_, ok := getPlayer(newPlayer)
	return !ok

}

func isValidCancel(cancel TxPublish, validDefences []TxPublish) bool {
	//check that defence not in pending
	//check that state is in Attack
	state, exist := getChallengeState(cancel.Action.Identifier)
	if !exist {
		return false
	}
	if hasDefenceInResults(cancel, validDefences) {
		return false
	}
	return state.Stage == Attack
}

func isValidReveal(reveal TxPublish, validDefences []TxPublish) bool {
	//Check that challenge state is set to defence in ledger
	//or that the defense is in results
	state, exist := getChallengeState(reveal.Action.Identifier)
	if !exist {
		return false
	}
	return state.Stage == Defence || hasDefenceInResults(reveal, validDefences)

}

func hasDefenceInResults(reveal TxPublish, defences []TxPublish) bool {
	id := reveal.Action.Identifier
	for _, defence := range defences {
		if defence.Action.Identifier == id {
			return true
		}
	}
	return false
}

func isValidDefence(defence TxPublish, validAttacks []TxPublish) bool {
	//check cancel in ledger
	//check attack in ledger or in validAttacks
	state, exists := getChallengeState(defence.Action.Identifier)
	if !exists {
		return false
	}
	//if state.Stage == Cancel{
	//	return false
	//}
	return state.Stage == Attack || hasAttackInResults(defence, validAttacks)
}

func hasAttackInResults(defence TxPublish, attacks []TxPublish) bool {
	id := defence.Action.Identifier
	for _, attack := range attacks {
		if attack.Action.Identifier == id {
			return true
		}
	}
	return false
}

func isValidAttack(attack TxPublish, balances map[string]int64) bool {
	if !hasSeenPlayer(attack.Action.Attacker, balances) ||
		!hasSeenPlayer(attack.Action.Defender, balances) {
		return false
	}
	upsertBalance(balances, attack.Action.Attacker)
	upsertBalance(balances, attack.Action.Defender)
	return haveEnoughMoney(attack.Action, balances)
}

func hasSeenPlayer(player string, balances map[string]int64) bool {
	if _, ok := balances[player]; ok {
		return true
	}
	if _, ok := getPlayer(player); ok {
		return true
	}
	return false
}

func upsertBalance(balances map[string]int64, name string) error {
	if _, ok := balances[name]; ok {
		return nil
	}
	player, ok := getPlayer(name)
	if !ok {
		return errors.New(fmt.Sprintf("unexpected error %s has no balance", name))

	}
	balances[name] = player.Balance
	return nil
}

func haveEnoughMoney(action GameAction, balances map[string]int64) bool {
	return balances[action.Attacker]-int64(action.Bet) >= 0 &&
		balances[action.Defender]-int64(action.Bet) >= 0
}

//func isNameClaimed(filename string) bool {
//	ledger.RLock()
//	defer ledger.RUnlock()
//	_, hasTx := ledger.m[filename]
//	return hasTx
//}

func isSpawnClaimed(name string) bool {
	_, exists := getPlayer(name)
	return exists
}

func isAttackClaimed(identifier uint64) bool {
	_, exist := getChallengeState(identifier)
	return exist
}

func isDefenceClaimed(identifier uint64) bool {
	challenge, exists := getChallengeState(identifier)
	if exists {
		return challenge.Stage > Attack
	}
	return false
}

func isRevealClaimed(identifier uint64) bool {
	challenge, exists := getChallengeState(identifier)
	if exists {
		if challenge.Stage == Reveal || challenge.Stage == Cancel {
			return true
		}
	}
	return false
}

func isCancelClaimed(identifier uint64) bool {
	challenge, exist := getChallengeState(identifier)
	if exist {
		return challenge.Stage == Attack
	}
	return false
}

func getHeadsCount() int {
	blockchain.RLock()
	defer blockchain.RUnlock()
	return len(blockchain.heads)
}

func isLongest(head [32]byte) bool {
	blockchain.RLock()
	defer blockchain.RUnlock()
	return head == blockchain.longest
}

func hasParent(block *Block) bool {
	blockchain.RLock()
	defer blockchain.RUnlock()
	_, exists := blockchain.m[block.Hash()]
	return exists
}

func hasBlock(block *Block) bool {
	blockchain.RLock()
	defer blockchain.RUnlock()
	_, hasBlock := blockchain.m[block.Hash()]
	return hasBlock
}

func getBlock(hash [32]byte) (*Block, error) {
	blockchain.RLock()
	defer blockchain.RUnlock()

	if block, hasBlock := blockchain.m[hash]; hasBlock {
		return &block, nil
	}
	return nil, errors.New(fmt.Sprintf("does not have block %s", hash))
}

func addBlockUnsafe(block Block) {
	currentHead, currentLength := getCurrentHeadUnsafe()
	hash := block.Hash()

	blockchain.m[hash] = block
	oldLedger := blockchain.heads[block.PrevHash]
	blockchain.heads[hash] = oldLedger
	blockchain.heads[hash] = ledger{
		players:    oldLedger.players,
		challenges: oldLedger.challenges,
		length:     oldLedger.length + 1,
	}

	applyBlockUnsafe(hash)
	delete(blockchain.heads, block.PrevHash)

	if oldLedger.length < currentLength {
		return
	}

	if block.PrevHash == currentHead {
		return
	}

	switchHeadFromUnsafe(currentHead)
}

func switchHeadFromUnsafe(previousHead [32]byte) {
	ancestor := getCommonAncestorUnsafe(previousHead)

	for hash := previousHead; hash != ancestor; {
		block := blockchain.m[hash]

		for _, tx := range block.Transactions {
			if !tx.Action.shouldDiscardTransactionUnsafe() {
				pendingTransactions.m[tx.Action.Type] =
					append(pendingTransactions.m[tx.Action.Type], tx)
			}
		}

		hash = block.PrevHash
	}
}

func getCommonAncestorUnsafe(other [32]byte) [32]byte {
	hashesToRoot := getChainHashesUnsafe()
	ancestor := findFirstInPathUnsafe(other, hashesToRoot)
	return ancestor
}

func findFirstInPathUnsafe(
	start [32]byte,
	hashesToRoot map[[32]byte]struct{},
) [32]byte {
	hash := start
	for {
		if hash == genesis {
			return genesis
		}
		block, _ := blockchain.m[hash]
		if _, inPath := hashesToRoot[hash]; inPath {
			return hash
		}
		hash = block.PrevHash
	}
}

func getCurrentHeadUnsafe() ([32]byte, int) {
	return blockchain.longest, blockchain.heads[blockchain.longest].length
}

func getChainHashesUnsafe() map[[32]byte]struct{} {
	length := blockchain.heads[blockchain.longest].length
	chain := make(map[[32]byte]struct{}, length)

	hash := blockchain.longest
	for {
		block, _ := blockchain.m[hash]
		if hash == genesis {
			return chain
		}
		chain[hash] = struct{}{}
		hash = block.PrevHash
	}
	return chain
}

func applyBlockUnsafe(hash [32]byte) {
	block := blockchain.m[hash]
	ledger := blockchain.heads[hash]
	txs := createTxsMap(block.Transactions)
	applyTxsToLedger(txs, &ledger)
}

func findNodeInPath(path [][32]byte, node [32]byte) (int, error) {
	for searchIndex := 0; searchIndex < len(path); searchIndex++ {
		if path[searchIndex] == node {
			return searchIndex, nil
		}
	}
	return -1, errors.New(fmt.Sprintf(
		"cannot find node %x in path %x",
		node,
		path,
	))
}
