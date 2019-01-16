{
  // Initialize new Response object.
  new(
    statusCode=error 'StatusCode not defined for Response object.',
    description=error 'Description not defined for Response object.',
    content=null,
  ):: {
    [statusCode]: {
      [if description != null then 'description']: description,
      [if content != null then 'content']: content,
    },

    addHeader(headers):: self {
      [statusCode]+: {
        headers+: headers,
      },
    },

    addContent(content):: self {
      [statusCode]+: {
        content+: content,
      },
    },
  },
}
