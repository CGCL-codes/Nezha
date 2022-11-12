# levm - Little Ethereum Virtual Machine

Run smart contracts on the Ethereum Virtual Machine (EVM) without the Ethereum blockchain. 

I created this to make it a little easier to use the EVM in other projects, and for testing smart contracts.

`levm.New(...)` creates a small wrapper for the ethereum virtual machine, and sets up the EVM with default parameters and the required stateDB.

# ↓ Installation

This project only has two dependencies. (go-ethereum contains the EVM, and splashkeys is an ecdsa key library)

**`go get github.com/CryptoKass/levm`**

### Dependencies
- `go get github.com/ethereum/go-ethereum`
- `go get github.com/CryptoKass/splashkeys`

<br>

# Example 
This is Example is located in the example sub-directory:

```go
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
```

contact: kasscrypto@gmail.com 👍

**HELP WANTED** would be great if you would create an issue if you find a bug.