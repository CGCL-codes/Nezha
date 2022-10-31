// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package whisperv6

import (
	"math/big"
	mrand "math/rand"
	"testing"
	"time"

	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/crypto"
)

var seed int64

// InitSingleTest should be called in the beginning of every
// test, which uses RNG, in order to make the tests
// reproduciblity independent of their sequence.
func InitSingleTest() {
	seed = time.Now().Unix()
	mrand.Seed(seed)
}

type FilterTestCase struct {
	f      *Filter
	id     string
	alive  bool
	msgCnt int
}

func generateFilter(t *testing.T, symmetric bool) (*Filter, error) {
	var f Filter
	f.Messages = make(map[common.Hash]*ReceivedMessage)

	const topicNum = 8
	f.Topics = make([][]byte, topicNum)
	for i := 0; i < topicNum; i++ {
		f.Topics[i] = make([]byte, 4)
		mrand.Read(f.Topics[i])
		f.Topics[i][0] = 0x01
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generateFilter 1 failed with seed %d.", seed)
		return nil, err
	}
	f.Src = &key.PublicKey

	if symmetric {
		f.KeySym = make([]byte, aesKeyLength)
		mrand.Read(f.KeySym)
		f.SymKeyHash = crypto.Keccak256Hash(f.KeySym)
	} else {
		f.KeyAsym, err = crypto.GenerateKey()
		if err != nil {
			t.Fatalf("generateFilter 2 failed with seed %d.", seed)
			return nil, err
		}
	}

	// AcceptP2P & PoW are not set
	return &f, nil
}

func generateTestCases(t *testing.T, SizeTestFilters int) []FilterTestCase {
	cases := make([]FilterTestCase, SizeTestFilters)
	for i := 0; i < SizeTestFilters; i++ {
		f, _ := generateFilter(t, true)
		cases[i].f = f
		cases[i].alive = mrand.Int()&int(1) == 0
	}
	return cases
}

func TestInstallFilters(t *testing.T) {
	InitSingleTest()

	const SizeTestFilters = 256
	w := New(&Config{})
	filters := NewFilters(w)
	tst := generateTestCases(t, SizeTestFilters)

	var err error
	var j string
	for i := 0; i < SizeTestFilters; i++ {
		j, err = filters.Install(tst[i].f)
		if err != nil {
			t.Fatalf("seed %d: failed to install filter: %s", seed, err)
		}
		tst[i].id = j
		if len(j) != keyIDSize*2 {
			t.Fatalf("seed %d: wrong filter id size [%d]", seed, len(j))
		}
	}

	for _, testCase := range tst {
		if !testCase.alive {
			filters.Uninstall(testCase.id)
		}
	}

	for i, testCase := range tst {
		fil := filters.Get(testCase.id)
		exist := fil != nil
		if exist != testCase.alive {
			t.Fatalf("seed %d: failed alive: %d, %v, %v", seed, i, exist, testCase.alive)
		}
		if exist && fil.PoW != testCase.f.PoW {
			t.Fatalf("seed %d: failed Get: %d, %v, %v", seed, i, exist, testCase.alive)
		}
	}
}

func TestInstallSymKeyGeneratesHash(t *testing.T) {
	InitSingleTest()

	w := New(&Config{})
	filters := NewFilters(w)
	filter, _ := generateFilter(t, true)

	// save the current SymKeyHash for comparison
	initialSymKeyHash := filter.SymKeyHash

	// ensure the SymKeyHash is invalid, for Install to recreate it
	var invalid common.Hash
	filter.SymKeyHash = invalid

	_, err := filters.Install(filter)

	if err != nil {
		t.Fatalf("Error installing the filter: %s", err)
	}

	for i, b := range filter.SymKeyHash {
		if b != initialSymKeyHash[i] {
			t.Fatalf("The filter's symmetric key hash was not properly generated by Install")
		}
	}
}

