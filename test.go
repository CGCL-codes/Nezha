package main

import (
	"Nezha/core"
	"Nezha/ethereum/go-ethereum/common"
	ecore "Nezha/ethereum/go-ethereum/core"
	"Nezha/evm/levm"
	"Nezha/evm/levm/tools"
	"Nezha/graph"
	"Nezha/utils"
	"bufio"
	"flag"
	"fmt"
	"github.com/chinuy/zipf"
	"github.com/panjf2000/ants"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"
)

const dbFile1 = "DAG_CG"
const dbFile2 = "DAG_ACG"
const dbFile3 = "DAG_Serial"
const dbFile4 = "DAG_Sim"
const dbFile5 = "DAG_Con"
const dbFile6 = "Eth_Test"
const fileName = "Exp_results.txt"

func main() {
	var addrNum uint64
	var txNum int
	var skew float64
	var blksize int
	var con int

	flag.Uint64Var(&addrNum, "a", 10000, "specify address number to use. defaults to 10000.")
	flag.IntVar(&txNum, "t", 200, "specify transaction number to use. defaults to 100.")
	flag.Float64Var(&skew, "s", 0.6, "specify skew to use. defaults to 0.2.")
	flag.IntVar(&blksize, "b", 200, "specify block size to use. defaults to 200.")
	flag.IntVar(&con, "c", 4, "specify block size to use. defaults to 4.")

	flag.Parse()

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	TestSerialExecution(addrNum, txNum, skew, w)
	TestConflictQueue(addrNum, txNum, skew, w, dbFile4)
	TestConflictGraph(addrNum, txNum, skew, w, dbFile4)
	TestSimulation(addrNum, txNum, skew, w)
}

// TestSimulation test concurrent transaction simulations
func TestSimulation(addrNum uint64, txNum int, skew float64, writer *bufio.Writer) {
	var evmPools []*levm.LEVM
	var fromAddress []common.Address
	var cAddress []common.Address

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	z := zipf.NewZipf(r, skew, addrNum)

	selectFunc := []string{"almagate", "updateBalance", "updateSaving", "sendPayment", "writeCheck", "getBalance"}

	abiObject, binData, err := tools.LoadContract("./SmallBank/small_bank_sol_SmallBank.abi",
		"./SmallBank/small_bank_sol_SmallBank.bin")
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < txNum; i++ {
		fromAddr := tools.NewRandomAddress()
		fromAddress = append(fromAddress, fromAddr)
		// create EVM instances
		lvm := levm.New(dbFile4, big.NewInt(0), fromAddr)
		lvm.NewAccount(fromAddr, big.NewInt(1e18))

		evmPools = append(evmPools, lvm)

		_, addr, _, err := lvm.DeployContract(fromAddr, binData)
		if err != nil {
			fmt.Println(err)
		}

		cAddress = append(cAddress, addr)
	}

	//fmt.Println(runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(100000, func(i interface{}) {
		n := i.(int)
		lvm := evmPools[n]
		fromAddr := fromAddress[n]
		addr := cAddress[n]

		rand.Seed(time.Now().UnixNano())
		//random := rand.Intn(4)
		random := rand.Float32()

		// read-write 50-50
		var function string
		if random <= 0.05 {
			function = selectFunc[5]
		} else {
			random2 := rand.Intn(5)
			function = selectFunc[random2]
		}

		addr1 := z.Uint64()
		addr2 := z.Uint64()

		utils.SelectFunctions(lvm, fromAddr, addr, abiObject, function, addr1, addr2)

		wg.Done()
	})
	defer p.Release()

	start := time.Now()

	wg.Add(1)
	go func() {
		for i := 0; i < txNum; i++ {
			wg.Add(1)
			_ = p.Invoke(i)
		}
		wg.Done()
	}()

	wg.Wait()
	duration := time.Since(start)
	writer.WriteString(fmt.Sprintf("Time of concurrently simulating transactions: %s\n", duration))
	writer.WriteString(fmt.Sprintf("===================================================\n"))
	writer.Flush()
}

