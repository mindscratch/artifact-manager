# artifact-manager

artifact-manager works with applications deployed using Marathon to update their artifacts.

## How does it work?

A client uploads a file to artifact-manager, which gets stored to a configurable location. Then
it queries Marathon to retrieve all of the applications that are running. For each application,
it'll check the artifacts being used. If an application depends on an artifact which has the
same name, specified in the http request, it'll be restarted.

### NFS / Local Disk

The initial implementation will work with a local file system, or at least one that acts like it (such as NFS). When
a file is uploaded, it will be written to disk. If the file is an archive (tarball, zip), it will be unpacked. It's
expected the archive has a directory inside of it, containing its files.

After putting the files to disk, a symlink will be created to it (name provided during upload).

## HTTP API

### Upload a File

To upload a file, POST the contents as the body of the request and provide the
URL parameters.

`POST /?name=<file name>&src=<source of symlink>&dst=<destination of symlink>`

Required URL parameters:

* name - the name of the file being uploaded (should just be the filename)

To have a symlink created, the following URL parameters must be provided:

* src - the source for the symlink that will be created
* dst - the destination for the symlink that will be created.

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
curl -X POST http://artifact-manager.marathon.mesos:8900/?name=myfile-2018-08-10.tgz&src=mydata&dst=myfile-latest -d @myfile-2018-08-10.tgz
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
curl -X POST http://artifact-manager.marathon.mesos:8900/?name=notes.txt&src=notes.txt&dst=notes-latest.txt -d @notes.txt
```

This file will be written to `/data` and a symlink will be created, the result being:

```
/data/notes.txt
/data/notes-latest.txt -> /data/notes.txt
```

## Example Uploading a file without creating a symlink

```
curl -X POST http://artifact-manager.marathon.mesos:8900/?name=notes.txt -d @notes.txt
```

The file will be written to `/data/notes.txt` and no symlink will be created.