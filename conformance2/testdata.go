package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"math/big"
	"reflect"
	"strings"

	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	image "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	mtExampleConf  = "application/vnd.example.oci.conformance"
	mtOctetStream  = "application/octet-stream"
	mtOCIConfig    = "application/vnd.oci.image.config.v1+json"
	mtOCIImage     = "application/vnd.oci.image.manifest.v1+json"
	mtOCIIndex     = "application/vnd.oci.image.index.v1+json"
	mtOCILayer     = "application/vnd.oci.image.layer.v1.tar"
	mtOCILayerPre  = "application/vnd.oci.image.layer.v1."
	mtOCILayerGz   = "application/vnd.oci.image.layer.v1.tar+gzip"
	mtOCILayerNd   = "application/vnd.oci.image.layer.nondistributable.v1.tar"
	mtOCILayerNdGz = "application/vnd.oci.image.layer.nondistributable.v1.tar+gzip"
	mtOCIEmptyJSON = "application/vnd.oci.empty.v1+json"
)

type testData struct {
	name      string // name of data set for logs
	tags      map[string]digest.Digest
	desc      map[digest.Digest]*image.Descriptor
	blobs     map[digest.Digest][]byte
	manifests map[digest.Digest][]byte
	manOrder  []digest.Digest // ordered list to push manifests, the last is optionally tagged
	referrers map[digest.Digest][]*image.Descriptor
}

func newTestData(name string) *testData {
	return &testData{
		name:      name,
		tags:      map[string]digest.Digest{},
		desc:      map[digest.Digest]*image.Descriptor{},
		blobs:     map[digest.Digest][]byte{},
		manifests: map[digest.Digest][]byte{},
		manOrder:  []digest.Digest{},
		referrers: map[digest.Digest][]*image.Descriptor{},
	}
}

type genComp int

const (
	genCompUncomp genComp = iota
	genCompGzip
)

type genOptS struct {
	algo                digest.Algorithm
	annotations         map[string]string
	annotationUniq      bool
	artifactType        string
	blobSize            int64
	comp                genComp
	configBytes         []byte
	configMediaType     string
	descriptorMediaType string
	extraField          bool
	layerCount          int
	layerMediaType      string
	platform            image.Platform
	platforms           []*image.Platform
	setData             bool
	subject             *image.Descriptor
	tag                 string
}

type genOpt func(*genOptS)

func genWithAlgo(algo digest.Algorithm) genOpt {
	return func(opt *genOptS) {
		opt.algo = algo
	}
}

func genWithAnnotations(annotations map[string]string) genOpt {
	return func(opt *genOptS) {
		if opt.annotations == nil {
			opt.annotations = annotations
		} else {
			for k, v := range annotations {
				opt.annotations[k] = v
			}
		}
	}
}

