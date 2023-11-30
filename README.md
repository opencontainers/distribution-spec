# OCI Distribution Specification

[![GitHub Actions status](https://github.com/opencontainers/distribution-spec/workflows/build/badge.svg)](https://github.com/opencontainers/distribution-spec/actions?query=workflow%3Abuild)

The OCI Distribution Spec project defines an API protocol to facilitate and standardize the distribution of content.

**[The specification can be found here](spec.md).**

This repository also provides [Go types](specs-go), and [registry conformance tooling](conformance).
The Go types and validation should be compatible with the current Go release; earlier Go releases are not supported.

Additional documentation about how this group operates:

- [Contributing](CONTRIBUTING.md)
- [Governance](GOVERNANCE.md)
- [Maintainers' Guide](MAINTAINERS_GUIDE.md)
- [Releases](RELEASES.md)
- [OCI Values](https://github.com/opencontainers/tob/blob/main/CHARTER.md#7-oci-values)

## Distributing OCI Images and other content

The OCI Distribution Spec is closely related to the [OCI Image Format Specification] project and the [OCI Runtime Specification] project.

The [OCI Image Format Specification] strictly defines the requirements for an OCI Image (container image), which consists of
a manifest, an optional image index, a set of filesystem layers, and a configuration.
The schema for OCI Image components is fully supported by the APIs defined in the OCI Distribution Specification.

The [OCI Runtime Specification] defines how to properly run a container "[filesystem bundle](https://github.com/opencontainers/runtime-spec/blob/master/bundle.md)"
which fully adheres to the OCI Image Format Specification. The OCI Runtime Specification is relevant to the OCI Distribution Specification in that they both support OCI Images,
and that container runtimes use the APIs defined in the OCI Distribution Specification to fetch pre-built container images and run them.

The [OCI Distribution Specification] (this project) is also designed generically enough to be leveraged as a distribution mechanism for
any type of content. The format of uploaded manifests, for example, need not necessarily adhere to the OCI Image Format Specification
so long as it references the blobs which comprise a given artifact.

[OCI Image Format Specification]: https://github.com/opencontainers/image-spec
[OCI Runtime Specification]: https://github.com/opencontainers/runtime-spec
[OCI Distribution Specification]: https://github.com/opencontainers/distribution-spec

## FAQ

For questions about the OCI Distribution Specification, please see the [FAQ](FAQ.md).

For general questions about OCI, please see the [FAQ on the OCI site](https://www.opencontainers.org/faq).

## Roadmap

The [GitHub milestones](https://github.com/opencontainers/distribution-spec/milestones) lay out the path to the future improvements.

# Extensions

The Distribution Specification Project includes a process and API for prototyping
and testing extensions to the Distribution API.

We invite contributions, comments, and reviews to these extensions.
These extensions will only move forward with significant support from registries, registry clients, and users.

Please see [here](https://github.com/opencontainers/distribution-spec/blob/main/extensions/README.md) for more details.

# Contributing

Development happens on GitHub for the spec.
Issues are used for bugs and actionable items and longer discussions can happen on the [mailing list](#mailing-list).

The specification and code is licensed under the Apache 2.0 license found in the `LICENSE` file of this repository.

## Discuss your design

The project welcomes submissions, but please let everyone know what you are working on.

Before undertaking a nontrivial change to this specification, send mail to the [mailing list](#mailing-list) to discuss what you plan to do.
This gives everyone a chance to validate the design, helps prevent duplication of effort, and ensures that the idea fits.
It also guarantees that the design is sound before code is written; a GitHub pull-request is not the place for high-level discussions.

Typos and grammatical errors can go straight to a pull-request.
When in doubt, start on the [mailing-list](#mailing-list).

## Meetings

Please see the [OCI org repository README](https://github.com/opencontainers/org#meetings) for the most up-to-date information on OCI contributor and maintainer meeting schedules.
You can also find links to meeting agendas and minutes for all prior meetings.

## Mailing List

You can subscribe and join the mailing list on [Google Groups](https://groups.google.com/a/opencontainers.org/forum/#!forum/dev).

## Chat

OCI discussion happens in the following chat rooms, which are all bridged together:

- #general channel on [OCI Slack](https://opencontainers.org/community/overview/#chat)
- #opencontainers:matrix.org

## Markdown style

To keep consistency throughout the Markdown files in the Open Container spec all files should be formatted one sentence per line.
This fixes two things: it makes diffing easier with git and it resolves fights about line wrapping length.
For example, this paragraph will span three lines in the Markdown source.

## Git commit

### Sign your work

The sign-off is a simple line at the end of the explanation for the patch, which certifies that you wrote it or otherwise have the right to pass it on as an open-source patch.
The rules are pretty simple: if you can certify the below (from [developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

then you just add a line to every git commit message:

    Signed-off-by: Jane Smith <jane@gmail.com>

using your real name (sorry, no pseudonyms or anonymous contributions.)

You can add the sign off when creating the git commit via `git commit -s`.

### Commit Style

Simple house-keeping for clean git history.
Read more on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/) or the Discussion section of [`git-commit(1)`](http://git-scm.com/docs/git-commit).

1. Separate the subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how
* If there was important/useful/essential conversation or information, copy or include a reference
8. When possible, one keyword to scope the change in the subject (i.e. "README: ...", "runtime: ...")
