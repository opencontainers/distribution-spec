{
  // Initialize new Server Variable object.
  new(
    name=error "New Server Variable method missing 'name' parameter.",
    default=error "Server Variable object missing required field 'default'.",
    description=null
  ):: {
    [name]: {
      default: default,
      [if description != null then 'description']: description,
    },

    addEnum(enum):: self {
      [name]+: {
        enum+: enum,
      },
    },
  },
}
