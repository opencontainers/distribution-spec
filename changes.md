# OCI Distribution Specification Changes

With the exception of de-minimis modifications (such as non-code impacting spelling corrections), this doc should be updated when changes are made to the specification, indicating what is different.

Each set of changes is given a letter corresponding to a set of modifications that were applied to the baseline specification.
These are merely for reference and shouldn't be used outside the specification other than to identify a set of modifications.

<dl>
  <dt>l</dt>
  <dd>
    <ul>
      <li>Document TOOMANYREQUESTS error code.</li>
    </ul>
  </dd>

  <dt>k</dt>
  <dd>
    <ul>
      <li>Document use of Accept and Content-Type headers in manifests endpoint.</li>
    </ul>
  </dd>

  <dt>j</dt>
  <dd>
    <ul>
      <li>Add ability to mount blobs across repositories.</li>
    </ul>
  </dd>

  <dt>i</dt>
  <dd>
    <ul>
      <li>Clarified expected behavior response to manifest HEAD request.</li>
    </ul>
  </dd>

  <dt>h</dt>
  <dd>
    <ul>
      <li>All mention of tarsum removed.</li>
    </ul>
  </dd>

  <dt>g</dt>
  <dd>
    <ul>
      <li>Clarify behavior of pagination behavior with unspecified parameters.</li>
    </ul>
  </dd>

  <dt>f</dt>
  <dd>
    <ul>
      <li>Specify the delete API for layers and manifests.</li>
    </ul>
  </dd>

  <dt>e</dt>
  <dd>
    <ul>
      <li>Added support for listing registry contents.</li>
      <li>Added pagination to tags API.</li>
      <li>Added common approach to support pagination.</li>
    </ul>
  </dd>

  <dt>d</dt>
  <dd>
    <ul>
      <li>Allow repository name components to be one character.</li>
      <li>Clarified that single component names are allowed.</li>
    </ul>
  </dd>

  <dt>c</dt>
  <dd>
    <ul>
      <li>Added section covering digest format.</li>
      <li>Added more clarification that manifest cannot be deleted by tag.</li>
    </ul>
  </dd>

  <dt>b</dt>
  <dd>
    <ul>
      <li>Added capability of doing streaming upload to PATCH blob upload.</li>
      <li>Updated PUT blob upload to no longer take final chunk, now requires entire data or no data.</li>
      <li>Removed `416 Requested Range Not Satisfiable` response status from PUT blob upload.</li>
    </ul>
  </dd>

  <dt>a</dt>
  <dd>
    <ul>
      <li>Added support for immutable manifest references in manifest endpoints.</li>
      <li>Deleting a manifest by tag has been deprecated.</li>
      <li>Specified `Docker-Content-Digest` header for appropriate entities.</li>
      <li>Added error code for unsupported operations.</li>
    </ul>
  </dd>
</dl>
