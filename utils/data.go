package utils

import (
	"Nezha/core"
	"Nezha/ethereum/go-ethereum/accounts/abi"
	"Nezha/ethereum/go-ethereum/common"
	"Nezha/ethereum/go-ethereum/core/vm"
	"Nezha/evm/levm"
	"Nezha/evm/levm/tools"
	"fmt"
	"github.com/chinuy/zipf"
	"github.com/panjf2000/ants"
	"math/big"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func txCollector(addrNum uint64, txNum int, skew float64) [][]*core.RWNode {
	var txs [][]*core.RWNode

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	z := zipf.NewZipf(r, skew, addrNum)
	//z := rand.NewZipf(r, skew, 1, addrNum)

	for i := 0; i < txNum; i++ {
		rAddr1 := z.Uint64()
		rAddr2 := z.Uint64()
		wAddr1 := z.Uint64()
		wAddr2 := z.Uint64()

		tx := core.CreateRWNode(strconv.FormatInt(int64(i), 10), uint32(i), [][]byte{[]byte(strconv.FormatUint(rAddr1, 10)),
			[]byte(strconv.FormatUint(rAddr2, 10))}, [][]byte{[]byte("1"), []byte("2")},
			[][]byte{[]byte(strconv.FormatUint(wAddr1, 10)), []byte(strconv.FormatUint(wAddr2, 10))},
			[][]byte{[]byte("1"), []byte("2")})
		txs = append(txs, tx)
	}

	return txs
}

// CaptureRWSet capture read/write set in a single thread
func CaptureRWSet(addrNum uint64, txNum int, skew float64, dbFile string) [][]*core.RWNode {
	var txs [][]*core.RWNode

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

		lvm := levm.New(dbFile, big.NewInt(0), fromAddr)
		lvm.NewAccount(fromAddr, big.NewInt(1e18))

		_, addr, _, err := lvm.DeployContract(fromAddr, binData)
		if err != nil {
			fmt.Println(err)
		}

		rand.Seed(time.Now().UnixNano())
		// random := rand.Intn(5)
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
		for {
			if addr2 != addr1 {
				break
			}
			addr2 = z.Uint64()
		}

		rMap, wMap := SelectFunctions2(lvm, fromAddr, addr, abiObject, function, addr1, addr2)

		// generate r/w set
		var rAddr [][]byte
		var rValue [][]byte
		var wAddr [][]byte
		var wValue [][]byte

		for key := range rMap {
			s := key.Bytes()
			v := rMap[key].Bytes()
			rAddr = append(rAddr, s)
			rValue = append(rValue, v)

			//s1 := ConvertByte2String(s)
			//v1 := ConvertByte2String(v)
			//fmt.Printf("T_%d, Read/value: %s%s\n", i, s1, v1)
		}

		for key := range wMap {
			s := key.Bytes()
			v := wMap[key].Bytes()
			wAddr = append(wAddr, s)
			wValue = append(wValue, v)

			//s1 := ConvertByte2String(s)
			//v1 := ConvertByte2String(v)
			//fmt.Printf("T_%d, Write/value: %s%s\n", i, s1, v1)
		}

		rwNodes := core.CreateRWNode(strconv.FormatInt(int64(i), 10), uint32(i), rAddr, rValue, wAddr, wValue)
		txs = append(txs, rwNodes)
	}

	return txs
}

// ConCaptureRWSet capture read/write set in multiple threads
func ConCaptureRWSet(addrNum uint64, txNum int, skew float64, dbFile string) [][]*core.RWNode {
	var evmPools []*levm.LEVM
	var fromAddress []common.Address
	var cAddress []common.Address
	var txs [][]*core.RWNode

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
		lvm := levm.New(dbFile, big.NewInt(0), fromAddr)
		lvm.NewAccount(fromAddr, big.NewInt(1e18))

		evmPools = append(evmPools, lvm)

		_, addr, _, err := lvm.DeployContract(fromAddr, binData)
		if err != nil {
			fmt.Println(err)
		}

		cAddress = append(cAddress, addr)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup
	var lock sync.Mutex
	//var rw = make(chan []*core.RWNode, txNum)

	p, _ := ants.NewPoolWithFunc(100000, func(i interface{}) {
		n := i.(int)
		lvm := evmPools[n]
		fromAddr := fromAddress[n]
		addr := cAddress[n]

		rand.Seed(time.Now().UnixNano())
		random := rand.Intn(4)

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

		rMap, wMap := SelectFunctions2(lvm, fromAddr, addr, abiObject, function, addr1, addr2)

		// generate r/w set
		var rAddr [][]byte
		var rValue [][]byte
		var wAddr [][]byte
		var wValue [][]byte

		for key := range rMap {
			s := key.Bytes()
			v := rMap[key].Bytes()
			rAddr = append(rAddr, s)
			rValue = append(rValue, v)
		}

		for key := range wMap {
			s := key.Bytes()
			v := wMap[key].Bytes()
			wAddr = append(wAddr, s)
			wValue = append(wValue, v)
		}

		rwNodes := core.CreateRWNode(strconv.FormatInt(int64(n), 10), uint32(n), rAddr, rValue, wAddr, wValue)
		lock.Lock()
		txs = append(txs, rwNodes)
		lock.Unlock()

		wg.Done()
	})
	defer p.Release()

	for i := 0; i < txNum; i++ {
		wg.Add(1)
		_ = p.Invoke(i)
	}

	wg.Wait()

	return txs
}