func TestInstallIdenticalFilters(t *testing.T) {
	InitSingleTest()

	w := New(&Config{})
	filters := NewFilters(w)
	filter1, _ := generateFilter(t, true)

	// Copy the first filter since some of its fields
	// are randomly gnerated.
	filter2 := &Filter{
		KeySym:   filter1.KeySym,
		Topics:   filter1.Topics,
		PoW:      filter1.PoW,
		AllowP2P: filter1.AllowP2P,
		Messages: make(map[common.Hash]*ReceivedMessage),
	}

	_, err := filters.Install(filter1)

	if err != nil {
		t.Fatalf("Error installing the first filter with seed %d: %s", seed, err)
	}

	_, err = filters.Install(filter2)

	if err != nil {
		t.Fatalf("Error installing the second filter with seed %d: %s", seed, err)
	}

	params, err := generateMessageParams()
	if err != nil {
		t.Fatalf("Error generating message parameters with seed %d: %s", seed, err)
	}

	params.KeySym = filter1.KeySym
	params.Topic = BytesToTopic(filter1.Topics[0])

	filter1.Src = &params.Src.PublicKey
	filter2.Src = &params.Src.PublicKey

	sentMessage, err := NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err := sentMessage.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}
	msg := env.Open(filter1)
	if msg == nil {
		t.Fatalf("failed to Open with filter1")
	}

	if !filter1.MatchEnvelope(env) {
		t.Fatalf("failed matching with the first filter")
	}

	if !filter2.MatchEnvelope(env) {
		t.Fatalf("failed matching with the first filter")
	}

	if !filter1.MatchMessage(msg) {
		t.Fatalf("failed matching with the second filter")
	}

	if !filter2.MatchMessage(msg) {
		t.Fatalf("failed matching with the second filter")
	}
}

func TestInstallFilterWithSymAndAsymKeys(t *testing.T) {
	InitSingleTest()

	w := New(&Config{})
	filters := NewFilters(w)
	filter1, _ := generateFilter(t, true)

	asymKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Unable to create asymetric keys: %v", err)
	}

	// Copy the first filter since some of its fields
	// are randomly gnerated.
	filter := &Filter{
		KeySym:   filter1.KeySym,
		KeyAsym:  asymKey,
		Topics:   filter1.Topics,
		PoW:      filter1.PoW,
		AllowP2P: filter1.AllowP2P,
		Messages: make(map[common.Hash]*ReceivedMessage),
	}

	_, err = filters.Install(filter)

	if err == nil {
		t.Fatalf("Error detecting that a filter had both an asymmetric and symmetric key, with seed %d", seed)
	}
}

func TestComparePubKey(t *testing.T) {
	InitSingleTest()

	key1, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate first key with seed %d: %s.", seed, err)
	}
	key2, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate second key with seed %d: %s.", seed, err)
	}
	if IsPubKeyEqual(&key1.PublicKey, &key2.PublicKey) {
		t.Fatalf("public keys are equal, seed %d.", seed)
	}

	// generate key3 == key1
	mrand.Seed(seed)
	key3, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate third key with seed %d: %s.", seed, err)
	}
	if IsPubKeyEqual(&key1.PublicKey, &key3.PublicKey) {
		t.Fatalf("key1 == key3, seed %d.", seed)
	}
}

