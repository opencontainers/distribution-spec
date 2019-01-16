{
  // Initialize new License object.
  new(
    name=error 'Name not defined for License object',
    url=null,
  ):: {
    name: name,
    [if url != null then 'url']: url,
  },
}
