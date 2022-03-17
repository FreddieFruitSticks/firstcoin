# firstcoin

A blockchain to learn about the fundamentals of decentralised consensus, loosely based off the bitcoin protocol.

Firstcoin has a few notable features worth mentioning. Each node has the following features:

1. Independent verification of each transaction based on a list of criteria.
2. Independent aggregation of all the transactions in to each block, with a proof-of-work demonstration.
3. Independent block verification and chain construction
4. Independent chain selection based on the most proof of work (albeit very basic in this project)

These high level rules bring about the emergent consensus property of the blockchain. However, in this project miners do not compete for blocks. The first node is simply chosen for demonstration purposes.
