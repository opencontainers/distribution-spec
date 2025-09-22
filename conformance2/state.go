package main

import (
	"fmt"
	"strings"
)

type state struct {
	APIStatus  map[stateAPIType]status
	Data       map[string]*testData
	DataStatus map[string]status
}

func stateNew() *state {
	return &state{
		APIStatus:  map[stateAPIType]status{},
		Data:       map[string]*testData{},
		DataStatus: map[string]status{},
	}
}

type stateAPIType int

const (
	stateAPITagList  stateAPIType = iota
	stateAPIBlobPush              // any blob push API
	stateAPIBlobPostOnly
	stateAPIBlobPostPut
	stateAPIBlobPatchChunk
	stateAPIBlobPatchStream
	stateAPIBlobMountSource
	stateAPIBlobMountAnonymous
	stateAPIBlobGetFull
	stateAPIBlobGetRange
	stateAPIBlobDelete
	stateAPIManifestPutDigest
	stateAPIManifestPutTag
	stateAPIManifestPutSubject
	stateAPIManifestGetDigest
	stateAPIManifestGetTag
	stateAPIManifestDeleteTag
	stateAPIManifestDeleteDigest
	stateAPIReferrers
	stateAPIMax // number of APIs for iterating
)

func (a stateAPIType) String() string {
	switch a {
	default:
		return "Unknown"
	case stateAPITagList:
		return "Tag listing"
	case stateAPIBlobPush:
		return "Blob push"
	case stateAPIBlobPostOnly:
		return "Blob post only"
	case stateAPIBlobPostPut:
		return "Blob post put"
	case stateAPIBlobPatchChunk:
		return "Blob chunked"
	case stateAPIBlobPatchStream:
		return "Blob streaming"
	case stateAPIBlobMountSource:
		return "Blob mount"
	case stateAPIBlobMountAnonymous:
		return "Blob anonymous mount"
	case stateAPIBlobGetFull:
		return "Blob get"
	case stateAPIBlobGetRange:
		return "Blob get range"
	case stateAPIBlobDelete:
		return "Blob delete"
	case stateAPIManifestPutDigest:
		return "Manifest put by digest"
	case stateAPIManifestPutTag:
		return "Manifest put by tag"
	case stateAPIManifestPutSubject:
		return "Manifest put with subject"
	case stateAPIManifestGetDigest:
		return "Manifest get by digest"
	case stateAPIManifestGetTag:
		return "Manifest get by tag"
	case stateAPIManifestDeleteTag:
		return "Manifest delete by tag"
	case stateAPIManifestDeleteDigest:
		return "Manifest delete by digest"
	case stateAPIReferrers:
		return "Referrers"
	}
}

func (a stateAPIType) MarshalText() ([]byte, error) {
	ret := a.String()
	if ret == "unknown" {
		return []byte(ret), fmt.Errorf("unknown API %d", a)
	}
	return []byte(ret), nil
}

func (a *stateAPIType) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	default:
		return fmt.Errorf("unknown API %s", b)
	case "Tag listing":
		*a = stateAPITagList
	case "Blob push":
		*a = stateAPIBlobPush
	case "Blob post only":
		*a = stateAPIBlobPostOnly
	case "Blob post put":
		*a = stateAPIBlobPostPut
	case "Blob chunked":
		*a = stateAPIBlobPatchChunk
	case "Blob streaming":
		*a = stateAPIBlobPatchStream
	case "Blob mount":
		*a = stateAPIBlobMountSource
	case "Blob anonymous mount":
		*a = stateAPIBlobMountAnonymous
	case "Blob get":
		*a = stateAPIBlobGetFull
	case "Blob get range":
		*a = stateAPIBlobGetRange
	case "Blob delete":
		*a = stateAPIBlobDelete
	case "Manifest put by digest":
		*a = stateAPIManifestPutDigest
	case "Manifest put by tag":
		*a = stateAPIManifestPutTag
	case "Manifest put with subject":
		*a = stateAPIManifestPutSubject
	case "Manifest get by digest":
		*a = stateAPIManifestGetDigest
	case "Manifest get by tag":
		*a = stateAPIManifestGetTag
	case "Manifest delete by tag":
		*a = stateAPIManifestDeleteTag
	case "Manifest delete by digest":
		*a = stateAPIManifestDeleteDigest
	case "Referrers":
		*a = stateAPIReferrers
	}
	return nil
}
