package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"

	digest "github.com/opencontainers/go-digest"
)

type testData struct {
	name      string // name of data set for logs
	repo      string // specifies the remote repository when the data has been pushed
	tag       string // specifies the tag used for last manifest push
	manifests map[digest.Digest][]byte
	blobs     map[digest.Digest][]byte
	manOrder  []digest.Digest // ordered list to push manifests, the last is optionally tagged
}

func newTestData(name, tag string) *testData {
	return &testData{
		name:      name,
		manifests: map[digest.Digest][]byte{},
		blobs:     map[digest.Digest][]byte{},
		manOrder:  []digest.Digest{},
		tag:       tag,
	}
}

func (td *testData) genBlob(algo digest.Algorithm, size int64) (digest.Digest, []byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := algo.FromBytes(b)
	td.blobs[dig] = b
	return dig, b, nil
}

// genLayer returns a new layer with:
// - compressed digest
// - uncompressed digest
// - layer body (tar+gzip compressed)
func (td *testData) genLayer(fileNum int) (digest.Digest, digest.Digest, []byte, error) {
	buf := &bytes.Buffer{}
	digUncomp := digest.Canonical.Digester()
	digComp := digest.Canonical.Digester()
	mwComp := io.MultiWriter(buf, digComp.Hash())
	gw := gzip.NewWriter(mwComp)
	mwUncomp := io.MultiWriter(gw, digUncomp.Hash())
	tw := tar.NewWriter(mwUncomp)
	bigRandNum, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	randNum := bigRandNum.Int64()
	file := fmt.Sprintf("Conformance test file contents for file number %d.\nTodays lucky number is %d\n", fileNum, randNum)
	err = tw.WriteHeader(&tar.Header{
		Name: fmt.Sprintf("./conformance-%d.txt", fileNum),
		Size: int64(len(file)),
		Mode: 0644,
	})
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	_, err = tw.Write([]byte(file))
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	err = tw.Close()
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	err = gw.Close()
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	body := buf.Bytes()
	td.blobs[digComp.Digest()] = body
	return digComp.Digest(), digUncomp.Digest(), body, nil
}

// genConfig returns a config for the given platform and list of uncompressed layer digests.
func (td *testData) genConfig(p platform, layers []digest.Digest) (digest.Digest, []byte, error) {
	config := image{
		Author:   "OCI Conformance Test",
		platform: p,
		RootFS: rootFS{
			Type:    "layers",
			DiffIDs: layers,
		},
	}
	body, err := json.Marshal(config)
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := digest.Canonical.FromBytes(body)
	td.blobs[dig] = body
	return dig, body, nil
}

// genManifest returns an image manifest with the selected config and compressed layer digests.
func (td *testData) genManifest(conf digest.Digest, layers []digest.Digest) (digest.Digest, []byte, error) {
	m := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    conf,
			Size:      int64(len(td.blobs[conf])),
		},
		Layers: make([]descriptor, len(layers)),
	}
	for i, l := range layers {
		m.Layers[i] = descriptor{
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Digest:    l,
			Size:      int64(len(td.blobs[l])),
		}
	}
	body, err := json.Marshal(m)
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := digest.Canonical.FromBytes(body)
	td.manifests[dig] = body
	td.manOrder = append(td.manOrder, dig)
	return dig, body, nil
}

// genIndex returns an index manifest with the specified layers and platforms.
func (td *testData) genIndex(platforms []*platform, manifests []digest.Digest) (digest.Digest, []byte, error) {
	if len(platforms) != len(manifests) {
		return digest.Digest(""), nil, fmt.Errorf("genIndex requires the same number of platforms and layers")
	}
	ind := index{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests:     make([]descriptor, len(manifests)),
	}
	for i, l := range manifests {
		ind.Manifests[i] = descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Digest:    l,
			Size:      int64(len(td.manifests[l])),
			Platform:  platforms[i],
		}
	}
	body, err := json.Marshal(ind)
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := digest.Canonical.FromBytes(body)
	td.manifests[dig] = body
	td.manOrder = append(td.manOrder, dig)
	return dig, body, nil
}
