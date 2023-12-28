# CRDT-SQL

This is a few things, and I'm writing about it on my blog post. Read on for more.
1. The Why of [Local First Software](https://www.ersin.nz/articles/local-first-software)
2. My first foray into writing the stack, foundational setup [Creating the Local First Stack](https://www.ersin.nz/articles/creating-the-local-first-stack)

This is totally aspirational, but I'm hoping to blog about the following in this series.

- Using the CRDT stack to build a real application.
- Deploying and running a tiny CRDT stack in the cloud.
- A desktop client built on the local first stack
- Seemless offline web application built on the local first stack

## What we have so far

[cmd/server](cmd/server) is a simple server that exposes an RPC server for synchronization

[frontend](frontend) is a simple web frontend that uses the RPC server to synchronize data

[crsql](crsql) is a simple Golang CRDT SQL library that uses the RPC server to synchronize data
