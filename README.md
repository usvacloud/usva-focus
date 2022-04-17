# usvad-focus

```
docker run -p 80:8080 mattipaksula/usvad-focus
```

## oneliner to update and leave running

```shell
docker pull mattipaksula/usvad-focus \
&& docker rm -f usvad-focus \
&& docker run -d --name usvad-focus --rm -p 80:8080 mattipaksula/usvad-focus \
&& docker logs -f usvad-focus
```
