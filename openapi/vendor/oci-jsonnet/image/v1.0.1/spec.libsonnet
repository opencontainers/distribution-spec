local config = import 'config.libsonnet';
local descriptor = import 'descriptor.libsonnet';
local examples = import 'examples.libsonnet';
local index = import 'index.libsonnet';
local manifest = import 'manifest.libsonnet';
local rootfs = import 'rootfs.libsonnet';

{
  config:: config.new,
  rootfs:: rootfs.new,
  index:: index.new,
  manifest:: manifest.new,
  descriptor:: descriptor.new,
  examples:: examples,
}
