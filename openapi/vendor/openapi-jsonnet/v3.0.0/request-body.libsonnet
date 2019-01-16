{
  // Initialize new Request Body object.
  new(
    content={},
    required=true,
    description=null,
  ):: {
    content: content,
    required: required,
    [if description != null then 'description']: description,

    addContent(content):: self {
        content+: content,
    },
  },
}
