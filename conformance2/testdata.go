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
	tags      map[string]digest.Digest
	desc      map[digest.Digest]*descriptor
	blobs     map[digest.Digest][]byte
	manifests map[digest.Digest][]byte
	manOrder  []digest.Digest // ordered list to push manifests, the last is optionally tagged
	referrers map[digest.Digest][]digest.Digest
}

func newTestData(name string) *testData {
	return &testData{
		name:      name,
		tags:      map[string]digest.Digest{},
		desc:      map[digest.Digest]*descriptor{},
		blobs:     map[digest.Digest][]byte{},
		manifests: map[digest.Digest][]byte{},
		manOrder:  []digest.Digest{},
		referrers: map[digest.Digest][]digest.Digest{},
	}
}

type genComp int

const (
	genCompUncomp genComp = iota
	genCompGzip
)

type genOptS struct {
	algo         digest.Algorithm
	artifactType string
	comp         genComp
	mediaType    string
	setData      bool
	subject      *descriptor
}

type genOpt func(*genOptS)

func genWithAlgo(algo digest.Algorithm) genOpt {
	return func(opt *genOptS) {
		opt.algo = algo
	}
}

func genWithArtifactType(artifactType string) genOpt {
	return func(opt *genOptS) {
		opt.artifactType = artifactType
	}
}

func genWithCompress(comp genComp) genOpt {
	return func(opt *genOptS) {
		opt.comp = comp
	}
}

func genWithMediaType(mediaType string) genOpt {
	return func(opt *genOptS) {
		opt.mediaType = mediaType
	}
}

func genSetData() genOpt {
	return func(opt *genOptS) {
		opt.setData = true
	}
}

func genWithSubject(subject descriptor) genOpt {
	return func(opt *genOptS) {
		opt.subject = &subject
	}
}

