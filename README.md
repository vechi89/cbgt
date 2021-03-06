cbgt
====

The cbgt project provides a golang library that helps manage
distributed partitions (or data shards) across an elastic cluster of
servers.

[![Build Status](https://travis-ci.org/couchbaselabs/cbgt.svg)](https://travis-ci.org/couchbaselabs/cbgt) [![Coverage Status](https://coveralls.io/repos/couchbaselabs/cbgt/badge.svg?branch=master&service=github)](https://coveralls.io/github/couchbaselabs/cbgt?branch=master)

#### Documentation

* [![GoDoc](https://godoc.org/github.com/couchbaselabs/cbgt?status.svg)](https://godoc.org/github.com/couchbaselabs/cbgt)
* [REST API Reference](http://labs.couchbase.com/cbft/api-ref/) -
  these REST API Reference docs come from cbft, which uses the cbgt
  library.
* [UI Screenshots](https://github.com/couchbaselabs/cbgt/issues/16) -
  these screenshots come from cbft, which uses the cbgt library.

NOTE: This library initializes math's random seed
(rand.Seed(time.Now().UTC().UnixNano())) for unique id generation.