func genWithAnnotationUniq() genOpt {
	return func(opt *genOptS) {
		opt.annotationUniq = true
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

func genWithConfigBytes(b []byte) genOpt {
	return func(opt *genOptS) {
		opt.configBytes = b
	}
}

func genWithConfigMediaType(mediaType string) genOpt {
	return func(opt *genOptS) {
		opt.configMediaType = mediaType
	}
}

func genWithDescriptorData() genOpt {
	return func(opt *genOptS) {
		opt.setData = true
	}
}

func genWithDescriptorMediaType(mediaType string) genOpt {
	return func(opt *genOptS) {
		opt.descriptorMediaType = mediaType
	}
}

func genWithExtraField() genOpt {
	return func(opt *genOptS) {
		opt.extraField = true
	}
}
func genWithLayerCount(count int) genOpt {
	return func(opt *genOptS) {
		opt.layerCount = count
	}
}

func genWithLayerMediaType(mediaType string) genOpt {
	return func(opt *genOptS) {
		opt.layerMediaType = mediaType
	}
}

func genWithPlatform(p image.Platform) genOpt {
	return func(opt *genOptS) {
		opt.platform = p
	}
}

func genWithPlatforms(platforms []*image.Platform) genOpt {
	return func(opt *genOptS) {
		opt.platforms = platforms
	}
}

func genWithBlobSize(size int64) genOpt {
	return func(opt *genOptS) {
		opt.blobSize = size
	}
}

func genWithSubject(subject image.Descriptor) genOpt {
	return func(opt *genOptS) {
		opt.subject = &subject
	}
}

func genWithTag(tag string) genOpt {
	return func(opt *genOptS) {
		opt.tag = tag
	}
}

func (td *testData) addBlob(b []byte, opts ...genOpt) (digest.Digest, error) {
	gOpt := genOptS{
		algo:                digest.Canonical,
		descriptorMediaType: mtOctetStream,
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	dig := gOpt.algo.FromBytes(b)
	td.blobs[dig] = b
	td.desc[dig] = &image.Descriptor{
		MediaType: gOpt.descriptorMediaType,
		Digest:    dig,
		Size:      int64(len(b)),
	}
	if gOpt.setData {
		td.desc[dig].Data = b
	}
	return dig, nil
}

func (td *testData) genBlob(opts ...genOpt) (digest.Digest, []byte, error) {
	gOpt := genOptS{
		blobSize: 2048,
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	b := make([]byte, gOpt.blobSize)
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
		mt = mtOCILayerGz
	case genCompUncomp:
		wUncomp = bufComp
		mt = mtOCILayer
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
	td.desc[digComp] = &image.Descriptor{
		MediaType: mt,
		Digest:    digComp,
		Size:      int64(len(bodyComp)),
	}
	if gOpt.setData {
		td.desc[digComp].Data = bodyComp
	}
	td.desc[digUncomp] = &image.Descriptor{
		MediaType: mtOCILayer,
		Digest:    digUncomp,
		Size:      int64(len(bodyUncomp)),
	}
	if gOpt.setData {
		td.desc[digUncomp].Data = bodyUncomp
	}
	return digComp, digUncomp, bodyComp, nil
}

// genConfig returns a config for the given platform and list of uncompressed layer digests.
func (td *testData) genConfig(p image.Platform, layers []digest.Digest, opts ...genOpt) (digest.Digest, []byte, error) {
	gOpt := genOptS{
		algo:            digest.Canonical,
		configMediaType: mtOCIConfig,
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	config := image.Image{
		Author:   "OCI Conformance Test",
		Platform: p,
		RootFS: image.RootFS{
			Type:    "layers",
			DiffIDs: layers,
		},
	}
	var body []byte
	var err error
	if !gOpt.extraField {
		body, err = json.Marshal(config)
	} else {
		body, err = json.Marshal(genAddJSONFields(config))
	}
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := gOpt.algo.FromBytes(body)
	td.blobs[dig] = body
	td.desc[dig] = &image.Descriptor{
		MediaType: gOpt.configMediaType,
		Digest:    dig,
		Size:      int64(len(body)),
	}
	if gOpt.setData {
		td.desc[dig].Data = body
	}
	return dig, body, nil
}

// genManifest returns an image manifest with the selected config and compressed layer digests.
func (td *testData) genManifest(conf image.Descriptor, layers []image.Descriptor, opts ...genOpt) (digest.Digest, []byte, error) {
	gOpt := genOptS{
		algo: digest.Canonical,
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	mt := mtOCIImage
	m := image.Manifest{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    mt,
		ArtifactType: gOpt.artifactType,
		Config:       conf,
		Layers:       layers,
		Subject:      gOpt.subject,
		Annotations:  gOpt.annotations,
	}
	if gOpt.annotationUniq {
		if m.Annotations == nil {
			m.Annotations = map[string]string{}
		} else {
			m.Annotations = maps.Clone(m.Annotations)
		}
		m.Annotations["org.example."+rand.Text()] = rand.Text()
	}
	var body []byte
	var err error
	if !gOpt.extraField {
		body, err = json.Marshal(m)
	} else {
		body, err = json.Marshal(genAddJSONFields(m))
	}
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := gOpt.algo.FromBytes(body)
	td.manifests[dig] = body
	td.manOrder = append(td.manOrder, dig)
	td.desc[dig] = &image.Descriptor{
		MediaType: m.MediaType,
		Digest:    dig,
		Size:      int64(len(body)),
	}
	if gOpt.setData {
		td.desc[dig].Data = body
	}
	at := m.ArtifactType
	if at == "" {
		at = m.Config.MediaType
	}
	if gOpt.subject != nil {
		td.referrers[gOpt.subject.Digest] = append(td.referrers[gOpt.subject.Digest], &image.Descriptor{
			MediaType:    m.MediaType,
			ArtifactType: at,
			Digest:       dig,
			Size:         int64(len(body)),
			Annotations:  m.Annotations,
		})
	}
	if gOpt.tag != "" {
		td.tags[gOpt.tag] = dig
	}
	return dig, body, nil
}

// genManifestFull creates an image with layers and a config
func (td *testData) genManifestFull(opts ...genOpt) (digest.Digest, error) {
	gOpt := genOptS{
		layerCount: 2,
		platform:   image.Platform{OS: "linux", Architecture: "amd64"},
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	digCList := []digest.Digest{}
	digUCList := []digest.Digest{}
	for l := range gOpt.layerCount {
		if gOpt.layerMediaType == "" || strings.HasPrefix(gOpt.layerMediaType, mtOCILayerPre) {
			// image
			digC, digUC, _, err := td.genLayer(l, opts...)
			if err != nil {
				return "", fmt.Errorf("failed to generate test data layer %d: %w", l, err)
			}
			digCList = append(digCList, digC)
			digUCList = append(digUCList, digUC)
		} else {
			// artifact
			lOpts := []genOpt{
				genWithDescriptorMediaType(gOpt.layerMediaType),
			}
			lOpts = append(lOpts, opts...)
			dig, _, err := td.genBlob(lOpts...)
			if err != nil {
				return "", fmt.Errorf("failed to generate test artifact blob: %w", err)
			}
			digCList = append(digCList, dig)
			digUCList = append(digUCList, dig)
		}
	}
	cDig := digest.Digest("")
	if gOpt.configMediaType == "" || gOpt.configMediaType == mtOCIConfig {
		// image config
		dig, _, err := td.genConfig(gOpt.platform, digUCList, opts...)
		if err != nil {
			return "", fmt.Errorf("failed to generate test data: %w", err)
		}
		cDig = dig
	} else {
		// artifact
		bOpts := []genOpt{
			genWithDescriptorMediaType(gOpt.configMediaType),
		}
		bOpts = append(bOpts, opts...)
		dig, err := td.addBlob(gOpt.configBytes, bOpts...)
		if err != nil {
			return "", fmt.Errorf("failed to generate test artifact config: %w", err)
		}
		cDig = dig
	}
	layers := make([]image.Descriptor, len(digCList))
	for i, lDig := range digCList {
		layers[i] = *td.desc[lDig]
	}
	mDig, _, err := td.genManifest(*td.desc[cDig], layers, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to generate test data: %w", err)
	}
	return mDig, nil
}

// genIndex returns an index manifest with the specified layers and platforms.
func (td *testData) genIndex(platforms []*image.Platform, manifests []digest.Digest, opts ...genOpt) (digest.Digest, []byte, error) {
	mt := mtOCIIndex
	gOpt := genOptS{
		algo: digest.Canonical,
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	if len(platforms) != len(manifests) {
		return digest.Digest(""), nil, fmt.Errorf("genIndex requires the same number of platforms and layers")
	}
	ind := image.Index{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    mt,
		ArtifactType: gOpt.artifactType,
		Manifests:    make([]image.Descriptor, len(manifests)),
		Subject:      gOpt.subject,
		Annotations:  gOpt.annotations,
	}
	for i, l := range manifests {
		d := *td.desc[l]
		d.Platform = platforms[i]
		ind.Manifests[i] = d
	}
	if gOpt.annotationUniq {
		if ind.Annotations == nil {
			ind.Annotations = map[string]string{}
		} else {
			ind.Annotations = maps.Clone(ind.Annotations)
		}
		ind.Annotations["org.example."+rand.Text()] = rand.Text()
	}
	var body []byte
	var err error
	if !gOpt.extraField {
		body, err = json.Marshal(ind)
	} else {
		body, err = json.Marshal(genAddJSONFields(ind))
	}
	if err != nil {
		return digest.Digest(""), nil, err
	}
	dig := gOpt.algo.FromBytes(body)
	td.manifests[dig] = body
	td.manOrder = append(td.manOrder, dig)
	td.desc[dig] = &image.Descriptor{
		MediaType: ind.MediaType,
		Digest:    dig,
		Size:      int64(len(body)),
	}
	if gOpt.setData {
		td.desc[dig].Data = body
	}
	if gOpt.subject != nil {
		td.referrers[gOpt.subject.Digest] = append(td.referrers[gOpt.subject.Digest], &image.Descriptor{
			MediaType:    ind.MediaType,
			ArtifactType: ind.ArtifactType,
			Digest:       dig,
			Size:         int64(len(body)),
			Annotations:  ind.Annotations,
		})
	}
	if gOpt.tag != "" {
		td.tags[gOpt.tag] = dig
	}
	return dig, body, nil
}

// genIndexFull creates an index with multiple images, including the image layers and configs
func (td *testData) genIndexFull(opts ...genOpt) (digest.Digest, error) {
	gOpt := genOptS{
		platforms: []*image.Platform{
			{OS: "linux", Architecture: "amd64"},
			{OS: "linux", Architecture: "arm64"},
		},
	}
	for _, opt := range opts {
		opt(&gOpt)
	}
	digImgList := []digest.Digest{}
	for _, p := range gOpt.platforms {
		iOpts := []genOpt{
			genWithPlatform(*p),
		}
		iOpts = append(iOpts, opts...)
		mDig, err := td.genManifestFull(iOpts...)
		if err != nil {
			return "", err
		}
		digImgList = append(digImgList, mDig)
	}
	iDig, _, err := td.genIndex(gOpt.platforms, digImgList, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to generate test data: %w", err)
	}
	return iDig, nil
}

func genAddJSONFields(v any) any {
	newT := reflect.StructOf([]reflect.StructField{
		{
			Name:      "Embed",
			Anonymous: true,
			Type:      reflect.TypeOf(v),
		},
		{
			Name: "Custom",
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag("json:\"org." + rand.Text() + "\""),
		},
	})
	newV := reflect.New(newT).Elem()
	newV.Field(0).Set(reflect.ValueOf(v))
	newV.FieldByName("Custom").SetString(rand.Text())
	return newV.Interface()
}
