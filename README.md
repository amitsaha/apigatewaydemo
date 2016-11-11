
```
sudo cat /etc/consul.d/web.json 
[sudo] password for asaha: 
{"service": {"name": "projects", "tags": [""], "port": 5000}}
asaha@localhost:~ $ consul agent -dev -config-dir=/etc/consul.d
```

### Resources

- https://peter.bourgon.org/applied-go-kit/#28


