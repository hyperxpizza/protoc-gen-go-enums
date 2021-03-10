# protoc-gen-go-enums

A plugin for [Protocol Buffers](https://developers.google.com/protocol-buffers) that generates XML and JSON marshal receivers for enums generated from `.proto` files. It enables enums to be marshaled/unmarshaled from their string variants.

To see an example output go to `e2e\e2e.xml.go`

## Prerequisites

- [Golang](https://golang.org/)
- [Protocol Buffer compiler](https://grpc.io/docs/protoc-installation/)
- [Go Task](https://github.com/go-task/task)

## How to build

For building the package run:

```
task build
```

or

```
go build -o protoc-gen-go-enums .
```

## How to run

There is an example `.proto` file at `e2e\e2e.proto` for generating it run the following command:

```
task gen_e2e_proto
```

or

```
protoc --plugin protoc-gen-go-enums --go-enums_out=json:e2e --go-enums_out=xml:e2e --proto_path=. --go_out=plugins=grpc:e2e e2e/e2e.proto
```

The `xml` or `json` prefix selects if a `.json.go` or `.xml.go` or both should be generated.

The output file will be `e2e\e2e.xml.go` and `e2e/e2e.json.go`

## How to run tests

```
task run_e2e
```

or if you want to rebuild the package and regenerate the e2e.xml.go file run

```
task e2e_all
```