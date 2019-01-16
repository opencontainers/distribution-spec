local img = import 'spec.libsonnet';

local layers = {
  one:
    img.descriptor(
      mediaType='application/vnd.oci.image.layer.v1.tar+gzip',
      digest='sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a',
      size=64265,
    ),

  two:
    img.descriptor(
      mediaType='application/vnd.oci.image.layer.v1.tar+gzip',
      digest='sha256:1303f6fada513d589891cbc286fd3219fe06156f2951f1d3cee8663a4070fc53',
      size=255262,
    ),

  three:
    img.descriptor(
      mediaType='application/vnd.oci.image.layer.v1.tar+gzip',
      digest='sha256:c1a9aedd4f3a489d779c0711b13251fddc4dacdf79bf0042beebe64a72d9fb92',
      size=658924,
    ),

  four:
    img.descriptor(
      mediaType='application/vnd.oci.image.layer.v1.tar+gzip',
      digest='sha256:27e61142701b0cce5b00017691edd5645d7dea724db4c02390973beec51503c0',
      size=5342245,
    ),
};

local manifests = {
  one:
    img.descriptor(
      mediaType='application/vnd.oci.image.manifest.v1+json',
      digest='sha:1bdbe415c4ecb05e6f0934bfd3cbcf7e569102d59cb6b49247a01a4fc7b69a69',
      size=345,
    ),

  two:
    img.descriptor(
      mediaType='application/vnd.oci.image.manifest.v1+json',
      digest='sha:e6a79d462a04b6c22941aff9b5771a5a42e233a311454d5e313a2b9c73e2ea43',
      size=698,
    ),

  three:
    img.descriptor(
      mediaType='application/vnd.oci.image.manifest.v1+json',
      digest='sha:63f95acf3bcf356d58e267fdcaeb9350e5371fbd2b96635c5febde803c7d98ce',
      size=532,
    ),
};

local configs = {
  one:
    img.descriptor(
      mediaType='application/vnd.oci.image.config.v1+json',
      digest='sha256:28f27f8eec1e498f6cdbdb0df9f75a44da95c8ab99852a3d77ce80b9b3347c6b',
      size=542,
    ),
};

