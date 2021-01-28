## Content Negotiation

TODO - Please see
[issue #212](https://github.com/opencontainers/distribution-spec/issues/212).

| `Accept`                                                                    | Multi?             | Result (type)           | Notes                                                                                        |
|:----------------------------------------------------------------------------|:-------------------|:------------------------|:---------------------------------------------------------------------------------------------|
| `<not present>`                                                             | :white_check_mark: | `manifest.v1+prettyjws` | `v1`, `linux/amd64`, for backward compatibility with old, non-v2 clients                     |
| `*/*`                                                                       | :white_check_mark: | `manifest.v1+prettyjws` | `v1`, `linux/amd64`, for backward compatibility with old, non-v2 clients (same as above)     |
| `application/json`                                                          | :white_check_mark: | `manifest.v1+prettyjws` | `v1`, `linux/amd64`, for backward compatibility with old, non-v2 clients (same as above)     |
| `manifest.v1+json`                                                          | :white_check_mark: | `manifest.v1+prettyjws` | `v1`, matching `Accept` header (`linux/amd64` for backward compatibility)                    |
| `manifest.v2+json`                                                          | :white_check_mark: | `manifest.v2+json`      | `v2`, matching `Accept` header (`linux/amd64` for backward compatibility)                    |
| `manifest.list.v2+json`                                                     | :white_check_mark: | `manifest.list.v2+json` | `v2 manifest list`, matching `Accept` header                                                 |
| `manifest.list.v2+json`,<br /> `manifest.v2+json`,<br /> `manifest.v1+json` | :white_check_mark: | `manifest.list.v2+json` | `v2 manifest list`, matching first `Accept` header                                           |
| `manifest.v2+json`,<br /> `manifest.list.v2+json`,<br /> `manifest.v1+json` | :white_check_mark: | `manifest.v2+json`      | :warning: prefers manifest list over manifest (ignoring order?)                              |
| `<not present>`                                                             | -                  | `manifest.v1+prettyjws` | same as multi-manifest repo                                                                  |
| `*/*`                                                                       | -                  | `manifest.v1+prettyjws` | same as multi-manifest repo                                                                  |
| `application/json`                                                          | -                  | `manifest.v1+prettyjws` | same as multi-manifest repo                                                                  |
| `manifest.v1+json`                                                          | -                  | `manifest.v1+prettyjws` | same as multi-manifest repo                                                                  |
| `manifest.v2+json`                                                          | -                  | `manifest.v2+json`      | same as multi-manifest repo                                                                  |
| `manifest.list.v2+json`                                                     | -                  | `manifest.list.v2+json` | :warning: somewhat unexpected to return a v1 manifest for a v2-capable client                |
| `manifest.list.v2+json`,<br /> `manifest.v2+json`,<br /> `manifest.v1+json` | -                  | `manifest.list.v2+json` | `v2 manifest`, matching first "acceptable" `Accept` header (by lack of a manifest-list)      |
| `manifest.v2+json`,<br /> `manifest.list.v2+json`,<br /> `manifest.v1+json` | -                  | `manifest.v2+json`      | `v2 manifest`, matching first "acceptable" `Accept` header (by lack of a manifest-list)      |
