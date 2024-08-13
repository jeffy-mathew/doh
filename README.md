# DNS Over ~~HTTPS~~ HTTP
This is a DNS over HTTP implementation (non HTTPS).
It could be used to be deployed as serverless/lamda functions.
Most of the serverless providers takes care of TLS termination at cloud provider's load balancer level, so this could be used as a DoH serverless function that doesn't deal with TLS termination.