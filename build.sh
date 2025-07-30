#!/bin/bash
gox -ldflags "-s -w ${LDFLAGS}" -output="build/goss_{{.OS}}_{{.Arch}}" --osarch="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64" ./cmd/goss