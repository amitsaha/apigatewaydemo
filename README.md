### Start the web application

### Start the gRPC service

### Consul 

```
# Install consul
...
# Copy the JSON files to /etc/consul.d

# Start the agent
$ consul agent -dev -config-dir=/etc/consul.d
```

### Start the API gateway



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
$ http POST 127.0.0.1:8000/projects/ title=MyProejct Auth-Header-V1:121
HTTP/1.1 200 OK
Content-Length: 31
Content-Type: application/json; charset=utf-8
Date: Tue, 15 Nov 2016 02:48:52 GMT

{
    "id": 123,
    "url": "Project-123"
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

- https://peter.bourgon.org/applied-go-kit/#28


