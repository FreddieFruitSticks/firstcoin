# Firstcoin

https://firstcoin.link/

A blockchain to learn about the fundamentals of decentralised consensus, loosely based off the bitcoin protocol.

Firstcoin has a few notable features worth mentioning. Each node has the following features:

1. Independent verification of each transaction based on a list of criteria.
2. Independent aggregation of all the transactions in to each block, with a proof-of-work demonstration.
3. Independent block verification and chain construction
4. Independent chain selection based on the most proof of work (albeit very basic in this project)

These high level rules bring about the emergent consensus property of the blockchain. However, in this project miners do not compete for blocks. The first node is simply chosen for demonstration purposes.


# Notes on building this project

1. Frontend is React (`npm run build` or `npm start`). Golang server handles serving the react site (`cp -r build/ ../firstcoin/web/`). Caution, at the time of this writing `npm start` runs on `localhost:3000`, but the service apis need to hit `locahost:8080`. You'll need to changes the service apis to hit `localhost:8080`
2. Golang server (`go install ./...` then to run `firstcoin <port> <address of peer>` eg `firstcoin 8081 localhost:8080`. Main node omits `<address of peer>`)
3. Docker compose deals with container orchastration (`docker compose up -d` - wait a few seconds to let the servers start before trying `localhost:8080`)
4. Firstnode-cdk deals with the AWS service orchastration (`cdk deploy/destroy` - see cdk docs)
5. To SSH in the ec2 instance the ssh key is stored in AWS System Manager > Parameter Store