func TestMatchEnvelope(t *testing.T) {
	InitSingleTest()

	fsym, err := generateFilter(t, true)
	if err != nil {
		t.Fatalf("failed generateFilter with seed %d: %s.", seed, err)
	}

	fasym, err := generateFilter(t, false)
	if err != nil {
		t.Fatalf("failed generateFilter() with seed %d: %s.", seed, err)
	}

	params, err := generateMessageParams()
	if err != nil {
		t.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}

	params.Topic[0] = 0xFF // topic mismatch

	msg, err := NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	if _, err = msg.Wrap(params); err != nil {
		t.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}

	// encrypt symmetrically
	i := mrand.Int() % 4
	fsym.Topics[i] = params.Topic[:]
	fasym.Topics[i] = params.Topic[:]
	msg, err = NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err := msg.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap() with seed %d: %s.", seed, err)
	}

	// symmetric + matching topic: match
	match := fsym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope() symmetric with seed %d.", seed)
	}

	// symmetric + matching topic + insufficient PoW: mismatch
	fsym.PoW = env.PoW() + 1.0
	match = fsym.MatchEnvelope(env)
	if match {
		t.Fatalf("failed MatchEnvelope(symmetric + matching topic + insufficient PoW) asymmetric with seed %d.", seed)
	}

	// symmetric + matching topic + sufficient PoW: match
	fsym.PoW = env.PoW() / 2
	match = fsym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope(symmetric + matching topic + sufficient PoW) with seed %d.", seed)
	}

	// symmetric + topics are nil (wildcard): match
	prevTopics := fsym.Topics
	fsym.Topics = nil
	match = fsym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope(symmetric + topics are nil) with seed %d.", seed)
	}
	fsym.Topics = prevTopics

	// encrypt asymmetrically
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed GenerateKey with seed %d: %s.", seed, err)
	}
	params.KeySym = nil
	params.Dst = &key.PublicKey
	msg, err = NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err = msg.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap() with seed %d: %s.", seed, err)
	}

	// encryption method mismatch
	match = fsym.MatchEnvelope(env)
	if match {
		t.Fatalf("failed MatchEnvelope(encryption method mismatch) with seed %d.", seed)
	}

	// asymmetric + mismatching topic: mismatch
	match = fasym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope(asymmetric + mismatching topic) with seed %d.", seed)
	}

	// asymmetric + matching topic: match
	fasym.Topics[i] = fasym.Topics[i+1]
	match = fasym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope(asymmetric + matching topic) with seed %d.", seed)
	}

	// asymmetric + filter without topic (wildcard): match
	fasym.Topics = nil
	match = fasym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope(asymmetric + filter without topic) with seed %d.", seed)
	}

	// asymmetric + insufficient PoW: mismatch
	fasym.PoW = env.PoW() + 1.0
	match = fasym.MatchEnvelope(env)
	if match {
		t.Fatalf("failed MatchEnvelope(asymmetric + insufficient PoW) with seed %d.", seed)
	}

	// asymmetric + sufficient PoW: match
	fasym.PoW = env.PoW() / 2
	match = fasym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope(asymmetric + sufficient PoW) with seed %d.", seed)
	}

	// filter without topic + envelope without topic: match
	env.Topic = TopicType{}
	match = fasym.MatchEnvelope(env)
	if !match {
		t.Fatalf("failed MatchEnvelope(filter without topic + envelope without topic) with seed %d.", seed)
	}

	// filter with topic + envelope without topic: mismatch
	fasym.Topics = fsym.Topics
	match = fasym.MatchEnvelope(env)
	if !match {
		// topic mismatch should have no affect, as topics are handled by topic matchers
		t.Fatalf("failed MatchEnvelope(filter without topic + envelope without topic) with seed %d.", seed)
	}
}

func TestMatchMessageSym(t *testing.T) {
	InitSingleTest()

	params, err := generateMessageParams()
	if err != nil {
		t.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}

	f, err := generateFilter(t, true)
	if err != nil {
		t.Fatalf("failed generateFilter with seed %d: %s.", seed, err)
	}

	const index = 1
	params.KeySym = f.KeySym
	params.Topic = BytesToTopic(f.Topics[index])

	sentMessage, err := NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err := sentMessage.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}
	msg := env.Open(f)
	if msg == nil {
		t.Fatalf("failed Open with seed %d.", seed)
	}

	// Src: match
	*f.Src.X = *params.Src.PublicKey.X
	*f.Src.Y = *params.Src.PublicKey.Y
	if !f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(src match) with seed %d.", seed)
	}

	// insufficient PoW: mismatch
	f.PoW = msg.PoW + 1.0
	if f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(insufficient PoW) with seed %d.", seed)
	}

	// sufficient PoW: match
	f.PoW = msg.PoW / 2
	if !f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(sufficient PoW) with seed %d.", seed)
	}

	// topic mismatch
	f.Topics[index][0]++
	if !f.MatchMessage(msg) {
		// topic mismatch should have no affect, as topics are handled by topic matchers
		t.Fatalf("failed MatchEnvelope(topic mismatch) with seed %d.", seed)
	}
	f.Topics[index][0]--

	// key mismatch
	f.SymKeyHash[0]++
	if f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(key mismatch) with seed %d.", seed)
	}
	f.SymKeyHash[0]--

	// Src absent: match
	f.Src = nil
	if !f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(src absent) with seed %d.", seed)
	}

	// key hash mismatch
	h := f.SymKeyHash
	f.SymKeyHash = common.Hash{}
	if f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(key hash mismatch) with seed %d.", seed)
	}
	f.SymKeyHash = h
	if !f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(key hash match) with seed %d.", seed)
	}

	// encryption method mismatch
	f.KeySym = nil
	f.KeyAsym, err = crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed GenerateKey with seed %d: %s.", seed, err)
	}
	if f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(encryption method mismatch) with seed %d.", seed)
	}
}

