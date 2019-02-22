# PASS metadata schemas

[![Build Status](https://travis-ci.com/OA-PASS/metadata-schemas.svg?branch=master)](https://travis-ci.com/OA-PASS/metadata-schemas)

This repository contains JSON schemas and example data intended to describe PASS submission metadata as per the [schemas for forms and validation](https://docs.google.com/document/d/1sLWGZR4kCvQVGv-TA5x8ny-AxL3ChBYNeFYW1eACsDw/edit) design, as well as a [schema service](https://docs.google.com/document/d/1Ki6HUYsEkKPduungp5gHmr7T_YrQUiaTipjorcSnf4s/edit) that will retrieve,
dereference, and place in the correct order all schemas relevant to a given set of pass Repositories.

## Schemas

The JSON schemas herein describe the JSON metadata payload of PASS [submission](https://oa-pass.github.io/pass-data-model/documentation/Submission.html) entities.  They serve two purposes
    1. Validation of submission metadata
    2. Generation of forms in the PASS user interface

These schemas follow a defined structure where properties in `/definitions/form/properties` are intended to be displayed by a UI, e.g.

    {
        "title": "Example schema",
        "description": "NIHMS-specific metadata requirements",
        "$id": "https://github.com/OA-PASS/metadata-schemas/jhu/example.json",
        "$schema": "http://json-schema.org/draft-07/schema#",
        "type": "object",
        "definitions": {
            "form": {
                "title": "Please provide the following information",
                "type": "object",
                "properties": {
                    "journal": {
                        "$ref": "global.json#/properties/journal"
                    },
                    "ISSN": {
                        "$ref": "global.json#/properties/ISSN"
                    }
                }
            },
        },
        "allOf": [
            {
                "$ref": "global.json#"
            },
            {
                "$ref": "#/definitions/form"
            }
        ]
    }

A pass [repository](https://oa-pass.github.io/pass-data-model/documentation/Repository.html) entity represents a target repository where
submissions may be submitted.  Each repository may link to one or more JSON schemas that define the repository's metadata requirements.
In terms of expressing a desired user interface experience, one may observe a general pattern of pointing to a "common" schema containing ubiquitous fields, and additionally pointing to a "repository-specific" schema containing any additional fields that are unique to a given repository.

As a concrete example, the NIHMS repository may point to the [common.json](jhu/common.json) schema, as well as the [nihms.json](jhu/nihms.json)
schema.

## Schema service

The schema service is an http service that accepts a list of PASS [repository](https://oa-pass.github.io/pass-data-model/documentation/Repository.html) entity URIs as `application/json` or newline delimited `text/plain`, in a POST request.  for example:

    [
        "http://pass.jhu.edu/fcrepo/rest/repositories/foo",
        "http://pass.jhu.edu/fcrepo/rest/repositories/bar",
    ]

For each repository, the schema service will retrieve the list of schemas relevant to the repository, place that list in the correct order (so
that schemas that provide the most dependencies are displayed first), and resolves all `$ref` references that might appear in the schema.

The result is an `application/json` response that contains a JSON list of schemas.

### building

Building the schema service requires go 1.11 or later.

The schema service may be built by running:

    go build ./cmd/schemas

.. which will create an executable in the current directory.  `go install ./cmd/schemas` may be used instead, which will install the binary to your `${GOPATH/bin}`.  If you have that in your `$PATH`, this is particularly convenient for building and running cli commands.

### running

To get a list of command line options, do 

    schemas help serve

To run the schema service,

    schemas serve /path/to/schemas

where `/path/to/schemas` is a directory, or files(s) containing JSON schemas.  This statically loads the set of schemas this service may return.  For example:

    $ ./schemas serve jhu/
    2019/02/21 16:40:40 Loaded schema https://github.com/OA-PASS/metadata-schemas/jhu/common.json
    2019/02/21 16:40:40 Loaded schema https://github.com/OA-PASS/metadata-schemas/jhu/global.json
    2019/02/21 16:40:40 Loaded schema https://github.com/OA-PASS/metadata-schemas/jhu/jscholarship.json
    2019/02/21 16:40:40 Loaded schema https://github.com/OA-PASS/metadata-schemas/jhu/nihms.json
    2019/02/21 16:40:40 Listening on port 59152

This output shows the random port the server is listening on, and lists the schemas it loaded.

#### configuration/options

The help page describes the possible commandline options.  Each option has a corresponding environment variable that may be used instead:

    $ ./schemas help serve
    NAME:
        schemas serve - Sereve the PASS schema service over http

    USAGE:
        schemas serve [command options] [ file | dir ] ...

    DESCRIPTION:


        An optional list of files or directories may be provided, which will be
        examined for the presence of schema files which will be used for static lookups


    OPTIONS:
        --external value, -e value  External (public) PASS baseuri [$PASS_EXTERNAL_FEDORA_BASEURL]
        --internal value, -i value  Internal (private) PASS baseuri [$PASS_FEDORA_BASEURL]
        --username value, -u value  Username for basic auth to Fedora [$PASS_FEDORA_USER]
        --password value, -p value  Password for basic auth to Fedora [$PASS_FEDORA_PASSWORD]
        --port value                Port for the schema service http endpoint (default: 0) [$SCHEMA_SERVICE_PORT]

Command line options have a short form (`-i`) or a long form (`--internal`), which may be user interchangably.  For example, the following
will run the schema service on port 8080, and user the username `myUser` and passeord `foo` for retrieving repository entities from the Fedora

    env SCHEMA_SERVICE_PORT=8080 schemas serve -u myUser --password foo  /path/to/schemas