// TestConflictGraph test concurrency control performance of CG
func TestConflictGraph(addrNum uint64, txNum int, skew float64, writer *bufio.Writer, dbFile string) {
	var al core.AlGraph
	var inValidTxs []int
	// concurrently simulate transactions to capture read/write sets
	txs := utils.ConCaptureRWSet(addrNum, txNum, skew, dbFile)
	start := time.Now()

	start1 := time.Now()
	// create conflict graph
	gSlice := core.NewBuildConflictGraph(txs)
	al.Init(gSlice)
	duration1 := time.Since(start1)
	writer.WriteString(fmt.Sprintf("Time of constructing cg: %s\n", duration1))

	start2 := time.Now()
	// cycle detection
	johnsonDAG := graph.NewJohnsonCE(&gSlice)
	abortedNum, abortedTx := johnsonDAG.Run()
	duration2 := time.Since(start2)
	writer.WriteString(fmt.Sprintf("Time of finding and removing cycles: %s\n", duration2))

	for i, t := range abortedTx {
		if t == true {
			inValidTxs = append(inValidTxs, i)
		}
	}

	start3 := time.Now()
	// topological sorting
	al.RebuildGraph(inValidTxs)
	commitOrder := al.BasicTopologicalSort()
	duration3 := time.Since(start3)
	writer.WriteString(fmt.Sprintf("Time of topological sorting: %s\n", duration3))

	db := OpenDB(dbFile1)

	start4 := time.Now()
	// commit transactions
	for _, v := range commitOrder {
		for _, vv := range txs[v] {
			if vv.Label == "w" {
				acc := core.CreateAccount(vv.RWSet.Key, vv.RWSet.Value)
				err := utils.StoreState(db, acc)
				if err != nil {
					log.Panic(err)
				}
			}
		}
	}
	duration4 := time.Since(start4)
	writer.WriteString(fmt.Sprintf("Time of committing transactions: %s\n", duration4))

	duration := time.Since(start)

	writer.WriteString(fmt.Sprintf("Abort rate is: %.3f\n", float64(abortedNum)/float64(len(txs))))
	writer.WriteString(fmt.Sprintf("Time of processing TXs on CG: %s\n", duration))
	writer.WriteString(fmt.Sprintf("===================================================\n"))
	writer.Flush()
}

// TestConflictQueue test concurrency control performance of ACG
func TestConflictQueue(addrNum uint64, txNum int, skew float64, writer *bufio.Writer, dbFile string) {
	// concurrently simulate transactions to capture read/write sets
	txs := utils.ConCaptureRWSet(addrNum, txNum, skew, dbFile)

	start := time.Now()

	start1 := time.Now()
	// create conflict graph
	queueGraph := core.CreateGraph(txs)
	duration1 := time.Since(start1)
	writer.WriteString(fmt.Sprintf("Time of graph construction: %s\n", duration1))

	start2 := time.Now()
	// rank division
	sequence := queueGraph.QueuesSort()
	duration2 := time.Since(start2)
	writer.WriteString(fmt.Sprintf("Time of rank divsion: %s\n", duration2))

	start3 := time.Now()
	// sorting
	commitOrder := queueGraph.DeSS(sequence)
	duration3 := time.Since(start3)
	writer.WriteString(fmt.Sprintf("Time of DeSS: %s\n", duration3))

	var keys []int
	for seq := range commitOrder {
		keys = append(keys, int(seq))
	}
	sort.Ints(keys)

	db := OpenDB(dbFile2)

	start4 := time.Now()
	// concurrently commit transactions
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(2000, func(i interface{}) {
		n := i.([]*core.RWNode)
		for _, rw := range n {
			acc := core.CreateAccount(rw.RWSet.Key, rw.RWSet.Value)
			err := utils.StoreState(db, acc)
			if err != nil {
				log.Panic(err)
			}
		}
		wg.Done()
	})
	defer p.Release()

	for _, n := range keys {
		for _, v := range commitOrder[int32(n)] {
			if len(v) > 0 {
				wg.Add(1)
				_ = p.Invoke(v)
			}
		}
		wg.Wait()
	}
	duration4 := time.Since(start4)
	writer.WriteString(fmt.Sprintf("Time of committing transactions: %s\n", duration4))

	duration := time.Since(start)
	count := queueGraph.GetAbortedNums()

	writer.WriteString(fmt.Sprintf("Abort rate is: %.3f\n", float64(count)/float64(len(txs))))
	writer.WriteString(fmt.Sprintf("Time of processing TXs on Nezha: %s\n", duration))
	writer.WriteString(fmt.Sprintf("===================================================\n"))
	writer.Flush()
}

