# txgroup

Package txgroup provides a simple way of handling multiple transactions from different databases as if they were only one, facilitating their progration, cancelation and execution.

> Note that it does not guarantee atomicity among transactions, one may succeed and the following one fail, leaving the system with an inconsistent state.

For an example on how to use the package head over to [example](./example).

`go get -u github.com/GGP1/txgroup`