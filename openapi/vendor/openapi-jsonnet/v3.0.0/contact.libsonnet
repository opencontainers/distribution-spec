{
  // Initalize new Contact object.
  new(
    name=null,
    url=null,
    email=null,
  ):: {
    [if name != null then 'name']: name,
    [if url != null then 'url']: url,
    [if email != null then 'email']: email,
  },
}
