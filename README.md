Store file download stats in Redis
==================================

A small service to run as an nginx `post_action` handler to track
file downloads in a Redis database in real time.

nginx config
------------

To make this work on the nginx side, you need to use the undocumented magic `post_action` directive, like so:

```
server {
    # .. other config goes here
    location / {
        # .. other config goes here
        post_action @track;
    }
    location @track {
        # Don't modify the $uri variable
        internal;

        # only ever send GET requests, all the magic is in headers anyway
        proxy_method GET;

        # Don't touch the Location header
        proxy_redirect off;

        # Set our magic header fields
        proxy_set_header X-Post-Action 1;
        proxy_set_header X-Track-URI $uri;
        proxy_set_header X-Track-Status $status;
        proxy_set_header X-Track-Complete $request_completion;

        # Don't actually pass the original request. We don't care.
        proxy_pass_request_body off;
        proxy_pass_request_headers off;

        # Point this to whereever your dl-server instance is running
        proxy_pass http://127.0.0.1:8725;
    }
```

Service configuration
---------------------

Use the command line options or a TOML file to set the configuration.
See [`settings.toml`](settings.toml) for an example.

License
-------

This tool is licensed under the [Apache License version
2.0](http://www.apache.org/licenses/LICENSE-2.0).
See the [`LICENSE`](LICENSE) file for details.
