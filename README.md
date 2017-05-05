# Demo API Gateway

This is a Demo API gateway implemented in golang. The `apigateway` code lives in the `apigateway` sub-directory and there are two example services:

- A Python web service, `webapp-1`
- A gRPC server, `rpc-app-1`

# Kubernetes deployment

Read about the journey [here](https://github.com/amitsaha/amitsaha.github.io/blob/site/content/demo-api-gateway-kubernetes.rst)

# TODO

API gateway:

- Aim to make the API gateway generic:
  - a YAML file with "/path" => "service", {"http", "grpc"} should be all that's required
- Implement service level rate limiting
- Implement auth header checking by contacting a real service backed via DB

General:

- Deploy in AWS
- Circuit breaking (linkerd)
- Implement structured and correlated logging (https://kubernetes.io/docs/tasks/debug-application-cluster/logging-elasticsearch-kibana/)
- Distributed tracing (linkerd)
- Telemetry (linkerd)
- Setup blue green deployment, CI for the individual services


### Make requests

### API Gateway Features


#### Auth header missing:

```bash
$ http POST 127.0.0.1:8000/projects/ title=MyProejct Auth-Header-V:121
HTTP/1.1 400 Bad Request
Content-Length: 31
Content-Type: text/plain; charset=utf-8
Date: Tue, 15 Nov 2016 02:50:19 GMT
X-Content-Type-Options: nosniff

Decode: Auth-Header-V1 missing
```

#### Rate limiting example response:

```
$ http POST 127.0.0.1:8000/projects/ title=MyProejct Auth-Header-V1:121
HTTP/1.1 503 Service Unavailable
Content-Length: 24
Content-Type: text/plain; charset=utf-8
Date: Tue, 15 Nov 2016 02:49:39 GMT
X-Content-Type-Options: nosniff

Do: rate limit exceeded

```

### Project creation

```bash
$ http POST 127.0.0.1:8000/projects/ Auth-Header-V1:3121 title=foobar11121
HTTP/1.1 200 OK
Content-Length: 39
Content-Type: application/json; charset=utf-8
Date: Sat, 26 Nov 2016 05:48:14 GMT

{
    "id": 123,
    "url": "Project-foobar11121"
}
```


### Verification service:


```bash
$ http GET 127.0.0.1:8000/verify/ id:=1233 token=vasds Auth-Header-V1:121
HTTP/1.1 200 OK
Content-Length: 40
Content-Type: application/json; charset=utf-8
Date: Tue, 15 Nov 2016 02:47:21 GMT

{
    "Err": null,
    "Message": "Verified: 1233"
}

```

### Resources

- https://peter.bourgon.org/applied-go-kit
- https://gokit.io


