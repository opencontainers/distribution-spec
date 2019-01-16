{
  // Initialize new info object.
  new(
    title=error 'Title not defined for Info object.',
    version='v1.0.0',
    description=null,
    termsOfService=null,
    contact=null,
    license=null
  ):: {
    title: title,
    version: version,
    [if description != null then 'description']: description,
    [if termsOfService != null then 'termsOfService']: termsOfService,
    [if contact != null then 'contact']: contact,
    [if license != null then 'license']: license,
  },
}
