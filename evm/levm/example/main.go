package main

import (
	"Nezha/evm/levm"
	"Nezha/evm/levm/tools"
	"fmt"
	"math/big"
)

var (
	fromAddr = tools.NewRandomAddress()
)

func main() {

	//make a new address evm
	fromAddr = tools.NewRandomAddress()

	//Load a contract from file
	abiObject, binData, err := tools.LoadContract("contract/example_sol_Example.abi", "contract/example_sol_Example.bin")
	fmt.Println("Abi\n", abiObject.Methods)

	// create new levm instance
	lvm := levm.New("./db", big.NewInt(0), fromAddr)

	// create a new account and set the balance
	// (needs enough balance to cover gas cost)
	lvm.NewAccount(fromAddr, big.NewInt(1e18))

	// deploy a contract
	code, addr, gas, err := lvm.DeployContract(fromAddr, binData)
	fmt.Println("contract code length:", len(code))
	fmt.Printf("contract address: %x\n", addr)
	fmt.Println("unused gas:", gas)
	fmt.Println("errors:", err)
}
