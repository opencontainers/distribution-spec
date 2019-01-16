{
  // Initialize Example object.
  new(
    summary=null,
    description=null,
    value=null,
    externalValue=null,
  ):: {
    [if summary != null then 'summary']: summary,
    [if description != null then 'description']: description,
    [if value != null then 'value']: value,
    [if externalValue != null then 'externalValue']: externalValue,
  },
}
