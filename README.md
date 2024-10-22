# Chitty Chat

## Description

The project is a mandatory assignment for Distributed Systems course at the IT University of Copenhagen.

The aim of the project is to create a GRPC-based chat application that allows clients to send messages and the server broadcasts them to all connected clients.

## Running the Code

To run the code, you need to have Go installed on your machine. You can download it from [here](https://go.dev/dl/).

Run the server:

```bash
cd server
go run server.go
```

In a separate terminal, run the client.

```bash
cd client
go run client.go
```

Send messages on the client terminal and see them being broadcasted to all connected clients.