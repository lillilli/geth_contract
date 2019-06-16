pragma solidity ^0.5.9;

contract Counter {
    int private count = 0;

    function incrementCounter() public {
        count += 1;
    }

    function decrementCounter() public {
        count -= 1;
    }

    function getCount() public view returns (int) {
        return count;
    }
}