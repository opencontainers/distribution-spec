{
  // Initialize new Encoding object.
  new(
    contentType=null,
    style=null,
    explode=null,
    allowReserved=null
  ):: {
    [if contentType != null then 'contentType']: contentType,
    [if style != null then 'style']: style,
    [if explode != null then 'explode']: explode,
    [if allowReserved != null then 'allowReserved']: allowReserved,

    addHeader(header):: self {
      headers+: header,
    },
  },
}
