{
  // Initialize Link object.
  new(
    operationId=error 'OperationId not defined for Link object.'.,
    description=null,
    requestBody=null,
    server=null,
  ):: {
    operationId: operationId,
    [if description != null then 'description']: description,
    [if requestBody != null then 'requestBody']: requestBody,
    [if server != null then 'server']: server,

    addParameter(parameter):: self {
      parameters+: parameter,
    },
  },
}
