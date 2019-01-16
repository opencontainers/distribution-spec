{
  // Initialize new Operation object.
  new(
    operationId=null,
    tags=null,
    summary=null,
    description=null,
    externalDocs=null,
    parameters=null,
    requestBody=null,
    servers=null,
    deprecated=false,
  ):: {
    [if tags != null then 'tags']: tags,
    [if summary != null then 'summary']: summary,
    [if description != null then 'description']: description,
    [if externalDocs != null then 'externalDocs']: externalDocs,
    [if operationId != null then 'operationId']: operationId,
    [if parameters != null then 'parameters']: parameters,
    [if requestBody != null then 'requestBody']: requestBody,
    deprecated: deprecated,

    addParameter(parameter):: self {
      parameters+: [parameter],
    },

    addResponse(response):: self {
      responses+: response,
    },

    addRequestBody(requestBody):: self {
      requestBody+: requestBody,
    },

    addTag(tag):: self {
      tags+: [tag],
    },

    addServer(server):: self {
      servers+: [server],
    },

  },
}