func SelectFunctions(lvm *levm.LEVM, fromAddr common.Address, cAddr common.Address, abiObject abi.ABI, funcName string,
	addr1 uint64, addr2 uint64) {
	switch funcName {
	case "almagate":
		_, err := lvm.CallContractABI(fromAddr, cAddr, big.NewInt(0), abiObject, "almagate",
			strconv.FormatUint(addr1, 10), strconv.FormatUint(addr2, 10))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		//fmt.Println("get output:", getOutput)
	case "getBalance":
		_, err := lvm.CallContractABI(fromAddr, cAddr, big.NewInt(0), abiObject, "getBalance",
			strconv.FormatUint(addr1, 10))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		//fmt.Println("get output:", getOutput)
	case "updateBalance":
		_, err := lvm.CallContractABI(fromAddr, cAddr, big.NewInt(0), abiObject, "updateBalance",
			strconv.FormatUint(addr2, 10), big.NewInt(100))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		//fmt.Println("get output:", getOutput)
	case "updateSaving":
		_, err := lvm.CallContractABI(fromAddr, cAddr, big.NewInt(0), abiObject, "updateSaving",
			strconv.FormatUint(addr1, 10), big.NewInt(100))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		//fmt.Println("get output:", getOutput)
	case "sendPayment":
		_, err := lvm.CallContractABI(fromAddr, cAddr, big.NewInt(0), abiObject, "sendPayment",
			strconv.FormatUint(addr1, 10), strconv.FormatUint(addr2, 10), big.NewInt(50))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		//fmt.Println("get output:", getOutput)
	case "writeCheck":
		_, err := lvm.CallContractABI(fromAddr, cAddr, big.NewInt(0), abiObject, "writeCheck",
			strconv.FormatUint(addr2, 10), big.NewInt(20))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		//fmt.Println("get output:", getOutput)
	default:
		fmt.Println("Invalid inputs")
	}
}

func SelectFunctions2(lvm *levm.LEVM, fromAddr common.Address, cAddr common.Address, abiObject abi.ABI, funcName string,
	addr1 uint64, addr2 uint64) (vm.Storage, vm.Storage) {
	switch funcName {
	case "almagate":
		rMap, wMap, _, err := lvm.CallContractABI2(fromAddr, cAddr, big.NewInt(0), abiObject, "almagate",
			strconv.FormatUint(addr1, 10), strconv.FormatUint(addr2, 10))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		return rMap, wMap
	case "getBalance":
		rMap, wMap, _, err := lvm.CallContractABI2(fromAddr, cAddr, big.NewInt(0), abiObject, "getBalance",
			strconv.FormatUint(addr1, 10))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		return rMap, wMap
	case "updateBalance":
		rMap, wMap, _, err := lvm.CallContractABI2(fromAddr, cAddr, big.NewInt(0), abiObject, "updateBalance",
			strconv.FormatUint(addr2, 10), big.NewInt(100))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		return rMap, wMap
	case "updateSaving":
		rMap, wMap, _, err := lvm.CallContractABI2(fromAddr, cAddr, big.NewInt(0), abiObject, "updateSaving",
			strconv.FormatUint(addr1, 10), big.NewInt(100))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		return rMap, wMap
	case "sendPayment":
		rMap, wMap, _, err := lvm.CallContractABI2(fromAddr, cAddr, big.NewInt(0), abiObject, "sendPayment",
			strconv.FormatUint(addr1, 10), strconv.FormatUint(addr2, 10), big.NewInt(50))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		return rMap, wMap
	case "writeCheck":
		rMap, wMap, _, err := lvm.CallContractABI2(fromAddr, cAddr, big.NewInt(0), abiObject, "writeCheck",
			strconv.FormatUint(addr2, 10), big.NewInt(20))
		if err != nil {
			fmt.Println("get error : ", err)
		}
		return rMap, wMap
	default:
		fmt.Println("Invalid inputs")
		return nil, nil
	}
}

func ProcessRWMap(rMap, wMap vm.Storage) (map[string]string, map[string]string) {
	var readSet = make(map[string]string)
	var writeSet = make(map[string]string)

	for key := range rMap {
		s := key.Bytes()
		v := rMap[key].Bytes()
		readSet[string(s)] = string(v)
	}

	for key := range wMap {
		s := key.Bytes()
		v := wMap[key].Bytes()
		writeSet[string(s)] = string(v)
	}

	return readSet, writeSet
}
