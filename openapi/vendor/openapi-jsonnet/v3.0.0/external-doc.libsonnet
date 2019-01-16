{
  // Initialize new External Documentation object.
  new(
    description,
    url
  ):: {
    [if description != null then 'description']: description,
    [if url != null then 'url']: url,
  },
}