func (td *testData) addBlob(b []byte, opts ...genOpt) (digest.Digest, error) {
	gOpt := genOptS{
		algo:      digest.Canonical,
		mediaType: "application/octet-stream",
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	dig := gOpt.algo.FromBytes(b)
	td.blobs[dig] = b
	td.desc[dig] = &descriptor{
		MediaType: gOpt.mediaType,
		Digest:    dig,
		Size:      int64(len(b)),
	}
	if gOpt.setData {
		td.desc[dig].Data = b
	}
	return dig, nil
}

func (td *testData) genBlob(size int64, opts ...genOpt) (digest.Digest, []byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig, err := td.addBlob(b, opts...)
	return dig, b, err
}

// genLayer returns a new layer containing a tar file returning:
// - compressed digest
// - uncompressed digest
// - layer body (tar+compression)
func (td *testData) genLayer(fileNum int, opts ...genOpt) (digest.Digest, digest.Digest, []byte, error) {
	gOpt := genOptS{
		comp: genCompGzip,
		algo: digest.Canonical,
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	bufUncomp := &bytes.Buffer{}
	bufComp := &bytes.Buffer{}
	var wUncomp io.Writer
	var mt string
	switch gOpt.comp {
	case genCompGzip:
		wUncomp = gzip.NewWriter(bufComp)
		mt = "application/vnd.oci.image.layer.v1.tar+gzip"
	case genCompUncomp:
		wUncomp = bufComp
		mt = "application/vnd.oci.image.layer.v1.tar"
	}
	wTar := tar.NewWriter(wUncomp)
	bigRandNum, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	randNum := bigRandNum.Int64()
	file := fmt.Sprintf("Conformance test file contents for file number %d.\nTodays lucky number is %d\n", fileNum, randNum)
	err = wTar.WriteHeader(&tar.Header{
		Name: fmt.Sprintf("./conformance-%d.txt", fileNum),
		Size: int64(len(file)),
		Mode: 0644,
	})
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	_, err = wTar.Write([]byte(file))
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	err = wTar.Close()
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	if closer, ok := wUncomp.(io.Closer); gOpt.comp != genCompUncomp && ok {
		err = closer.Close()
	}
	if err != nil {
		return digest.Digest(""), digest.Digest(""), nil, err
	}
	bodyComp := bufComp.Bytes()
	bodyUncomp := bufUncomp.Bytes()
	digComp := gOpt.algo.FromBytes(bodyComp)
	digUncomp := gOpt.algo.FromBytes(bodyUncomp)
	td.blobs[digComp] = bodyComp
	td.desc[digComp] = &descriptor{
		MediaType: mt,
		Digest:    digComp,
		Size:      int64(len(bodyComp)),
	}
	if gOpt.setData {
		td.desc[digComp].Data = bodyComp
	}
	td.desc[digUncomp] = &descriptor{
		MediaType: "application/vnd.oci.image.layer.v1.tar",
		Digest:    digUncomp,
		Size:      int64(len(bodyUncomp)),
	}
	if gOpt.setData {
		td.desc[digUncomp].Data = bodyUncomp
	}
	return digComp, digUncomp, bodyComp, nil
}

// genConfig returns a config for the given platform and list of uncompressed layer digests.
func (td *testData) genConfig(p platform, layers []digest.Digest, opts ...genOpt) (digest.Digest, []byte, error) {
	gOpt := genOptS{
		algo:      digest.Canonical,
		mediaType: "application/vnd.oci.image.config.v1+json",
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
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
	dig := gOpt.algo.FromBytes(body)
	td.blobs[dig] = body
	td.desc[dig] = &descriptor{
		MediaType: gOpt.mediaType,
		Digest:    dig,
		Size:      int64(len(body)),
	}
	if gOpt.setData {
		td.desc[dig].Data = body
	}
	return dig, body, nil
}

// genManifest returns an image manifest with the selected config and compressed layer digests.
func (td *testData) genManifest(conf digest.Digest, layers []digest.Digest, opts ...genOpt) (digest.Digest, []byte, error) {
	gOpt := genOptS{
		algo:      digest.Canonical,
		mediaType: "application/vnd.oci.image.manifest.v1+json",
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	m := manifest{
		SchemaVersion: 2,
		MediaType:     gOpt.mediaType,
		ArtifactType:  gOpt.artifactType,
		Config:        *td.desc[conf],
		Layers:        make([]descriptor, len(layers)),
		Subject:       gOpt.subject,
	}
	for i, l := range layers {
		m.Layers[i] = *td.desc[l]
	}
	body, err := json.Marshal(m)
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := gOpt.algo.FromBytes(body)
	td.manifests[dig] = body
	td.manOrder = append(td.manOrder, dig)
	td.desc[dig] = &descriptor{
		MediaType: gOpt.mediaType,
		Digest:    dig,
		Size:      int64(len(body)),
	}
	if gOpt.setData {
		td.desc[dig].Data = body
	}
	if gOpt.subject != nil {
		td.referrers[gOpt.subject.Digest] = append(td.referrers[gOpt.subject.Digest], dig)
	}
	return dig, body, nil
}

// genIndex returns an index manifest with the specified layers and platforms.
func (td *testData) genIndex(platforms []*platform, manifests []digest.Digest, opts ...genOpt) (digest.Digest, []byte, error) {
	gOpt := genOptS{
		algo:      digest.Canonical,
		mediaType: "application/vnd.oci.image.index.v1+json",
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	if len(platforms) != len(manifests) {
		return digest.Digest(""), nil, fmt.Errorf("genIndex requires the same number of platforms and layers")
	}
	ind := index{
		SchemaVersion: 2,
		MediaType:     gOpt.mediaType,
		ArtifactType:  gOpt.artifactType,
		Manifests:     make([]descriptor, len(manifests)),
		Subject:       gOpt.subject,
	}
	for i, l := range manifests {
		d := *td.desc[l]
		d.Platform = platforms[i]
		ind.Manifests[i] = d
	}
	body, err := json.Marshal(ind)
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := gOpt.algo.FromBytes(body)
	td.manifests[dig] = body
	td.manOrder = append(td.manOrder, dig)
	td.desc[dig] = &descriptor{
		MediaType: gOpt.mediaType,
		Digest:    dig,
		Size:      int64(len(body)),
	}
	if gOpt.setData {
		td.desc[dig].Data = body
	}
	return dig, body, nil
}
