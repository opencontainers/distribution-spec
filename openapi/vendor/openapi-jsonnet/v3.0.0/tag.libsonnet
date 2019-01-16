{
  // Initialize new Tag object.
  new(
    name=error 'Name not defined for Tag object',
    description=null,
    externalDocs=null
  ):: {
    name: name,
    [if description != null then 'description']: description,
    [if externalDocs != null then 'externalDocs']: externalDocs,
  },
}