{

  index:
    img.index()
    .addManifest(
      descriptor=manifests.one,
      architecture='amd64',
      os='linux',
    )
    .addManifest(
      descriptor=manifests.two,
      architecture='arm',
      variant='v7',
      os='linux',
    )
    .addManifest(
      descriptor=manifests.three,
      architecture='amd64',
      os='windows',
      osVersion='10.0.143453.1066',
      osFeatures=['win32k'],
    )

    .addAnnotation('org.opencontainers.image.created', '2018-12-16 20:11:14.351488997-06:00')
    .addAnnotation('org.opencontainers.image.authors', 'Pepper Potts <pepper.potts@starkindustries.com>')
    .addAnnotation('org.opencontainers.image.url', 'https://app-site.com')
    .addAnnotation('org.opencontainers.image.documentation', 'https://app-site.com/docs')
    .addAnnotation('org.opencontainers.image.source', 'https://app-site.com/code/app.git')
    .addAnnotation('org.opencontainers.image.version', 'v1.4.5')
    .addAnnotation('org.opencontainers.image.vendor', 'App Corporation')
    .addAnnotation('org.opencontainers.image.licenses', 'Apache License 2.0')
    .addAnnotation('org.opencontainers.image.ref.name', 'App')
    .addAnnotation('org.opencontainers.image.title', 'App Thing Doer')
    .addAnnotation('org.opencontainers.image.description', 'App is an application that does things.')
  ,

  manifest:
    img.manifest(
      config=configs.one,
    )
    .addLayer(layers.one)
    .addLayer(layers.two)
    .addLayer(layers.three)
    .addLayer(layers.four)

    .addAnnotation('org.opencontainers.image.created', '2018-12-16 20:11:14.351488997-06:00')
    .addAnnotation('org.opencontainers.image.authors', 'Pepper Potts <pepper.potts@starkindustries.com>')
    .addAnnotation('org.opencontainers.image.url', 'https://app-site.com')
    .addAnnotation('org.opencontainers.image.documentation', 'https://app-site.com/docs')
    .addAnnotation('org.opencontainers.image.source', 'https://app-site.com/code/app.git')
    .addAnnotation('org.opencontainers.image.version', 'v1.4.5')
    .addAnnotation('org.opencontainers.image.vendor', 'App Corporation')
    .addAnnotation('org.opencontainers.image.licenses', 'Apache License 2.0')
    .addAnnotation('org.opencontainers.image.ref.name', 'App')
    .addAnnotation('org.opencontainers.image.title', 'App Thing Doer')
    .addAnnotation('org.opencontainers.image.description', 'App is an application that does things.')
  ,

  config:
    img.config(
      created='2018-12-16T20:38:54.690851953Z',
      author='Pepper Potts',
      user='app-user',
      workingDir='/var/lib/app',
      stopSignal='SIGTERM',
    )
    .addDiffID('sha256:e3161859d1779d8330428ed745008710a1ecfb9f494c2e1b062be4cc0ba9ee2a')
    .addDiffID('sha256:44cc372097a48dd8e9fdb9a4cbbe4540f07a56706b772830623d95d0ba6f0622')
    .addDiffID('sha256:c103de080df4338c445c423dee488d5001c7e986e1857f3e7c4483cf160231b2')
    .addDiffID('sha256:6143c61b46168fc5b97d917ce275b6dc4bfa14e835387b9659fb8867e8eaeadd')

    .addHistory(
      created='2018-12-16 20:36:25.543469672-06:00',
      author='Pepper Potts <pepper.potts@starkindustries.com>',
      createdBy='/usr/local/bin/oci-compliant-build-tool ./build/app',
      comment='App base layer.',
    )
    .addHistory(
      created='2018-12-16 20:36:43.287882890-06:00',
      author='Pepper Potts <pepper.potts@starkindustries.com>',
      createdBy='/usr/local/bin/oci-compliant-build-tool ./build/app',
      comment='App dependency one.',
    )
    .addHistory(
      created='2018-12-16 20:37:17.990254281-06:00',
      author='Pepper Potts <pepper.potts@starkindustries.com>',
      createdBy='/usr/local/bin/oci-compliant-build-tool ./build/app',
      comment='App dependency two.',
    )
    .addHistory(
      created='2018-12-16 20:37:37.250792666-06:00',
      author='Pepper Potts <pepper.potts@starkindustries.com>',
      createdBy='/usr/local/bin/oci-compliant-build-tool ./build/app',
      comment='App overwrite default config.',
    )

    .addExposedPort('9000', 'tcp')

    .addEnv('PATH=/bin:/usr/bin:/usr/local/bin')
    .addEnv('APP_TLS_CERT=/etc/ssl/app/server.pem')
    .addEnv('APP_TLS_KEY=/etc/ssl/app/server-key.pem')

    .addVolume('/etc/app')
    .addVolume('/var/lib/app')
    .addVolume('/var/log/app')

    .addLabel('org.opencontainers.image.created', '2018-12-16 20:11:14.351488997-06:00')
    .addLabel('org.opencontainers.image.authors', 'Pepper Potts <pepper.potts@starkindustries.com>')
    .addLabel('org.opencontainers.image.url', 'https://app-site.com')
    .addLabel('org.opencontainers.image.documentation', 'https://app-site.com/docs')
    .addLabel('org.opencontainers.image.source', 'https://app-site.com/code/app.git')
    .addLabel('org.opencontainers.image.version', 'v1.4.5')
    .addLabel('org.opencontainers.image.vendor', 'App Corporation')
    .addLabel('org.opencontainers.image.licenses', 'Apache License 2.0')
    .addLabel('org.opencontainers.image.ref.name', 'App')
    .addLabel('org.opencontainers.image.title', 'App Thing Doer')
    .addLabel('org.opencontainers.image.description', 'App is an application that does things.')

    .addEntryPoint('/usr/local/bin/app')
    .addCmd('--config /etc/app/config.json --logdir /var/log/app/app.log'),


}
