# splashmerkle - WIP:
[![GODOC](https://godoc.org/github.com/CryptoKass/splashmerkle?status.svg)](https://godoc.org/github.com/CryptoKass/splashmerkle) 


# Getting stated:
```shell
go get github.com/CryptoKass/splashmerkle
```
If you do not have the go command on your system, you need to [Install Go](http://golang.org/doc/install) first

<br></br>
## quick start:
```
//make a tree
tree := splashmerkle.Tree{}
tree.ConstructTree{ inputs }

//get the merkle root
merkleroot := tree.Root.Bytes()
```

<br></br>
# Documentaion:
Visit [GoDoc](https://godoc.org/github.com/CryptoKass/splashmerkle) 



# Contribution: 
**If I got something wrong (which I almost certainly have) please let me know:**
- Pull requests welcomed!
- Feedback: cryptokass@gmail.com

---

*Readme last updated: 2018.12.27*