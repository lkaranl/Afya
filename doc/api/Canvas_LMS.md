# Canvas LMS

Canvas LMS includes a REST API for accessing and modifying data externally from the main application, in your own programs and scripts. This documentation describes the resources that make up the API.

To get started, you'll want to review the general basics, including the information below and the page on [Authentication using OAuth2](/services/canvas/oauth2/file.oauth.md).

## API Changes

For API resources, such as the API Change Log for additions, changes, deprecations, and removals, view the [Canvas API page](https://community.canvaslms.com/t5/Change-Log/tkb-p/changelog) in the Canvas Community.

## API Policy

Please carefully review [the Canvas API Policy](https://www.instructure.com/policies/api-policy) before using the API.

## Schema

All API access is over HTTPS, against your normal Canvas domain.

All API responses are in [JSON format](http://www.json.org/).

All integer ids in Canvas are 64 bit integers. String ids are also used in Canvas.

To force all ids to strings add the request header `Accept: application/json+canvas-string-ids` This will cause Canvas to return even integer IDs as strings, preventing problems with languages (particularly JavaScript) that can't properly process large integers.

All boolean parameters can be passed as true/false, t/f, yes/no, y/n, on/off, or 1/0. When using JSON format, a literal true/false is preferred, rather than as a string.

For POST and PUT requests, parameters are sent using standard [HTML form encoding](http://www.w3.org/TR/html4/interact/forms.html#h-17.13.4) (the application/x-www-form-urlencoded content type).

POST and PUT requests may also optionally be sent in [JSON format](http://www.json.org/) format. The content-type of the request must be set to application/json in this case. There is currently no way to upload a file as part of a JSON POST, the multipart form type must be used.

As an example, this HTML form request:

```bash
name=test+name&file_ids[]=1&file_ids[]=2&sub[name]=foo&sub[message]=bar&flag=y
```

would translate into this JSON request:

```json
{ "name": "test name", "file_ids": [1,2], "sub": { "name": "foo", "message": "bar" }, "flag": true }
```

With either encoding, all timestamps are sent and returned in ISO 8601 format (UTC time zone):

```
YYYY-MM-DDTHH:MM:SSZ
```

## Authentication

API authentication is done with OAuth2. If possible, using the HTTP Authorization header is recommended. Sending the access token in the query string or POST parameters is also supported.

OAuth2 Token sent in header:

```bash
curl -H "Authorization: Bearer <ACCESS-TOKEN>" "https://canvas.instructure.com/api/v1/courses"
```

OAuth2 Token sent in query string:

```bash
curl "https://canvas.instructure.com/api/v1/courses?access_token=<ACCESS-TOKEN>"
```

Read more about [OAuth2 and how to get access tokens.](/services/canvas/oauth2/file.oauth.md)

## SSL

Note that if you make an API call using HTTP instead of HTTPS, you will be redirected to HTTPS. However, at that point, the credentials have already been sent in clear over the internet. Please make sure that you are using HTTPS.

## Canvas Experiences

Canvas LMS supports several experiences including Canvas Career and Canvas for Elementary. The vast majority of these API resources are shared, though some are applicable only to certain experiences.

## About this Documentation

This documentation is generated directly from the Canvas LMS code. You can generate this documentation yourself if you've set up a local Canvas environment following the instructions on [Github](https://www.github.com/instructure/canvas-lms/wiki). Run the following command from your Canvas directory:

```bash
bundle exec rake doc:api
```

### OpenAPI 3.0 Specification

Canvas also provides an OpenAPI 3.0 specification that can be generated from the same YARD documentation. This modern format is compatible with tools like Swagger UI, Postman, and various code generators.

To generate the OpenAPI 3.0 specification:

```bash
bundle exec rake doc:openapi
```

This will create `public/doc/openapi/canvas.openapi.yaml` containing the OpenAPI 3.0 specification for the Canvas API.

The OpenAPI spec includes:

* All API endpoints with their HTTP methods and paths
* Request parameters (path, query, and body)
* Response schemas
* Authentication requirements
* Server configuration

#### ✅ Successful Response (with status 200 or 201):

A successful response usually looks something like this (this is just an example):

```json
{
  "id": 123,
  "name": "Introduction to Biology",
  "course_code": "BIO101"
}
```

Each response includes data fields (like `id`, `name`, and `course_code`) clearly describing what you've requested. You can use these fields directly in your applications.

#### ⚠️ Error Response (status 4xx):

If something goes wrong, you'll get an error response, for example:

```json
{
  "errors": [
    {
      "message": "Invalid access token."
    }
  ]
}
```

Errors come with clear messages indicating the problem. Common issues include authentication errors, incorrect parameters, or permission issues.

#### 📑 Pagination and Large Data:

Some services, for example Canvas LMS, split API calls into multiple pages when you request a lot of data (like a long list of courses or users):

* Check for a Link header in the response, which provides URLs for the next or previous page of results.
* Example header:

```bash
Link: <https://canvas.instructure.com/api/v1/courses?page=2>; rel="next"
```

***

This documentation is generated directly from the Canvas LMS source code, available [on Github](https://github.com/instructure/canvas-lms).