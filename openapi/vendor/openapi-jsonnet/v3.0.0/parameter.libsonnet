{
  // Initialize Parameter object.
  new(
    name=error 'Name not defined for Parameter object.',
    paramIn=error 'ParamIn not defined for Parameter object.',
    description=null,
    required=true,
    deprecated=false,
    allowEmptyValue=false,
    style='simple',
    explode=false,
    allowReserved=false,
    schema=null,
    example=null,
    examples=null,
    content=null,
  ):: {
    name: name,
    'in': paramIn,
    required: required,
    deprecated: deprecated,
    allowEmptyValue: allowEmptyValue,
    style: style,
    explode: explode,
    allowReserved: allowReserved,
    [if description != null then 'description']: description,
    [if schema != null then 'schema']: schema,
    [if example != null then 'example']: example,
    [if examples != null then 'examples']: examples,
    [if content != null then 'content']: content,
  },
}