func TestMatchMessageAsym(t *testing.T) {
	InitSingleTest()

	f, err := generateFilter(t, false)
	if err != nil {
		t.Fatalf("failed generateFilter with seed %d: %s.", seed, err)
	}

	params, err := generateMessageParams()
	if err != nil {
		t.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}

	const index = 1
	params.Topic = BytesToTopic(f.Topics[index])
	params.Dst = &f.KeyAsym.PublicKey
	keySymOrig := params.KeySym
	params.KeySym = nil

	sentMessage, err := NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err := sentMessage.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}
	msg := env.Open(f)
	if msg == nil {
		t.Fatalf("failed to open with seed %d.", seed)
	}

	// Src: match
	*f.Src.X = *params.Src.PublicKey.X
	*f.Src.Y = *params.Src.PublicKey.Y
	if !f.MatchMessage(msg) {
		t.Fatalf("failed MatchMessage(src match) with seed %d.", seed)
	}

	// insufficient PoW: mismatch
	f.PoW = msg.PoW + 1.0
	if f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(insufficient PoW) with seed %d.", seed)
	}

	// sufficient PoW: match
	f.PoW = msg.PoW / 2
	if !f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(sufficient PoW) with seed %d.", seed)
	}

	// topic mismatch
	f.Topics[index][0]++
	if !f.MatchMessage(msg) {
		// topic mismatch should have no affect, as topics are handled by topic matchers
		t.Fatalf("failed MatchEnvelope(topic mismatch) with seed %d.", seed)
	}
	f.Topics[index][0]--

	// key mismatch
	prev := *f.KeyAsym.PublicKey.X
	zero := *big.NewInt(0)
	*f.KeyAsym.PublicKey.X = zero
	if f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(key mismatch) with seed %d.", seed)
	}
	*f.KeyAsym.PublicKey.X = prev

	// Src absent: match
	f.Src = nil
	if !f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(src absent) with seed %d.", seed)
	}

	// encryption method mismatch
	f.KeySym = keySymOrig
	f.KeyAsym = nil
	if f.MatchMessage(msg) {
		t.Fatalf("failed MatchEnvelope(encryption method mismatch) with seed %d.", seed)
	}
}

func cloneFilter(orig *Filter) *Filter {
	var clone Filter
	clone.Messages = make(map[common.Hash]*ReceivedMessage)
	clone.Src = orig.Src
	clone.KeyAsym = orig.KeyAsym
	clone.KeySym = orig.KeySym
	clone.Topics = orig.Topics
	clone.PoW = orig.PoW
	clone.AllowP2P = orig.AllowP2P
	clone.SymKeyHash = orig.SymKeyHash
	return &clone
}

func generateCompatibeEnvelope(t *testing.T, f *Filter) *Envelope {
	params, err := generateMessageParams()
	if err != nil {
		t.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
		return nil
	}

	params.KeySym = f.KeySym
	params.Topic = BytesToTopic(f.Topics[2])
	sentMessage, err := NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err := sentMessage.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap with seed %d: %s.", seed, err)
		return nil
	}
	return env
}

