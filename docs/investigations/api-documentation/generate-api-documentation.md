# Generating API documentation

## Overview

This document discusses how to create a static documentation page from our GraphQL schema.

As of right now, there is no static page with our API documentation, so if any problem occurs the user can either check the `examples` directory or read the schema on his own.
We need a solution that would be available on one of the Compass endpoints and would utilise our generated examples.

## Possible solutions

### 1. Links to the examples on the API playground
The GraphQL playground that we use on our endpoints generates documentation with comments from the schema.
We could create a gqlgen plugin that would add a relative example link to every query and mutation. That way the user
would be able to check how the request looks right away.

Work that has to be done
* gqlgen plugin
* hosting and serving the examples on the director

Pros
* relatively small amount of work to be done
* user sees the version of the examples that he is using
* requires almost no maintenance

Cons
* does not look very impressive (plain text in the browser)

### 2. Using a 3rd party tool to generate static html
[Dociql](https://github.com/wayfair/dociql) is a tool that generates html files with the documented schema.
It uses the introspection query to fetch the API schema and a .yaml file to configure the output.
 
 Unfortunately we have todescribe every query and mutation in the config file if we want them to be sorted properly. Maintaining the config file would be
a chore so we would have to write a tool that generates the config file with the descriptions and examples. Sample config file can be found [here](https://github.com/wayfair/dociql/blob/master/config.yml)

Apart from that we would have to solve the authentication since our endpoints are secured and dociql doesn't support authentication
out of the box. It also supports the 'try it now' feature which too has to be configured to work on the domain that compass
is hosted at the moment.

Work that has to be done
* handling the authentication
* generating the config file
* creating a prow pipeline
* configuring the 'try it now' feature

Pros
* looks much better (example [here](https://wayfair.github.io/dociql/))

Cons
* requires much more work
* has to be maintained

