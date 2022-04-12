# usvad-fiesta

```
docker run -p 80:8080 mattipaksula/usvad-fiesta
```

## oneliner to update and leave running

```shell
docker pull mattipaksula/usvad-fiesta \
&& docker rm -f usvad-fiesta \
&& docker run -d --name usvad-fiesta --rm -p 80:8080 mattipaksula/usvad-fiesta \
&& docker logs -f usvad-fiesta
```
