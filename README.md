# Regen: Redundancy Generator

## Table of Contents

- [About](#about)
- [Getting Started](#getting_started)
- [Usage](#usage)
- [Benchmarking](#benchmarking)

## About <a name = "about"></a>

Regen is a data redundancy generator for archive files. It uses a brand new technique to produce redundancy on logical data and employs computing power in the process of data recovery. This allows the technology to be used for the purpose of long term archive data preservation where multiple storage devices are not available.

## Getting Started <a name = "getting_started"></a>

This repository contains a Go module with packages "regen" and "cmd". The regen package contains exported functions, giving access to all of the functionality for use in imported projects. Function documentation is available in [Go packages website](https://pkg.go.dev/github.com/sharuvan/regen).

### Prerequisites

Go compiler is required to install and build the program. It can be downloaded from [Go website](https://go.dev/doc/install). The program is compatible with all major desktop platforms and CPU architectures.


### Installing

To add the regen package into your Go module:
```
go get github.com/sharuvan/regen/regen
```

To build and install the CLI program from cmd package:
```
go install github.com/sharuvan/regen@latest
```

## Usage <a name = "usage"></a>

Available commands and flags can be explored using:
```
regen help
```

To generate redundancy on an archive file:
```
regen generate --file cats.zip --percentage 5 --checksum 64
```

To verify integrity of archive file:
```
regen verify --file cats.zip
```

To regenerate archive file when corruption is detected:
```
regen regenerate --file cats.zip
```

A checksum block size of 64 bytes are more appropriate for archives smaller than 1GB. Using 128 byte checksum blocks will significantly decrease the generate time.

## Benchmarking <a name = "benchmarking"></a>

The repository contains a sample archive file "cats.zip" in the root directory. Copy this into "testdata" directory and set configurations in file "regenerate_test.go" before executing benchmark tests with Go tools.

To execute burst error benchmarking:
```
cp cats.zip testdata/cats.zip
go test -run ^Benchmark -bench BenchmarkRandomBurst -count 1 -benchtime=1x -timeout 0 ./regen
```
To execute bit error benchmarking:
```
cp cats.zip testdata/cats.zip
go test -run ^Benchmark -bench BenchmarkRandomBit -count 1 -benchtime=1x -timeout 0 ./regen
```