{
  // Initialize new Paths object
  new():: {

    addPath(
      path=error 'Path not defined for Paths object.',
      pathItem=error 'Path Item not defined for Paths object.',
    ):: self {
      [path]: pathItem,
    },

  },

}
