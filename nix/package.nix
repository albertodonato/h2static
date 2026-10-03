{
  lib,
  buildGoModule,
}:

buildGoModule rec {
  pname = "h2static";
  version = "2.4.8";

  src = lib.cleanSource ../.;

  vendorHash = "sha256-OxEJOeNcpinyQoCWZwVMHiY/XzYhkhhI4Duwt3SFa1E=";

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X github.com/albertodonato/h2static/internal/version.appVersion=${version}"
  ];

  meta = {
    description = "Tiny static web server with TLS and HTTP/2 support";
    homepage = "https://github.com/albertodonato/h2static";
    license = lib.licenses.eupl12;
    maintainers = [
      {
        name = "Alberto Donato";
        github = "albertodonato";
        githubId = 3472143;
      }
    ];
    mainProgram = "h2static";
  };
}
