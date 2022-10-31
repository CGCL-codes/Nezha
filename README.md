# Nezha

The implementation of the paper: *"Nezha: Exploiting Concurrency for Transaction Processing in DAG-based Blockchains" ICDCS'22*

Nezha is the first efficient concurrency control mechanism towards DAG-based blockchains. It contians an address-based conflict graph (ACG) while incorporating address dependencies as edges to capture all conflicting transactions. Besides, it contains a hierarchical sorting (HS) algorithm to derive a total order between transactions based on ACG.

This repository contains a testing tool for evaluating the performance of Nezha concurrency control and two baseline schemes: SP (serial processing) and CG (conflict graph). Note that, in the paper, we integrate Nezha into a prevailing DAG-based blockchain system OHIE, but we do not present the demo of OHIE in this repository (just for testing). We will upload a complete DAG-based blockchain demo containing Nezha transaction processing module in the future.

## Requriments

To build and deploy Nezha, you need to install golang (version >= 1.16).

Latest Golang versions for various operating systems are detailed [here](https://go.dev/dl/).

## Datasets

By default, we use smallbank smart contract in the directory `SmallBank` to issue transactions. SmallBanck contract contains 7 functions to simulate transfer actions between differen bank accounts. We utilize an evm calling tool (https://github.com/CryptoKass/levm) to call smart contracts without the need for Ethereum blockchain. You can use your own custom smart contract if you want, but you need to generate `.bin` and `.abi` files when you use it. For more details on how to compile smart contracts, please refer to https://goethereumbook.org/en/smart-contract-compile/.

## How to use

### Build codes

`go build test.go`

### Run tests

`./test -a=$1 -t=$2 -s=$3`

where `-a` denotes the number of calling account addresses, `-t` denotes the number of generated transactions, and `-s` denotes the Zipfian skew value (the maximum skew value is 2.0). Please note that the skew value is better set below 1.5 since the CG scheme consumes considerable memory resources. When the skew value is set too high, the process will be killed due to the out of memory. In the paper, we utilize 64GB RAM to run the set of experiments.

### Check results

`cat ./Exp_results.txt `

You can see the comparison results between SP, Nezha, and CG schemes. We also present the transaction simulation latency at the last line.

## Citation

If you are using our codes for research purpose, please cite our paper

```
@inproceedings{xiao2022nezha,
  title={Nezha: Exploiting Concurrency for Transaction Processing in DAG-based Blockchains},
  author={Xiao, Jiang and Zhang, Shijie and Zhang, Zhiwei and Li, Bo and Dai, Xiaohai and Jin, Hai},
  booktitle={2022 IEEE 42nd International Conference on Distributed Computing Systems (ICDCS)},
  pages={269--279},
  year={2022},
  organization={IEEE}
}
```



