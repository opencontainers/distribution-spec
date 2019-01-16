{
  // Initalize new Responses object.
  new():: {

    addResponse(statusCode, response):: self {
      [statusCode]: response,
    },

  },
}
