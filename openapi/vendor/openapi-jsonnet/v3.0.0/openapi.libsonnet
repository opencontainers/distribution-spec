{
  info: import 'info.libsonnet',
  contact: import 'contact.libsonnet',
  license: import 'license.libsonnet',
  server: import 'server.libsonnet',
  serverVariable: import 'server-variable.libsonnet',
  paths: import 'paths.libsonnet',
  pathItem: import 'path-item.libsonnet',
  response: import 'response.libsonnet',
  operation: import 'operation.libsonnet',
  parameter: import 'parameter.libsonnet',
  mediaType: import 'media-type.libsonnet',
  header: import 'header.libsonnet',
  requestBody: import 'request-body.libsonnet',
  example: import 'example.libsonnet',

  // Initalize new OpenAPI spec.
  new(
    info=error 'Info not defined for OpenAPI object.',
    paths=error 'Paths not defined for OpenAPI object.',
    servers=null,
    tags=null,
    externalDocs=null,
  ):: {
    openapi: '3.0.0',
    info: info,
    paths: paths,
    [if servers != null then 'servers']: servers,
    [if tags != null then 'tags']: tags,
    [if externalDocs != null then 'externalDocs']: externalDocs,
  },
}
