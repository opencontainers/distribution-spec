---
title: "HTTP API V2"
description: "Specification for the Registry API."
keywords: registry, on-prem, images, tags, repository, distribution, api, advanced
---
# Open Container Initiative Distribution Specification


## Table of Contents

- [Overview](#overview)
	- [Introduction](#introduction)
	- [Historical Context](#historical-context)
- [Definitions](#defintions)
	- [Common Terms](#introduction)
	- [Document Language](#document-language)
- [Conformance](#conforfmance)
	- [Minimum Requirements](#minimum-requirements)
	- [Official Certification](#official-certification)
	- [Workflow Categories](#workflow-categories)
		1. [Pull](#pull)
		2. [Push](#pull)
		3. [Pull](#pull)
		4. [Pull](#pull)
- [HTTP API](#http-api)
	- [Endpoints](#endpoints)
	- [Error Codes](#error-codes)

## Overview

### Introduction

The Open Container Initiative Distribution Specification defines an API protocol to facilitate and standardize the distribution of content, especially related to container images and associated artifacts.

### Historical Context

The spec is based on the specification for the [Docker Registry HTTP API V2 protocol](https://github.com/docker/distribution/blob/5cb406d511b7b9163bff9b6439072e4892e5ae3b/docs/spec/api.md).

For relevant details and a history leading up to this specification, please see the following issues:

- [moby/moby#8093](https://github.com/moby/moby/issues/8093)
- [moby/moby#9015](https://github.com/moby/moby/issues/9015)
- [docker/docker-registry#612](https://github.com/docker/docker-registry/issues/612)

## Definitions

### Common Terms

Several terms are used frequently in this document and warrant basic definitions:

- **Registry**: a HTTP service which implements this spec
- **Client**: a tool that communicates with registries over HTTP
- **Push**: the act of uploading content to a registry
- **Pull**: the act of downloading content from a registry
- **Artifact**: a single piece of content, made up of a manifest and one or more layers
- **Manifest**: a JSON document which defines an artifact
- **Layer**: a single part of all the parts which comprise an artifact
- **Config**: a special layer defined at the top of a manifest containing artifact metadata
- **Blob**: a single binary content stored in a registry
- **Digest**: a unique blob identifier
- **Content**: a general term for content that can be downloaded from a registry (manifest or blob)

### Document Language

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119) (Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", BCP 14, RFC 2119, March 1997).

## Conformance

### Minimum Requirements

For a registry to be considered fully conformant against this specification, it must implement the HTTP endpoints required by each of the four (4) major workflow categories:

1. **Pull** (REQUIRED) - Ability to fetch content from a registry
2. **Push** - Ability to publish content to a registry
3. **Content Discovery** - Ability to list or otherwise query the content stored in a registry
4. **Content Management** - Ability to delete (or otherwise manipulate) content stored in a registry

At a bare minimum, registries claiming to be "OCI-Compliant" MUST support all facets of the pull workflow.

In order to test a registry's conformance against these workflows, please use the [conformance testing tool](./conformance/).

### Official Certification

Registry providers can self-cetify by submitting conformance results to [opencontainers/oci-conformance](https://github.com/opencontainers/oci-conformance).

### Workflow Categories

#### Pull

TODO: describe the Pull category and the high-level details

#### Push

TODO: describe the Push category and the high-level details

#### Content Discovery

TODO: describe the Content Discovery category and the high-level details

#### Content Management

TODO: describe the Content Management category and the high-level details

## HTTP API

### Endpoints

TODO: table of API endpoints

### Error Codes

TODO: table of error codes
