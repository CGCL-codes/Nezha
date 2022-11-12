pragma solidity 0.5.8;  //The lowest compiler version

contract Example {
    
    int public vala;
    int public valb;
    
    function SetValA(int a) public {
        vala = a;
    }
    
    function SetValB(int b) public {
        valb = b;
    }
}
