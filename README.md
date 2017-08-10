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

## Example

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
curl -X POST http://artifact-manager.marathon.mesos:8900/?src=mydata&dst=myfile-latest -d @myfile-2018-08-10.tgz
```

This file will be written to `/data` and then extracted, so we'd up with:

```
/data/a42dacf31de.tgz   <-- random name for the file that was uploaded
/data/mydata/
    import_info.txt
```

A symlink, is created using the `src` and `dst` URL parameters, so we'd end up with:

```
/data/myfile-latest -> /data/mydata
```