// TestSerialExecution test serial transaction processing
func TestSerialExecution(addrNum uint64, txNum int, skew float64, writer *bufio.Writer) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	z := zipf.NewZipf(r, skew, addrNum)

	fromAddr := tools.NewRandomAddress()
	abiObject, binData, err := tools.LoadContract("./SmallBank/smallbank_sol_SmallBank.abi",
		"./SmallBank/smallbank_sol_SmallBank.bin")
	if err != nil {
		fmt.Println(err)
	}
	lvm := levm.New(dbFile3, big.NewInt(0), fromAddr)

	lvm.NewAccount(fromAddr, big.NewInt(1e18))

	// deploy a contract
	_, addr, _, err := lvm.DeployContract(fromAddr, binData)
	if err != nil {
		fmt.Println(err)
	}

	selectFunc := []string{"almagate", "updateBalance", "updateSaving", "sendPayment", "writeCheck", "getBalance"}

	start := time.Now()

	// call smart contracts to update the state
	for i := 0; i < txNum; i++ {
		rand.Seed(time.Now().UnixNano())
		random := rand.Intn(2)

		// read-write 50-50
		var function string
		if random == 0 {
			function = selectFunc[5]
		} else {
			random2 := rand.Intn(5)
			function = selectFunc[random2]
		}

		addr1 := z.Uint64()
		addr2 := z.Uint64()

		utils.SelectFunctions(lvm, fromAddr, addr, abiObject, function, addr1, addr2)
	}

	stateDB := lvm.GetStateDB()
	// obtain the root hash of MPT
	root := stateDB.IntermediateRoot(false)
	stateDB.Commit(false)
	stateDB.Database().TrieDB().Commit(root, true)

	duration := time.Since(start)
	writer.WriteString(fmt.Sprintf("===================================================\n"))
	writer.WriteString(fmt.Sprintf("Time of serial transaction processing: %s\n", duration))
	writer.WriteString(fmt.Sprintf("===================================================\n"))
	writer.Flush()
}

func TestAppConcurrency(txNum int, blksize int, con int, addrNum uint64, skew float64) {
	avgNum := con * blksize
	loop := math.Ceil(float64(txNum / avgNum))
	count := 0
	db := OpenDB(dbFile5)
	var wg sync.WaitGroup

	runtime.GOMAXPROCS(runtime.NumCPU())

	p, _ := ants.NewPoolWithFunc(100000, func(i interface{}) {
		n := i.([]*core.RWNode)
		for _, rw := range n {
			acc := core.CreateAccount(rw.RWSet.Key, rw.RWSet.Value)
			err := utils.StoreState(db, acc)
			if err != nil {
				log.Panic(err)
			}
		}
		wg.Done()
	})
	defer p.Release()

	start := time.Now()

	for i := 0; i < int(loop); i++ {
		var exeNum int
		var keys []int

		if i == int(loop)-1 {
			exeNum = txNum - i*avgNum
		} else {
			exeNum = avgNum
		}

		txs := utils.ConCaptureRWSet(addrNum, exeNum, skew, dbFile5)
		queueGraph := core.CreateGraph(txs)
		sequence := queueGraph.QueuesSort()
		commitOrder := queueGraph.DeSS(sequence)

		for seq := range commitOrder {
			keys = append(keys, int(seq))
		}
		sort.Ints(keys)

		for _, n := range keys {
			for _, v := range commitOrder[int32(n)] {
				if len(v) > 0 {
					wg.Add(1)
					_ = p.Invoke(v)
				}
			}

			wg.Wait()
		}

		abortedNum := queueGraph.GetAbortedNums()
		count += abortedNum

		// simulate the latency of committing
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(start)
	fmt.Printf("Time of processing transactions: %s\n", duration)
	fmt.Printf("Abort rate is: %.3f\n", float64(count)/float64(txNum))
}

// TestReplayingTx test a single transaction's replaying
func TestReplayingTx(nonce uint64, from, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (map[string]string, map[string]string, []byte, error) {
	var tx *core.EthTransaction

	// verdict if it is a contract creation tx
	if &to == nil {
		tx = core.NewContractCreation(nonce, from, amount, gasLimit, gasPrice, data)
	} else {
		tx = core.NewEthTransaction(nonce, from, to, amount, gasLimit, gasPrice, data)
	}

	lvm := levm.New(dbFile6, big.NewInt(0), tx.From())
	gasPool := new(ecore.GasPool).AddGas(uint64(1000000000))

	rMap, wMap, output, err := lvm.ReplayTransaction(*tx, gasPool)
	if err != nil {
		return nil, nil, nil, err
	}

	// commit to the database
	stateDB := lvm.GetStateDB()
	root := stateDB.IntermediateRoot(false)
	stateDB.Commit(false)
	stateDB.Database().TrieDB().Commit(root, true)

	if rMap != nil && wMap != nil {
		readSet, writeSet := utils.ProcessRWMap(rMap, wMap)
		return readSet, writeSet, output, nil
	}

	return nil, nil, output, nil
}

func OpenDB(dbFile string) *leveldb.DB {
	db, err := utils.LoadDB(dbFile)
	if err != nil {
		log.Panic(err)
	}

	return db
}
