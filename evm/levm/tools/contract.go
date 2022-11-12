package tools

import (
	"Nezha/ethereum/go-ethereum/accounts/abi"
	"Nezha/ethereum/go-ethereum/common/hexutil"
	"io/ioutil"
	"os"
)

// LoadContract will open and decode a contracts
// Application Blockchain Interface and Binary files.
func LoadContract(abiPath, binPath string) (abi.ABI, []byte, error) {

	// load ABI
	abiFile, err := os.Open(abiPath)
	if err != nil {
		return abi.ABI{}, nil, err
	}
	abiObject, err := abi.JSON(abiFile)
	if err != nil {
		return abiObject, nil, err
	}

	//load and decode bin
	binRaw, err := ioutil.ReadFile(binPath)
	if err != nil {
		return abiObject, nil, err
	}
	binData, err := hexutil.Decode("0x" + string(binRaw))

	return abiObject, binData, err
}
