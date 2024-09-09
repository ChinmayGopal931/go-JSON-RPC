// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract MOCKERC20PERMIT is ERC20Permit, Ownable {
    constructor(string memory _name, string memory _symbol, uint8 _decimals)
        ERC20Permit(_name)
        ERC20(_name, _symbol)
        Ownable(msg.sender)
    {}

    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
    }

    function burn(address from, uint256 value) public {
        require(from == _msgSender() || allowance(from, _msgSender()) >= value, "ERC20: burn amount exceeds allowance");
        _burn(from, value);
    }
}
