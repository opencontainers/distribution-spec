local openapi = import 'openapi.libsonnet';

local contact = openapi.contact.New(
  name='John Doe',
  url='http://localhost',
  email='john.doe@localhost',
);

local license = openapi.license.New(
  name='Apache 2.0',
  url='http://www.apache.org/licenses/LICENSE-2.0.html',
);

local info = openapi.info.New(
  title='Test App',
  version='v1.1.1',
  description='A test app.',
  termsOfService='http://localhost/tos',
  contact=contact,
  license=license,
);

local username = openapi.serverVariable.New(
  name='username',
  default='demo-0',
  description='Username description',
  enum=['demo-1', 'demo-2', 'demo-3']
);

local port = openapi.serverVariable.New(
  name='port',
  default='443',
  description='Username description',
  enum=['80', '75', '22']
);

local servers = [

  openapi.server.New(
    url='http://server-1.localhost',
    description='Server-1 description',
  )
  .addVariable(username),

  openapi.server.New(
    url='http://server-2.localhost',
    description='Server-2 description',
  )
  .addVariable(port)
  .addVariable(username),
];

local respfoo =
  openapi.response.new('Not Found');

local getFooResp =
  openapi.responses.new()
  .addResponse('404', respfoo)
;

local getFoo =
  openapi.operation.new()
  .addResponses(getFooResp)
;

local fooPath =
  openapi.pathItem.new()
  .addOperation('GET', getFoo);

local paths =
  openapi.paths.new()
  .addPath(
    '/foo',
    fooPath
  );

openapi.New(
  info=info,
  paths=paths,
  servers=servers,
)
