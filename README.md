# artifact-manager

artifact-manager provides an HTTP API allowing clients to upload files. When a file is
uploaded and a docker application running in [Marathon](https://mesosphere.github.io/marathon/)
has a volume whose `hostPath` matches, it will get restarted.

Instructions on how to contribute can be found in the `Contribute` section.

## How does it work?

A client uploads a file to artifact-manager, which gets stored to a configurable location. In
the background, queries are made to Marathon to retrieve all of the applications that are
running. For each application, it'll check the volumes being used. If an application depends on
a volume whose `hostPath` matches either the `name` or `dst` (prefixed by the directory being used by the artifact-manager) specified in the http request, it'll be restarted.

### NFS / Local Disk

The initial implementation will work with a local file system, or at least one that acts like it (such as NFS). When a file is uploaded, it will be written to disk. If the HTTP request included a `src` and `dst` two things will happen. One, if the file is an archive (tarball, zip, etc), it will be unpacked (**note**: It's expected the archive has a directory inside of it, containing its files.) Second, a symlink will be created from `src` to `dst`.

## Usage

To run the application, simply execute the binary. There are a few configuration options which all have
default values. You can set the values via command-line flags, or using environment variables.

```
> ./artifact-manager --help
Usage of artifact-manager:
  -addr string
        address to listen on
  -debug
        enable debug logging
  -dir string
        directory where files will be managed (default "/tmp")
  -marathon-hosts string
        comma-delimited list of marathon hosts, "host:port" (default "localhost:8080")
  -marathon-query-interval duration
        time to wait between queries to marathon (default 10s)
  -port int
        port to listen on (default 8900)

Note: environment variables can be defined to override any command-line flag.
The variables are equivalent to the command-line flag names, except that they should be upper-case, hypens replaced by underscores andprefixed with "AM_" (excluding double quotes)
```

For example, to start the artifact-manager listening on port 9000:

```
// set the port via command-line fag
./artifact-manager -port 9000

// set the port via environment variable
AM_PORT=9000 ./aratifact-manager

// alternate method of setting environment variable
export AM_PORT=9000
./artifact-manager
```

## HTTP API

### Upload a File

To upload a file, POST the contents as the body of the request and provide the
URL parameters.

`POST /?name=<file name>&src=<source of symlink>&dst=<destination of symlink>`

Required URL parameters:

* name - the name of the file being uploaded (should just be the filename)

To have a symlink created, the following URL parameters must be provided:

* src - the source for the symlink (relative to the artifact-manager `dir`) that will be created
* dst - the destination for the symlink (relative to the artifact-manager `dir`) that will be created.

If the file is an "archive", it will only be extracted if the `src` and `dst` URL parameters
are provided.

## Example Uploading a tgz file

In this example, we'll assume the artifact-manager has been configured to manage files in a directory named
`/data`.

Imagine we have a file named `myfile-2018-08-10.tgz` that has a directory, which contains a file. If we
extracted the tarball we'd have something like:

```
mydata/
   important_info.txt
```

Submitting this file with cURL would look like:

```
curl -X POST http://artifact-manager.marathon.mesos:8900/?name=myfile-2018-08-10.tgz&src=mydata&dst=myfile-latest --data-binary @myfile-2018-08-10.tgz
```

This file will be written to `/data` and then extracted, so we'd end up with:

```
/data/myfile-2018-08-10.tgz
/data/mydata/
    import_info.txt
```

A symlink, is created using the `src` and `dst` URL parameters, so we'd end up with:

```
/data/myfile-latest -> /data/mydata
```

## Example Uploading a txt file

In this example, we'll assume the artifact-manager has been configured to manage files in a directory named
`/data`.

Imagine we have a file named `notes.txt`.

Submitting this file with cURL would look like:

```
curl -X POST http://artifact-manager.marathon.mesos:8900/?name=notes.txt&src=notes.txt&dst=notes-latest.txt --data-binary @notes.txt
```

This file will be written to `/data` and a symlink will be created, the result being:

```
/data/notes.txt
/data/notes-latest.txt -> /data/notes.txt
```

## Example Uploading a file without creating a symlink

```
curl -X POST http://artifact-manager.marathon.mesos:8900/?name=notes.txt --data-binary @notes.txt
```

The file will be written to `/data/notes.txt` and no symlink will be created.

## Contribute

```
cd $GOPATH/src
mkdir -p apex
cd apex
git clone git@github.com:mindscratch/artifact-manager.git


# build
make build
```
