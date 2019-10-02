# Testnet

The purpose of this package is to enable the creation of docker-compose testnets in unit tests.

The tests are longer-running than the standard unit tests, so they are only invoked when the 
build tag `bdd` is added, e.g. `go test -v -tags=bdd`. The BDD tests log rather frivolously, but 
can be quited by adding the flag `quiet` e.g.`go test -v -tags="bdd quiet"`.   