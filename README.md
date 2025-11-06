protoc-gen-redact (PGR)
=======================
[![Build and Publish](https://github.com/menta2k/protoc-gen-redact/workflows/Build%20and%20Publish/badge.svg)](https://github.com/menta2k/protoc-gen-redact/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/menta2k/protoc-gen-redact/v3?dropcache)](https://goreportcard.com/report/github.com/menta2k/protoc-gen-redact/v3)
[![Go Reference](https://pkg.go.dev/badge/github.com/menta2k/protoc-gen-redact/v3.svg)](https://pkg.go.dev/github.com/menta2k/protoc-gen-redact/v3)
[![License](https://img.shields.io/badge/license-apache2-mildgreen.svg)](./LICENSE)
[![GitHub release](https://img.shields.io/github/release/menta2k/protoc-gen-redact.svg)](https://github.com/menta2k/protoc-gen-redact/releases)

_protoc-gen-redact (PGR)_ is a protoc plugin to redact field values in GRPC client calls from the server. This plugin
adds support to protoc-generated code to redact certain fields in the GRPC calls.

## Attribution

This project is a derivative work based on the original [protoc-gen-redact](https://github.com/arrakis-digital/protoc-gen-redact) by **Shivam Rathore** (Copyright 2020).

**Original Author:** Shivam Rathore
**Original Project:** https://github.com/arrakis-digital/protoc-gen-redact
**Contributors:** John Castronuovo

This fork includes enhancements and modifications including:
- Comprehensive error handling and validation
- Extensive test suite (374+ tests)
- Support for proto3 optional fields with correct pointer semantics
- Custom template file support for code generation
- Integration tests with actual protoc compilation
- Improved documentation and examples

All modifications are licensed under the Apache License 2.0, consistent with the original project.

Developers only need import the PGR extension and annotate the messages or fields in their proto files to redact:

```protobuf
syntax = "proto3";

package user;

import "redact/redact.proto";
import "google/protobuf/empty.proto";

option go_package = "github.com/arrakis-digital/protoc-gen-redact/v3/examples/user/pb;user";

message User {
    // User credentials
    string username = 1;
    string password = 2 [(redact.redact) = true]; // default redaction

    // User information
    string email = 3 [(redact.custom).string = "r*d@ct*d"];
    string name = 4;
    Location home = 5;
    message Location {
        double lat = 1;
        double lng = 2;
    }
}

service Chat {
    rpc GetUser(GetUserRequest) returns (User);
    rpc GetUserInternal(GetUserRequest) returns (User) {
        option (redact.method_skip) = true;
    }
    rpc ListUsers (google.protobuf.Empty) returns (ListUsersResponse) {
        option (redact.internal_method) = true;
    }
}

message GetUserRequest {
    string username = 1;
}

message ListUsersResponse {
    repeated User users = 1;
}

```

## Advanced Features

### Custom Code Generation Templates

protoc-gen-redact supports using custom templates for code generation, allowing you to modify the generated code to match your specific requirements.

#### Using a Custom Template

To use a custom template, pass the `template_file` parameter to protoc via `--redact_opt`:

```bash
protoc \
  --plugin=protoc-gen-redact=/path/to/protoc-gen-redact \
  --redact_out=. \
  --redact_opt=template_file=/path/to/your/template.tmpl \
  your_proto_file.proto
```

#### Example Template

An example template is provided in `examples/custom-template.tmpl`. You can use this as a starting point for your customizations:

```bash
protoc \
  --plugin=protoc-gen-redact=./protoc-gen-redact \
  --redact_out=. \
  --redact_opt=template_file=./examples/custom-template.tmpl \
  examples/user/pb/user.proto
```

#### Documentation

For complete documentation on custom templates including:
- Template structure and data types
- Available template functions
- Use cases and examples
- Troubleshooting guide

See [examples/CUSTOM_TEMPLATE.md](examples/CUSTOM_TEMPLATE.md)

## Development and CI/CD

This project includes a comprehensive build system and CI/CD pipeline:

### Build System (Makefile)

The project uses Make for build automation with 50+ targets organized into categories:

```bash
# See all available targets
make help

# Development workflow
make fmt              # Format code
make lint             # Run all linters
make test             # Run all tests
make test-short       # Quick tests during development
make build            # Build the plugin

# Before committing
make pre-commit       # Run fmt + lint + test-short

# Full CI pipeline
make ci-full          # Full CI with coverage and buf checks
```



Request for Contribution
------------------------
Contributors are more than welcome and much appreciated. Please feel free to open a PR to improve anything you don't
like, or would like to add.

Please make your changes in a specific branch and create a pull request into master! If you can, please make sure all
the changes work properly and does not affect the existing functioning.

No PR is too small! Even the smallest effort is countable.

License and Attribution
-----------------------

This project is licensed under the [Apache License 2.0](./LICENSE).

Copyright 2020 Shivam Rathore (Original Work)
Copyright 2025 Contributors (Modifications)

This is a derivative work based on the original protoc-gen-redact project. All attribution notices, copyright statements, and license terms from the original work have been retained in accordance with the Apache License 2.0.

See the [NOTICE](./NOTICE) file for detailed attribution and a list of modifications

