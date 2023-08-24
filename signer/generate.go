package main

//go:generate mockgen --source models/queue.go -destination mocks/queue.go -package mocks
//go:generate mockgen --source models/datastore.go -destination mocks/datastore.go -package mocks -mock_names Datastore=Store
//go:generate mockgen --source models/signer.go -destination mocks/signer.go -package mocks
