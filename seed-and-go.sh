#!/bin/bash

cd cmd/seedlouds
go build
./seedlouds
cd ../..
go build
./go-loud
