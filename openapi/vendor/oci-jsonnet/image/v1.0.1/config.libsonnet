local errConfig(param) =
  error std.format('%s not defined for Config object.', param);

{
  new(
    os='linux',
    architecture='amd64',
    rootfsType='layers',
    diffIDs=[],
    created=null,
    author=null,
    config=null,
    history=[],
    user=null,
    exposedPorts=null,
    env=null,
    entryPoint=null,
    cmd=null,
    volumes=null,
    workingDir=null,
    labels=null,
    stopSignal=null,
  ):: {
    os: os,
    architecture: architecture,
    [if created != null then 'created']: created,
    [if author != null then 'author']: author,
    [if history != null then 'history']: history,
    config: {
      [if user != null then 'User']: user,
      [if exposedPorts != null then 'ExposedPorts']: exposedPorts,
      [if env != null then 'Env']: env,
      [if entryPoint != null then 'Entrypoint']: entryPoint,
      [if cmd != null then 'Cmd']: cmd,
      [if volumes != null then 'Volumes']: volumes,
      [if workingDir != null then 'WorkingDir']: workingDir,
      [if labels != null then 'Lables']: labels,
      [if stopSignal != null then 'StopSignal']: stopSignal,
    },
    rootfs: {
      type: rootfsType,
      diff_ids: diffIDs,
    },

    addDiffID(digest):: self {
      rootfs+: {
        diff_ids+: [digest],
      },
    },

    addExposedPort(port='8080', protocol='tcp'):: self {
      config+: {
        ExposedPorts+: {
          [std.format('%s/%s', [port, protocol])]: {},
        },
      },

    },

    addEnv(env):: self {
      config+: {
        Env+: [env],
      },
    },

    addVolume(path):: self {
      config+: {
        Volumes+: {
          [path]: {},
        },
      },
    },

    addLabel(key, value):: self {
      config+: {
        Labels+: {
          [key]: value,
        },
      },
    },

    addEntryPoint(entryPoint):: self {
      config+: {
        EntryPoint+: std.split(entryPoint, ' '),
      },
    },

    addCmd(cmd):: self {
      config+: {
        Cmd+: std.split(cmd, ' '),
      },
    },

    addHistory(
      created=null,
      author=null,
      createdBy=null,
      comment=null,
      emptyLayer=false,
    ):: self {
      history+: [{
        [if created != null then 'created']: created,
        [if author != null then 'author']: author,
        [if createdBy != null then 'created_by']: createdBy,
        [if comment != null then 'comment']: comment,
        empty_layer: emptyLayer,
      }],
    },
  },
}
