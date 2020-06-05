# grpc-vpn

<p align="left">
<a href="https://circleci.com/gh/gjbae1212/grpc-vpn"><img src="https://circleci.com/gh/gjbae1212/grpc-vpn.svg?style=svg"></a>
<a href="https://hits.seeyoufarm.com"/><img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fgjbae1212%2Fgrpc-vpn"/></a> 
<a href="https://img.shields.io/badge/language-golang-blue"><img src="https://img.shields.io/badge/language-golang-blue" alt="language" /></a>
<a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-GREEN.svg" alt="license" /></a>
</p>

**GRPC-VPN** is the **VPN** server and client which supports authentications such as Google OpenId Connect or AWS IAM, using GRPC.  
  
Other authentications(LDAP, ...) will be to apply it, if you will implement custom [ServerAuthMethod](https://github.com/gjbae1212/grpc-vpn/tree/master/auth) for VPN server and [ClientAuthMethod](https://github.com/gjbae1212/grpc-vpn/tree/master/auth) for VPN client.

<br/>

**Motivation**  
Of course, the well-made VPN is already around us.  

But many VPN aren't to support multiple authentications. (Generally support to LDAP, configuration file having shared key)  
  
I want to apply multiple authentications by circumstances, so protocol for authentication should organize the plugin which implements it.  

<br/>

**GRPC-VPN**
- It's composed of server and client.    
- It's implemented using **Golang**.
- Communicates via **GRPC**.
- Use Tun device.
- Server on VPN can only run on **Linux**, and Client on VPN can run on **Linux** or **Mac**.
- Supports to inject **custom authentication function**.
- Issued JWT is expired after 1.day. (later it will support to change mannually.)
  
## Why GRPC?
For multiple authentications will support, authentication flow and connection flow should definitely distinguish.  
  
Authentication flow needs to a little unification protocol for applying various case.  
  
Connection flow should have the unitary policy regardless of authentications.  
So I was to use JWT authentication and GRPC.   

Using GRPC can implement the unification authentication flow, also using JWT and GRPC stream can connect VPN regardless of authentications.   
    
<br/>
<p align="center">
<img src="https://storage.googleapis.com/gjbae1212-asset/grpc-vpn/main.png"/>
</p>
<br/>

## Getting Started
- It's possible to be used both a third-party library and standalone application.
- **Server on VPN can only run on Linux, and Client on VPN can run on Linux or Mac.**
- It **must run** with **sudo**, becausue it's included network setting that can change on kernel mode.
  
  
### 1. Be used Third-Party Library.
> You can write code and run, if you utilize it on circumstances.   

**1. Non Authentication**
```go
# -------------------------------------------------
# SERVER
# -------------------------------------------------

import (
    "github.com/gjbae1212/grpc-vpn/server"    
)

s, _ = server.NewVpnServer(
            server.WithVpnSubNet("ex) 192.168.0.100/24"),
      		server.WithGrpcPort("ex) 443"),
      		server.WithVpnJwtSalt("ex) jwt salt"),
      		server.WithVpnJwtExpiration(24*time.Hour),
      		server.WithGrpcTlsCertification("ex) tls cert"),
      		server.WithGrpcTlsPem("ex) tls pem"),       
)
s.Run()

# ------------------------------------------------- 
# CLIENT 
# -------------------------------------------------

import (
    "github.com/gjbae1212/grpc-vpn/client"
        "github.com/gjbae1212/grpc-vpn/auth"
)

c, _ := client.NewVpnClient(
            client.WithServerAddr("ex) server addr"),
     		client.WithServerPort("ex) server port"),
     		client.WithSelfSignedCertification("ex) server tls cert"),     
)
c.Run()
```
<br/>

**2. Authentication [Google OpenID Connect](https://developers.google.com/identity/protocols/oauth2/native-app)**
```go
# -------------------------------------------------
# SERVER
# -------------------------------------------------

import (
    "github.com/gjbae1212/grpc-vpn/server"
    "github.com/gjbae1212/grpc-vpn/auth"
)

authGoogle, _ := auth.NewServerManagerForGoogleOpenID(
		"ex) google openid connect client_id",
		"ex) google openid connect client_secret",
		"ex) if your GSUITE is using, it's domain name to allow",
		[]string{"gjbae1212@gmail.com", "blahblah"},
	)
authMethod, _ := authGoogle.ServerAuth()

s, _ := server.NewVpnServer(
    server.WithVpnSubNet("ex) 192.168.0.100/24"),
    server.WithGrpcPort("ex) 443"),
    server.WithVpnJwtSalt("ex) jwt salt"),
    server.WithVpnJwtExpiration(24*time.Hour),
    server.WithGrpcTlsCertification("ex) tls cert"),
    server.WithGrpcTlsPem("ex) tls pem"),
    server.WithAuthMethods([]auth.ServerAuthMethod{authMethod}), // authentication
)
s.Run()

# ------------------------------------------------- 
# CLIENT 
# -------------------------------------------------

import (
    "github.com/gjbae1212/grpc-vpn/client"
    "github.com/gjbae1212/grpc-vpn/auth"
)

authGoogle, _ := auth.NewClientManagerForGoogleOpenID(
		"ex) google openid connect client_id",
		"ex) google openid connect client_secret",
	)
	
authMethod, _ := authGoogle.ClientAuth()
c, _ := client.NewVpnClient(
    client.WithServerAddr("ex) server addr"),
    client.WithServerPort("ex) server port"),
    client.WithSelfSignedCertification("ex) server tls cert"),
    client.WithAuthMethod(authMethod), // authentication
)
c.Run()

```
<br/>

**3. Authentication(AWS IAM)**
```go
# -------------------------------------------------
# SERVER
# -------------------------------------------------

import (
    "github.com/gjbae1212/grpc-vpn/server"
    "github.com/gjbae1212/grpc-vpn/auth"
)

authAws, _ := auth.NewServerManagerForAwsIAM(
		"ex) allow aws account id",
		[]string{"gjbae1212@gmail.com", "blahblah"},
	)
authMethod, _ := authAws.ServerAuth()

s, _ := server.NewVpnServer(
    server.WithVpnSubNet("ex) 192.168.0.100/24"),
    server.WithGrpcPort("ex) 443"),
    server.WithVpnJwtSalt("ex) jwt salt"),
    server.WithVpnJwtExpiration(24*time.Hour),
    server.WithGrpcTlsCertification("ex) tls cert"),
    server.WithGrpcTlsPem("ex) tls pem"),
    server.WithAuthMethods([]auth.ServerAuthMethod{authMethod}), // authentication
)
s.Run()

# ------------------------------------------------- 
# CLIENT 
# -------------------------------------------------

import (
    "github.com/gjbae1212/grpc-vpn/client"
    "github.com/gjbae1212/grpc-vpn/auth"
)

authAws, _ := auth.NewClientManagerForAwsIAM(
		"ex) aws key",
		"ex) aws secret access key",
	)
authMethod, _ := authAws.ClientAuth()

c, _ := client.NewVpnClient(
    client.WithServerAddr("ex) server addr"),
    client.WithServerPort("ex) server port"),
    client.WithSelfSignedCertification("ex) server tls cert"),
    client.WithAuthMethod(authMethod), // authentication
)
c.Run()
```

### 2. Be used Standalone Application.
> You can run an application which already built.

**1. Download or build**
```bash
$ git clone https://github.com/gjbae1212/grpc-vpn.git
$ cd grpc-vpn
$ bash script/make.sh build_vpn_server # make an application to dist directory.
$ bash script/make.sh build_vpn_client # make an application to dist directory.
```
<br/>

**2. Run Server**
server config(config.yaml)
```yaml
# Template
vpn:
  port: "" # Required(vpn port)
  subnet: "" # Required(vpn subnet(private ip range), ex) 192.168.0.100/24)
  log_path: "" # Required(log path)
  jwt_salt: "" # Required(random string)
  jwt_expiration: "" # Required(expire-time in JWT), ex) 100ms, 10m, 2h30m, ...  
  tls_certification: "" # Required(tls cert)
  tls_pem: "" # Required(tls pem)

auth: # Optional 
  google_openid: # Optional(if you want to google openid connect authentication)
    client_id: "" # Google client id
    client_secret: "" # Google client secret
    hd: "" # If you are using GSuite, domain name for allowing.
    allow_emails: # Allow emails
      - ""
  aws_iam: # Optional(if you want to aws iam authentication)
    account_id: "" # Allow AWS Account ID 
    allow_users: # Allow users
      - ""

----------------------------------------------------------------------

# 1. Example Non Authentication
vpn:
  port: "8080" 
  subnet: "192.168.0.100/24" 
  log_path: "" 
  jwt_salt: "hello-world" 
  jwt_expiration: "2h10m" 
  tls_certification: "blahblah" 
  tls_pem: "blahblah" 

# 2. Example Google OpenId Connect (Reference https://developers.google.com/identity/protocols/oauth2/native-app)
vpn:
  port: "8080" 
  subnet: "192.168.0.100/24" 
  log_path: "" 
  jwt_salt: "hello-world" 
  jwt_expiration: "1h10m"
  tls_certification: "blahblah" 
  tls_pem: "blahblah"
auth:  
  google_openid: 
    client_id: "google client id" 
    client_secret: "google client secret" 
    hd: "" 
    allow_emails: 
      - "blahblah@gmail.com"

# 3. Example Aws IAM
vpn:
  port: "8080" 
  subnet: "192.168.0.100/24" 
  log_path: "" 
  jwt_salt: "hello-world"
  jwt_expiration: "3h"
  tls_certification: "blahblah" 
  tls_pem: "blahblah"
auth:    
  aws_iam: 
    account_id: "aws accoount id"  
    allow_users: 
      - "blahblah"

```
Run
```bash
$ cd grpc-vpn/dist
$ sudo vpn-server-linux run -c "config.yaml path" 
```
<br/>

**3. Run Client**
client config(config.yaml)
```yaml
# Template
vpn:
  addr: "" # Required(vpn server addr)
  port: "" # Required(vpn server port)
  insecure: true or false # Required (true is to disable tls, false is to enable tls)
  self_signed_certification: "" # Optional(If you are using self-signed certification, you must insert it.)
auth: # Optional
  google_openid: # Optional(if your vpn-server support to google openid connect authentication)
    client_id: ""
    client_secret: ""
  aws_iam:  # Optional(if your vpn-server support to aws iam authentication)
    access_key: ""
    secret_access_key: ""

---------------------------------------------------------------

# 1. Example Non Authentication
vpn:
  addr: "localhost" 
  port: "8080"
  insecure: false 
  self_signed_certification: ""   

# 2. Example Google OpenId Connect (Reference https://developers.google.com/identity/protocols/oauth2/native-app)
vpn:
  addr: "localhost" 
  port: "8080"
  insecure: false
  self_signed_certification: ""   
auth: 
  google_openid: 
    client_id: "gogole client id"
    client_secret: "google client secret"

# 3. Example Aws IAM
vpn:
  addr: "localhost" 
  port: "8080"
  insecure: false
  self_signed_certification: ""  
auth: 
  aws_iam:  # Optional(if your vpn-server support to aws iam authentication)
    access_key: "aws access key"
    secret_access_key: "aws secret key"  
Run
```bash
$ cd grpc-vpn/dist

# MAC
$ sudo vpn-client-darwin run -c "config.yaml path"

# LINUX 
$ sudo vpn-client-linux run -c "config.yaml path" 
```

## License
This project is following The MIT.
