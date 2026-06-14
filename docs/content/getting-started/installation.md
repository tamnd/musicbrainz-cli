---
title: "Installation"
description: "Install musicbrainz from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/musicbrainz-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `musicbrainz` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/musicbrainz-cli/cmd/musicbrainz@latest
```

That puts `musicbrainz` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/musicbrainz-cli
cd musicbrainz-cli
make build        # produces ./bin/musicbrainz
./bin/musicbrainz version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/musicbrainz:latest --help
```

## Checking the install

```bash
musicbrainz version
```

prints the version and exits.