func TestWatchers(t *testing.T) {
	InitSingleTest()

	const NumFilters = 16
	const NumMessages = 256
	var i int
	var j uint32
	var e *Envelope
	var x, firstID string
	var err error

	w := New(&Config{})
	filters := NewFilters(w)
	tst := generateTestCases(t, NumFilters)
	for i = 0; i < NumFilters; i++ {
		tst[i].f.Src = nil
		x, err = filters.Install(tst[i].f)
		if err != nil {
			t.Fatalf("failed to install filter with seed %d: %s.", seed, err)
		}
		tst[i].id = x
		if len(firstID) == 0 {
			firstID = x
		}
	}

	lastID := x

	var envelopes [NumMessages]*Envelope
	for i = 0; i < NumMessages; i++ {
		j = mrand.Uint32() % NumFilters
		e = generateCompatibeEnvelope(t, tst[j].f)
		envelopes[i] = e
		tst[j].msgCnt++
	}

	for i = 0; i < NumMessages; i++ {
		filters.NotifyWatchers(envelopes[i], false)
	}

	var total int
	var mail []*ReceivedMessage
	var count [NumFilters]int

	for i = 0; i < NumFilters; i++ {
		mail = tst[i].f.Retrieve()
		count[i] = len(mail)
		total += len(mail)
	}

	if total != NumMessages {
		t.Fatalf("failed with seed %d: total = %d, want: %d.", seed, total, NumMessages)
	}

	for i = 0; i < NumFilters; i++ {
		mail = tst[i].f.Retrieve()
		if len(mail) != 0 {
			t.Fatalf("failed with seed %d: i = %d.", seed, i)
		}

		if tst[i].msgCnt != count[i] {
			t.Fatalf("failed with seed %d: count[%d]: get %d, want %d.", seed, i, tst[i].msgCnt, count[i])
		}
	}

	// another round with a cloned filter

	clone := cloneFilter(tst[0].f)
	filters.Uninstall(lastID)
	total = 0
	last := NumFilters - 1
	tst[last].f = clone
	filters.Install(clone)
	for i = 0; i < NumFilters; i++ {
		tst[i].msgCnt = 0
		count[i] = 0
	}

	// make sure that the first watcher receives at least one message
	e = generateCompatibeEnvelope(t, tst[0].f)
	envelopes[0] = e
	tst[0].msgCnt++
	for i = 1; i < NumMessages; i++ {
		j = mrand.Uint32() % NumFilters
		e = generateCompatibeEnvelope(t, tst[j].f)
		envelopes[i] = e
		tst[j].msgCnt++
	}

	for i = 0; i < NumMessages; i++ {
		filters.NotifyWatchers(envelopes[i], false)
	}

	for i = 0; i < NumFilters; i++ {
		mail = tst[i].f.Retrieve()
		count[i] = len(mail)
		total += len(mail)
	}

	combined := tst[0].msgCnt + tst[last].msgCnt
	if total != NumMessages+count[0] {
		t.Fatalf("failed with seed %d: total = %d, count[0] = %d.", seed, total, count[0])
	}

	if combined != count[0] {
		t.Fatalf("failed with seed %d: combined = %d, count[0] = %d.", seed, combined, count[0])
	}

	if combined != count[last] {
		t.Fatalf("failed with seed %d: combined = %d, count[last] = %d.", seed, combined, count[last])
	}

	for i = 1; i < NumFilters-1; i++ {
		mail = tst[i].f.Retrieve()
		if len(mail) != 0 {
			t.Fatalf("failed with seed %d: i = %d.", seed, i)
		}

		if tst[i].msgCnt != count[i] {
			t.Fatalf("failed with seed %d: i = %d, get %d, want %d.", seed, i, tst[i].msgCnt, count[i])
		}
	}

	// test AcceptP2P

	total = 0
	filters.NotifyWatchers(envelopes[0], true)

	for i = 0; i < NumFilters; i++ {
		mail = tst[i].f.Retrieve()
		total += len(mail)
	}

	if total != 0 {
		t.Fatalf("failed with seed %d: total: got %d, want 0.", seed, total)
	}

	f := filters.Get(firstID)
	if f == nil {
		t.Fatalf("failed to get the filter with seed %d.", seed)
	}
	f.AllowP2P = true
	total = 0
	filters.NotifyWatchers(envelopes[0], true)

	for i = 0; i < NumFilters; i++ {
		mail = tst[i].f.Retrieve()
		total += len(mail)
	}

	if total != 1 {
		t.Fatalf("failed with seed %d: total: got %d, want 1.", seed, total)
	}
}

func TestVariableTopics(t *testing.T) {
	InitSingleTest()

	const lastTopicByte = 3
	var match bool
	params, err := generateMessageParams()
	if err != nil {
		t.Fatalf("failed generateMessageParams with seed %d: %s.", seed, err)
	}
	msg, err := NewSentMessage(params)
	if err != nil {
		t.Fatalf("failed to create new message with seed %d: %s.", seed, err)
	}
	env, err := msg.Wrap(params)
	if err != nil {
		t.Fatalf("failed Wrap with seed %d: %s.", seed, err)
	}

	f, err := generateFilter(t, true)
	if err != nil {
		t.Fatalf("failed generateFilter with seed %d: %s.", seed, err)
	}

	for i := 0; i < 4; i++ {
		env.Topic = BytesToTopic(f.Topics[i])
		match = f.MatchEnvelope(env)
		if !match {
			t.Fatalf("failed MatchEnvelope symmetric with seed %d, step %d.", seed, i)
		}

		f.Topics[i][lastTopicByte]++
		match = f.MatchEnvelope(env)
		if !match {
			// topic mismatch should have no affect, as topics are handled by topic matchers
			t.Fatalf("MatchEnvelope symmetric with seed %d, step %d.", seed, i)
		}
	}
}
