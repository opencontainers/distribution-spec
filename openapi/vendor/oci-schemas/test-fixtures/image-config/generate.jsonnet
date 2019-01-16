local golden = {
  created: '2015-10-31T22:22:56.015925234Z',
  author: 'Alyssa P. Hacker <alyspdev@example.com>',
  architecture: 'amd64',
  os: 'linux',
  config: {
    User: '1:1',
    ExposedPorts: {
      '8080/tcp': {},
    },
    Env: [
      'PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin',
      'FOO=bar',
    ],
    Entrypoint: [
      '/bin/sh',
    ],
    Cmd: [
      '--foreground',
      '--config',
      '/etc/my-app.d/default.cfg',
    ],
    Volumes: {
      '/var/job-result-data': {},
      '/var/log/my-app-logs': {},
    },
    StopSignal: 'SIGKILL',
    WorkingDir: '/home/alice',
    Labels: {
      'com.example.project.git.url': 'https://example.com/project.git',
      'com.example.project.git.commit': '45a939b2999782a3f005621a8d0f29aa387e1d6b',
    },
  },
  rootfs: {
    diff_ids: [
      'sha256:9d3dd9504c685a304985025df4ed0283e47ac9ffa9bd0326fddf4d59513f0827',
      'sha256:2b689805fbd00b2db1df73fae47562faac1a626d5f61744bfe29946ecff5d73d',
    ],
    type: 'layers',
  },
  history: [
    {
      created: '2015-10-31T22:22:54.690851953Z',
      created_by: '/bin/sh -c #(nop) ADD file:a3bc1e842b69636f9df5256c49c5374fb4eef1e281fe3f282c65fb853ee171c5 in /',
    },
    {
      created: '2015-10-31T22:22:55.613815829Z',
      created_by: '/bin/sh -c #(nop) CMD ["sh"]',
      empty_layer: true,
    },
  ],
};

// PASS: Required fields only.
local goldenMinimal = {
  architecture: 'amd64',
  os: 'linux',
  rootfs: {
    type: 'layers',
    diff_ids: [
      'sha256:9d3dd9504c685a304985025df4ed0283e47ac9ffa9bd0326fddf4d59513f0827',
    ],
  },
};

// FAIL: Env field invalid.
local envInvalid = {
  architecture: 'amd64',
  os: 'linux',
  config: {
    Env: ['invalid'],
  },
  rootfs: {
    type: 'layers',
    diff_ids: [
      'sha256:9d3dd9504c685a304985025df4ed0283e47ac9ffa9bd0326fddf4d59513f0827',
    ],
  },
};

{
  'golden.json': golden,
  'golden-minimal.json': goldenMinimal,
  'config-env-invalid.json': envInvalid,
}
