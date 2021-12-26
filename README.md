# firstcoin

A replica of the bitcoin protocol (work in progress).

Some features and limitations:

1. The SIGHASH is always assumed to be ALL
2. The only script is pay-to-public-key-hash
3. Transaction fees are fixed, and are not variable (e.g. not calculated based on bitsize of transaction).
4. Transaction serialisation is not explicit and transaction size limitations are not adhered to.
5. The implementation of the version prefix in the base58Check encoding only supports bitcoin addresses
6. The wallet is a single randomly generated bitcoin address

(more features and limitations to follow)
