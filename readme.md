# ![BDM](bdm.svg) Binary Data Manager

![CI](https://github.com/cry-inc/bdm/workflows/CI/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/cry-inc/bdm)](https://goreportcard.com/report/github.com/cry-inc/bdm)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

The Binary Data Manager is a system that allows users to create a versioned repository of packages. A package is just a folder with a bunch of files in it. Packages are immutable, once published they stay always the same and cannot be modified. If you want to change a package, you must publish a new version of the package. The system was designed to handle binary assets used for testing in software development, but can be also used for other kinds of assets and artifacts that can be represented as a simple set of files.

The system works centralized with a repository server. This server holds all packages. Clients can upload new packages or download them from the server. The server uses HTTP to communicate with clients.

Both, server and client, are contained in the same portable CLI tool called "bdm". It can be used to start a server or to download and upload packages and needs no additional files or dependencies.

## Features

* Client and server in a single portable binary (implemented in pure Go)
* File deduplication for the server, network transfer and caches
* File verification and identification using fast cryptographic hashes (BLAKE3)
* Packages are described and validated using JSON manifests
* Compressed server side storage and network data transfer (zstd)
* Optional client side caching to avoid network transfers
* Intelligent downloading/restore of packages to minimize time and costs
* Simple user system with separate read, write and admin permissions
* Web interface can be used to create tokens for use with the command line client or HTTP API
* Simple web interface for browsing and downloading packages without a client application
* Built-in HTTPS support for automated Let's Encrypt certificate (or bring you own certificate)
* Docker image for easy deployment (see below)

## Quickstart

1. Show version number of application: `bdm -about`
2. Start the repository server: `bdm -server -guestreading -guestwriting -port=2323`
3. Upload a new package: `bdm -upload -package="foo" -input="path/to/foo-folder/" -remote="http://127.0.0.1:2323"`
4. Download the package: `bdm -download -package="foo" -version=1 -output="where/to/put/foo/" -remote="http://127.0.0.1:2323"`
5. Verify existing download: `bdm -check -package="foo" -version=1 -input="where/to/check/foo/" -remote="http://127.0.0.1:2323"`
6. Open the URL `http://127.0.0.1:2323` in your browser for the web UI to inspect packages
7. Run `bdm -help` to show additional CLI documentation

## Limitations

* Packages cannot contain empty folders, just like git repositories
* No support or integration of existing account systems

## Docker

1. Use `docker build . -t=bdm` to build your own Docker image from the source code
2. You can also run `docker pull ghcr.io/cry-inc/bdm:latest` to download the latest pre-built image (based on the master branch)
3. Run `docker run --rm -p 2323:2323 -v /host/folder:/bdmdata bdm` to start a HTTP server on the (default) port 2323 and a persistent data location on the host file system. BDM will create an default admin account and display the randomly generated initial password during the first start.
4. Run `docker run --rm -p 443:443 -e BDM_PORT=443 -e BDM_HTTPS_CERT=/path/cert.pem -e BDM_HTTPS_KEY=/path/key.pem -v /host/folder:/bdmdata bdm` to start a HTTPS server using a pre-existing certificate. The certificate and key files need to be mounted into the container.
5. Run `docker run --rm -p 2323:2323 -p 80:80 -e BDM_LETS_ENCRYPT=mydomain.com -v /host/folder:/bdmdata bdm` to start a HTTPS server using a cached Let's Encrypt certificate. In this case port 80 needs to be reachable from the Internet. After the certificate acquisition it will redirect to the HTTPS port of the server.
6. Check the Dockerfile for additional optional environment variables.

## User accounts and tokens

To avoid all accounts and permissions, you can use the arguments `-guestreading` and `-guestwriting` when starting the server. This will allow everyone to download and upload packages without any restrictions. THIS IS NOT RECOMMENDED! Even for private networks I suggested to at least use a shared secret token for writing to restrict uploading new packages.

When starting the server for the first time, BDM will create an default admin account with a random password. This default admin is only created if the user database is empty. You can customize the name of the default admin account using the argument `-defaultuser youradminname`. The random password will be printed only once during startup to the console. You can use the password to log into the web interface and create more users and tokens.

A token is a kind of special long password that can be used without a user name. You need them to upload and download packages with the client if guest access is not enabled. Each token can have specific permissions and belongs to a user. If the user no longer exists, the token will stop working. If a user no longer has the permissions required by the token, it will also stop working. Tokens can be created and deleted in your profile using the web interface.

## Why another package server/client?

There are already lots of existing systems for packages and binary artifact management. All of them have different pros and cons and are often intended for very different purposes. Systems like NuGet and NPM are designed around managing libraries for application development. DVC is tailored to Machine Learning and use with git. Other systems, like Microsofts Universal Packages and the Generic Artifactory packages, require expensive software licenses or are only implemented by paid cloud services.

At my work we were looking for a system that could reliably manage large and small binary test assets for automated and manual software tests. We already tried and used a lot of different approaches over the years. SVN and GIT repos (with and without LFS), network file shares, OneDrive/Box.com, NuGet and more. All of them failed us eventually for different reasons.

After coming up with the idea of an immutable data store behind an HTTP interface, I decided to get my hands dirty and implement a prototype in my spare time.

The specific requirements that I was trying to satisfy were:
* Lightweight and portable server and client application
* File deduplication and compression to minimize storage and bandwidth costs
* Intelligent package restore (omits downloading and restoring existing files)
* Local caching for slow or expensive Internet connections
* Easy backup of all package repository data on the server
* Simple API or client library for integration in custom applications and scripts
* Downloading of packages and files using a persistent URL via browser without the need for a client application
* Robust verification of all package data
* Client application that can be used manually or scripted to upload and download packages

## Implementation Details

A package is described by a manifest. A manifest is a JSON document that contains package metadata like name, version number and publication date. It also contains a list of all files contained in the package. For each file it contains the file path relative to the package root folder. It also contains the file size in bytes and a BLAKE3 hash of the file content.

The hash and the file size together are used to identify what is called an object. Objects consist of a hash, a size and the content itself, but have no file name. This means that if you have the same file in two different folders of your package, both files will refer to the same object, even when they have different file names.

When a client tries to publish a new package version, it first generates locally what is called an unpublished manifest. The difference to a published manifest from above is that it does not yet have a version number and a publication date, since these are assigned by the server. After generating the manifest, the client checks for each file if the server has already a corresponding object. If not, it uploads the missing object. Since we upload only objects and not files, duplicate files between different packages (regardless of package name or version) are only stored once on the server.

After the object upload, the client will try to publish the manifest. The server will then check if all objects used in the manifest exist already on the server. If that is the case, the server will assign a new version number and publication date to publish the manifest.

When downloading a package, the client will first get the package manifest from the server. Then it will use the file list from the downloaded manifest to compare it with the output directory. If a file from the manifest already exists with the correct file size and hash, it will be skipped and not downloaded again. Different or missing files will be downloaded from the server. If client caching is enabled, the client will look for manifests and objects in the local cache to avoid network traffic. If the manifest or object needs to be downloaded, it will be also added to the cache on-the-fly.

To minimize required disk space on the server for object storage, all objects are stored using ZSTD compression. To minimize network traffic, the objects are also compressed using ZSTD when they are transferred between the client and server. To minimize the memory footprint of the client and server, all file IO around the objects is implemented using streaming operations, including the compression/decompression steps. This also means that there is no hard limit for file sizes.
